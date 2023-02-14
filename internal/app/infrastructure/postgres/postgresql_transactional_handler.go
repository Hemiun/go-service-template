package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog"

	"github.com/jackc/pgx/v4/log/zerologadapter"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"go-service-template/internal/app/infrastructure"
	"go-service-template/internal/app/repository/basedbhandler"
)

// InFunc - func type for passing into transaction
type InFunc func(context.Context) error

// TxFunc - func type for
type TxFunc func(ctx context.Context, f InFunc) error

// PostgresqlHandlerTX struct for interactions with RDBMS Postgres.
type PostgresqlHandlerTX struct {
	infrastructure.SugarLogger
	pool         *pgxpool.Pool
	dataSource   string
	pgPoolConfig PgPoolConfig
}

// PgPoolConfig - struct for pgpool params
type PgPoolConfig struct {
	// MaxConnLifetime is the duration since creation after which a connection will be automatically closed.
	MaxConnLifetime time.Duration `yaml:"maxConnLifetime"`

	// MaxConnIdleTime is the duration after which an idle connection will be automatically closed by the health check.
	MaxConnIdleTime time.Duration `yaml:"maxConnIdleTime"`

	//MaxConns  is the maximum size of the pool.
	MaxConns int32 `yaml:"maxConns"`

	// MinConns is the minimum size of the pool.
	MinConns int32 `yaml:"minConns"`

	// HealthCheckPeriod is the duration between checks of the health of idle connections.
	HealthCheckPeriod time.Duration `yaml:"healthCheckPeriod"`

	// LazyConnect. If set to true, pool doesn't do any I/O operation on initialization. And connects to the server only when the pool starts to be used.
	LazyConnect bool `yaml:"lazyConnect"`

	// LogPGX - enable logging inside pgx
	LogPGX bool `env:"LOG_PGX" yaml:"logPGX"`

	// LogLevel - set log level for pgx. Available values are trace, debug, info, warn, error, none
	LogLevel string `env:"PGX_LOG_LEVEL" yaml:"pgxLogLevel"`
}

// DatabaseConfig - struct for db params
type DatabaseConfig struct {
	// Address - database host name
	Address string `env:"DB_ADDRESS" yaml:"address" validate:"required"`

	// Port - database port
	Port int32 `env:"DB_PORT" yaml:"port" validate:"required"`

	// DB - database name
	DB string `env:"DB_NAME" yaml:"db" validate:"required"`

	// Username  - name of user, that used for db connection
	Username string `env:"DB_USERNAME" yaml:"username" validate:"required"`

	// Password - password for username
	Password string `env:"DB_PASSWORD" yaml:"password" validate:"required"`
}

// NewPostgresqlHandlerTX return new PostgresqlHandlerTX
func NewPostgresqlHandlerTX(ctx context.Context, dataSource string, pgPoolConfig PgPoolConfig) (*PostgresqlHandlerTX, error) {
	// Format DSN
	//("postgresql://%s:%s@%s:%s/%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Dbname)
	postgresqlHandler := new(PostgresqlHandlerTX)
	postgresqlHandler.dataSource = dataSource
	postgresqlHandler.pgPoolConfig = pgPoolConfig
	if err := postgresqlHandler.Init(ctx); err != nil {
		return nil, err
	}
	return postgresqlHandler, nil
}

// Init -  func for initialisation
func (handler *PostgresqlHandlerTX) Init(ctx context.Context) error {
	log := infrastructure.GetBaseLogger(ctx)
	internalLog := zerologadapter.NewLogger(*log, zerologadapter.WithContextFunc(func(ctx context.Context, logWith zerolog.Context) zerolog.Context {
		// You can use zerolog.hlog.IDFromCtx(ctx) or even
		// zerolog.log.Ctx(ctx) to fetch the whole logger instance from the
		// context if you want.
		requestID, ok := ctx.Value(infrastructure.CtxKeyRequestID{}).(string)
		if ok {
			logWith = logWith.Str(infrastructure.RequestIDField, requestID)
		}
		correlationID, ok := ctx.Value(infrastructure.CtxKeyCorrelationID{}).(string)
		if ok {
			logWith = logWith.Str(infrastructure.CorrelationIDField, correlationID)
		}
		return logWith
	}))
	poolConfig, err := pgxpool.ParseConfig(handler.dataSource)
	if err != nil {
		log.Error().Err(err).Msg("can't parse datasource")
		return err
	}

	if handler.pgPoolConfig.MaxConns > 0 {
		poolConfig.MaxConns = handler.pgPoolConfig.MaxConns
	}
	if handler.pgPoolConfig.MinConns > 0 {
		poolConfig.MinConns = handler.pgPoolConfig.MinConns
	}

	if handler.pgPoolConfig.MaxConnIdleTime > 0 {
		poolConfig.MaxConnIdleTime = handler.pgPoolConfig.MaxConnIdleTime
	}

	if handler.pgPoolConfig.MaxConnLifetime > 0 {
		poolConfig.MaxConnLifetime = handler.pgPoolConfig.MaxConnLifetime
	}
	if handler.pgPoolConfig.HealthCheckPeriod > 0 {
		poolConfig.HealthCheckPeriod = handler.pgPoolConfig.HealthCheckPeriod
	}

	poolConfig.LazyConnect = handler.pgPoolConfig.LazyConnect
	if handler.pgPoolConfig.LogPGX {
		poolConfig.ConnConfig.Logger = internalLog
		switch handler.pgPoolConfig.LogLevel {
		case "trace":
			poolConfig.ConnConfig.LogLevel = pgx.LogLevelTrace
		case "debug":
			poolConfig.ConnConfig.LogLevel = pgx.LogLevelDebug
		case "info":
			poolConfig.ConnConfig.LogLevel = pgx.LogLevelInfo
		case "warn":
			poolConfig.ConnConfig.LogLevel = pgx.LogLevelWarn
		case "error":
			poolConfig.ConnConfig.LogLevel = pgx.LogLevelError
		case "none":
			poolConfig.ConnConfig.LogLevel = pgx.LogLevelNone
		default:
			poolConfig.ConnConfig.LogLevel = pgx.LogLevelInfo
		}
	}
	pool, err := pgxpool.ConnectConfig(ctx, poolConfig)
	if err != nil {
		log.Error().Err(err).Msg("can't start connection pool")
		return err
	}
	handler.pool = pool
	return nil
}

