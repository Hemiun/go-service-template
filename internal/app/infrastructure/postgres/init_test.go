package postgres

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	"go-service-template/internal/app/testcontainer"

	"github.com/jackc/pgx/v4"
	"go-service-template/internal/app/infrastructure"
)

var (
	target *PostgresqlHandlerTX
	dsn    string
	log    *infrastructure.Logger
)

func initDatabase(ctx context.Context) {
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		log.Panic().Err(err).Msg("Can't init connection")
	}
	_, err = conn.Exec(ctx, "create table if not exists test_table (a numeric primary key , b varchar)")
	if err != nil {
		log.Panic().Err(err).Msg("Can't create test table")
	}
	_, err = conn.Exec(ctx, "delete from test_table")
	if err != nil {
		log.Panic().Err(err).Msg("Can't clear test table")
	}
}

func TestMain(m *testing.M) {
	ctx := context.Background()
	infrastructure.InitGlobalLogger("debug", "go-service-template", "")
	log = infrastructure.GetBaseLogger(ctx)
	flag.Parse()
	if !testing.Short() {
		var err error
		cfg := testcontainer.DatabaseContainerConfig{
			DatabaseName:      "territory",
			OwnerSchema:       "test",
			OwnerSchemaPass:   "test",
			ServiceSchema:     "test_ms",
			ServiceSchemaPass: "test_ms",
			Timeout:           time.Minute,
		}

		db, err := testcontainer.NewDatabaseContainer(ctx, cfg)
		if err != nil {
			log.Panic().Err(err).Msg("can't init db container")
		}
		err = db.PrepareDB(ctx)
		if err != nil {
			log.Panic().Err(err).Msg("can't prepare db")
		}
		defer db.Close(ctx)
		dsn = fmt.Sprintf("postgres://%s:%s@127.0.0.1:%d/%s?sslmode=disable", cfg.OwnerSchema, cfg.OwnerSchemaPass, db.Port(ctx), cfg.DatabaseName)

		initDatabase(ctx)
		target, err = NewPostgresqlHandlerTX(ctx, dsn, PgPoolConfig{})
		if err != nil {
			log.Panic().Err(err).Msg("can't init PostgresqlHandlerTX")
		}
	}

	os.Exit(m.Run())
}
