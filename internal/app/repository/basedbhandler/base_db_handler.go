package basedbhandler

import (
	"context"
)

// DBHandler - interface for interaction with db
//
//go:generate mockgen -destination=mocks/mock_postgres_handler.go -package=mocks . DBHandler
type DBHandler interface {
	Execute(ctx context.Context, statement string, args ...interface{}) error
	ExecuteBatch(ctx context.Context, statement string, args [][]interface{}) error
	Query(ctx context.Context, statement string, args ...interface{}) (Rows, error)
	QueryRow(ctx context.Context, statement string, args ...interface{}) (Row, error)
	GetNextID(ctx context.Context, statement string) (int64, error)
	Close(ctx context.Context) error
}

// TransactionalDBHandler - interface for interaction with db in transaction mode
type TransactionalDBHandler interface {
	Execute(ctx context.Context, statement string, args ...interface{}) error
	ExecuteBatch(ctx context.Context, statement string, args [][]interface{}) error
	Query(ctx context.Context, statement string, args ...interface{}) (Rows, error)
	QueryRow(ctx context.Context, statement string, args ...interface{}) (Row, error)
	GetNextID(ctx context.Context, statement string) (int64, error)
	Transactioner
}

// Transactioner interface for working with transaction
type Transactioner interface {
	// Commit -  fix transaction
	Commit(ctx context.Context) error

	// Rollback transaction
	Rollback(ctx context.Context) error

	// NewTx - starts new transaction
	NewTx(ctx *context.Context) error
}

// Rows - interface for working with rows
type Rows interface {
	Scan(dest ...interface{}) error
	Next() bool
}

// Row  - interface for working with singe row
//
//go:generate mockgen -destination=mocks/mock_row.go -package=mocks . Row
type Row interface {
	Scan(dest ...interface{}) error
}