// Commit - call commit for transaction from context
func (handler *PostgresqlHandlerTX) Commit(ctx context.Context) error {
	tx, err := handler.getTx(ctx)
	if err != nil {
		// Ошибку не логируем, т.к. это часть сценария работы без транзакции
		return err
	}
	err = tx.Commit(ctx)
	if err != nil {
		handler.LogError(ctx, "Can't commit transaction", err)
		return err
	}
	return err
}

// Rollback - - call rollback for transaction from context
func (handler *PostgresqlHandlerTX) Rollback(ctx context.Context) error {
	tx, err := handler.getTx(ctx)
	if err != nil {
		// Ошибку не логируем, т.к. это часть сценария работы без транзакции
		return err
	}
	err = tx.Rollback(ctx)
	if err != nil {
		handler.LogError(ctx, "Can't rollback transaction", err)
		return err
	}
	return err
}

// NewTx - create a new transaction and put it into context
func (handler *PostgresqlHandlerTX) NewTx(ctx *context.Context) error {
	// 1. Checks is a transaction present in context. if it is, no new one is created
	if _, err := handler.getTx(*ctx); err == nil {
		// Ошибку не логируем т.к. это ожидаемое поведение
		return nil
	}

	// 2. New transaction creation
	newTx, err := handler.pool.Begin(*ctx)
	if err != nil {
		handler.LogError(*ctx, "PostgresqlHandlerTX: can't create tx", err)
		return err
	}

	// 3. New context with transaction
	newCtx := context.WithValue(*ctx, infrastructure.CtxKeyTransaction{}, newTx)
	*ctx = newCtx

	return nil
}

// getTx - get transaction from context
func (handler *PostgresqlHandlerTX) getTx(ctx context.Context) (pgx.Tx, error) {
	var err error
	ctxValue := ctx.Value(infrastructure.CtxKeyTransaction{})
	if ctxValue == nil {
		handler.LogDebug(ctx, "PostgresqlHandlerTX: can't get tx")
		return nil, ErrTxNotFound
	}
	tx, ok := ctxValue.(pgx.Tx)
	if !ok {
		err = ErrTxTypeConversation
		handler.LogError(ctx, "PostgresqlHandlerTX: can't get tx", err)
		return nil, err
	}
	return tx, err
}

// WithTx - transaction method factory
func (handler *PostgresqlHandlerTX) WithTx(_ context.Context) TxFunc {
	return func(ctx context.Context, f InFunc) (err error) {
		// 1. Start new transaction
		if err = handler.NewTx(&ctx); err != nil {
			handler.LogPanic(ctx, "can't start transaction", err)
			return err
		}

		// 2. Process panic. If it is, call rollback
		defer func() {
			if r := recover(); r != nil {
				handler.LogPanic(ctx, "Panic. Try to rollback", err)
				if err = handler.Rollback(ctx); err != nil {
					handler.LogPanic(ctx, "Can't rollback transaction", err)
				}
			}
		}()

		// 2. Execute logic under transaction
		err = f(ctx)
		if err != nil {
			if err = handler.Rollback(ctx); err != nil {
				handler.LogPanic(ctx, "Can't rollback transaction", err)
			}
		} else {
			if err = handler.Commit(ctx); err != nil {
				handler.LogPanic(ctx, "Can't commit transaction", err)
			}
		}
		return err
	}
}

