package postgres

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"

	"github.com/stretchr/testify/assert"

	"go-service-template/internal/app/infrastructure"
)

func TestIntegrationPostgresqlHandlerTX_getTx(t *testing.T) {
	if testing.Short() {
		t.Skip("skip integration tests in short mode")
	}
	baseContext := context.Background()
	ctxWithTransaction := context.Background()
	err := target.NewTx(&ctxWithTransaction)
	if err != nil {
		t.Error("can't init transaction")
	}

	// value - any object different from tx
	ctxBad := context.WithValue(baseContext, infrastructure.CtxKeyTransaction{}, errors.New("test"))
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "PostgresqlHandlerTX. getTx. Case #1. Positive",
			args: args{
				ctx: ctxWithTransaction,
			},
			wantErr: false,
		},
		{
			name: "PostgresqlHandlerTX. getTx. Case #2. Empty ",
			args: args{
				ctx: baseContext,
			},
			wantErr: true,
		},
		{
			name: "PostgresqlHandlerTX. getTx. Case #3. Bad transaction in context",
			args: args{
				ctx: ctxBad,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := target.getTx(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("getTx() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestIntegrationPostgresqlHandlerTX_Transaction(t *testing.T) {
	if testing.Short() {
		t.Skip("skip integration tests in short mode")
	}
	type args struct {
		query  string
		commit bool
		rowCnt int
		resCnt int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "PostgresqlHandlerTX. Commit. Case #1.",
			args: args{
				query:  "insert into test_table (a,b) values ($1, $2)",
				commit: true,
				rowCnt: 1,
				resCnt: 1,
			},
			wantErr: false,
		},
		{
			name: "PostgresqlHandlerTX. Rollback. Case #2.",
			args: args{
				query:  "insert into test_table (a,b) values ($1, $2)",
				commit: false,
				rowCnt: 2,
				resCnt: 0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			txFunc := target.WithTx(context.Background())
			err := txFunc(ctx, func(ctx context.Context) error {
				var err error
				fmt.Println("insert data undo transaction")
				for i := 0; i < tt.args.rowCnt; i++ {
					err = target.Execute(ctx, tt.args.query, 10, tt.name)
				}

				return err
			})
			var pe *pgconn.PgError
			if err != nil && errors.As(err, &pe) {
				if pe.Code != pgerrcode.UniqueViolation {
					t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				}
			}

			fmt.Println("check result")
			rows, err := target.Query(context.Background(), "select * from test_table where b=$1", tt.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Query() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			var cnt int
			for rows.Next() {
				cnt += 1
			}
			assert.Equal(t, tt.args.resCnt, cnt, "expected %d rows, got %d", tt.args.resCnt, cnt)
		})
	}
}

func TestIntegrationPostgresqlHandlerTX_ExecuteBatch(t *testing.T) {
	if testing.Short() {
		t.Skip("skip integration tests in short mode")
	}

	const batchStatement = "insert into test_table (a, b) values ($1, $2)"
	type args struct {
		statement string
		size      int
		a         []int
		b         []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "ExecuteBatch test#1",
			args: args{
				statement: batchStatement,
				size:      0,
				a:         nil,
				b:         nil,
			},
			wantErr: false,
		},
		{
			name: "ExecuteBatch test#2",
			args: args{
				statement: batchStatement,
				size:      1,
				a:         []int{50},
				b:         []string{"str"},
			},
			wantErr: false,
		},
		{
			name: "ExecuteBatch test#3",
			args: args{
				statement: batchStatement,
				size:      3,
				a:         []int{100, 200, 300},
				b:         []string{"str1", "str2", "str3"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var paramArr [][]interface{}
			if tt.args.size > 0 {
				for i := 0; i < tt.args.size; i++ {
					var paramLine []interface{}
					paramLine = append(paramLine, tt.args.a[i])
					paramLine = append(paramLine, tt.args.b[i])
					paramArr = append(paramArr, paramLine)
				}
			}
			if err := target.ExecuteBatch(context.Background(), tt.args.statement, paramArr); (err != nil) != tt.wantErr {
				t.Errorf("ExecuteBatch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIntegrationPostgresqlHandlerTX_ExecuteTransactionBatch(t *testing.T) {
	if testing.Short() {
		t.Skip("skip integration tests in short mode")
	}

	const batchStatement = "insert into test_table (a, b) values ($1, $2)"
	type args struct {
		statement string
		size      int
		a         []int
		b         []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "ExecuteBatch test#1",
			args: args{
				statement: batchStatement,
				size:      0,
				a:         nil,
				b:         nil,
			},
			wantErr: false,
		},
		{
			name: "ExecuteBatch test#2",
			args: args{
				statement: batchStatement,
				size:      1,
				a:         []int{70},
				b:         []string{"str"},
			},
			wantErr: false,
		},
		{
			name: "ExecuteBatch test#3",
			args: args{
				statement: batchStatement,
				size:      3,
				a:         []int{1000, 2000, 3000},
				b:         []string{"str1", "str2", "str3"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var paramArr [][]interface{}
			if tt.args.size > 0 {
				for i := 0; i < tt.args.size; i++ {
					var paramLine []interface{}
					paramLine = append(paramLine, tt.args.a[i])
					paramLine = append(paramLine, tt.args.b[i])
					paramArr = append(paramArr, paramLine)
				}
			}
			ctxTx := context.Background()
			assert.NoError(t, target.NewTx(&ctxTx))
			if err := target.ExecuteBatch(ctxTx, tt.args.statement, paramArr); (err != nil) != tt.wantErr {
				t.Errorf("ExecuteBatch() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.NoError(t, target.Commit(ctxTx))
		})
	}
}

func TestIntegrationPostgresqlHandlerTX_QueryRow(t *testing.T) {
	if testing.Short() {
		t.Skip("skip integration tests in short mode")
	}

	baseContext := context.Background()

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "PostgresqlHandlerTX. QueryRow. Case #1. Positive",
			args: args{ctx: baseContext},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var res int
			row, err := target.QueryRow(tt.args.ctx, "select 10")
			assert.NoError(t, err)
			err = row.Scan(&res)
			assert.NoError(t, err)
			assert.Equal(t, 10, res)
		})
	}
}