// Execute - method for statement execution
func (handler *PostgresqlHandlerTX) Execute(ctx context.Context, statement string, args ...interface{}) error {
	tx, err := handler.getTx(ctx)
	statement = handler.clearStatement(statement)

	if err == nil { //nolint:nestif
		if len(args) > 0 {
			_, err = tx.Exec(ctx, statement, args...)
		} else {
			_, err = tx.Exec(ctx, statement)
		}
	} else {
		conn, e := handler.pool.Acquire(ctx)
		if e != nil {
			handler.LogError(ctx, "Can't acquire connection from pool", err)
			return e
		}
		defer conn.Release()

		if len(args) > 0 {
			_, e = conn.Exec(ctx, statement, args...)
		} else {
			_, e = conn.Exec(ctx, statement)
		}
		err = e
	}
	if err != nil {
		handler.LogError(ctx, "Can't execute statement", err)
		return err
	}
	return nil
}

// ExecuteBatch method for batch statement execution
func (handler *PostgresqlHandlerTX) ExecuteBatch(ctx context.Context, statement string, args [][]interface{}) error {
	var (
		err error
		ct  pgconn.CommandTag
		br  pgx.BatchResults
	)

	statement = handler.clearStatement(statement)
	batch := &pgx.Batch{}
	if len(args) > 0 {
		for _, argset := range args {
			batch.Queue(statement, argset...)
		}
	} else {
		return nil
	}
	tx, err := handler.getTx(ctx)

	if err == nil {
		br = tx.SendBatch(ctx, batch)
	} else {
		conn, err2 := handler.pool.Acquire(ctx)
		if err2 != nil {
			handler.LogError(ctx, "Can't acquire connection from pool", err)
			return err2
		}
		defer conn.Release()
		br = conn.SendBatch(ctx, batch)
	}
	ct, err = br.Exec()
	defer br.Close() //nolint:errcheck
	if err != nil {
		handler.LogError(ctx, "Can't execute batch statement", err)
		return err
	}
	fmt.Println(ct.RowsAffected())
	return nil
}

// QueryRow -  method for  one row SELECT statement
func (handler *PostgresqlHandlerTX) QueryRow(ctx context.Context, statement string, args ...interface{}) (basedbhandler.Row, error) {
	var row pgx.Row

	statement = handler.clearStatement(statement)
	tx, err := handler.getTx(ctx)
	if err == nil { //nolint:nestif
		if len(args) > 0 {
			row = tx.QueryRow(ctx, statement, args...)
		} else {
			row = tx.QueryRow(ctx, statement)
		}
	} else {
		conn, err := handler.pool.Acquire(ctx)
		if err != nil {
			handler.LogError(ctx, "Can't acquire connection from pool", err)
			return nil, err
		}
		defer conn.Release()
		if len(args) > 0 {
			row = conn.QueryRow(ctx, statement, args...)
		} else {
			row = conn.QueryRow(ctx, statement)
		}
	}
	return row, nil
}

// Query -  method for arbitrary SELECT statement
func (handler *PostgresqlHandlerTX) Query(ctx context.Context, statement string, args ...interface{}) (basedbhandler.Rows, error) {
	var rows pgx.Rows

	statement = handler.clearStatement(statement)
	tx, err := handler.getTx(ctx)

	if err == nil { //nolint:nestif
		if len(args) > 0 {
			rows, err = tx.Query(ctx, statement, args...)
		} else {
			rows, err = tx.Query(ctx, statement)
		}
	} else {
		conn, e := handler.pool.Acquire(ctx)
		if e != nil {
			handler.LogError(ctx, "Can't acquire connection from pool", err)
			return nil, e
		}
		defer conn.Release()
		if len(args) > 0 {
			rows, e = conn.Query(ctx, statement, args...)
		} else {
			rows, e = conn.Query(ctx, statement)
		}
		err = e
	}
	if err != nil {
		handler.LogError(ctx, "Can't execute query", err)
		return nil, err
	}
	return rows, nil
}

// GetNextID - get next value from sequence
func (handler *PostgresqlHandlerTX) GetNextID(ctx context.Context, sequenceName string) (int64, error) {
	var row pgx.Row
	var id int64

	statement := fmt.Sprintf("SELECT nextval('%s')", sequenceName)

	row, err := handler.QueryRow(ctx, statement)
	if err != nil {
		handler.LogError(ctx, "can't get next value from sequence", err)
		return 0, err
	}

	err = row.Scan(&id)
	if err != nil {
		handler.LogError(ctx, "can't scan value from response", err)
		return 0, err
	}

	return id, nil
}

// Close - close connection pool
func (handler *PostgresqlHandlerTX) Close(_ context.Context) error {
	if handler != nil {
		handler.pool.Close()
	}
	return nil
}

func (handler *PostgresqlHandlerTX) clearStatement(query string) string {
	buf := strings.ReplaceAll(query, "\n", " ")
	buf = strings.ReplaceAll(buf, "\t", " ")
	return buf
}
