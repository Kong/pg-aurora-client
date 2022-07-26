package pool

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"github.com/kong/pg-aurora-client/pkg/model"
	"time"
)

type ValidateWriteFunc func(ctx context.Context, store *model.Store) (interface{}, error)

type ValidateReadFunc func(ctx context.Context, store *model.Store) (interface{}, error)

type auroraValidator struct {
	store             *model.Store
	validateWriteFunc ValidateWriteFunc
	validateReadFunc  ValidateReadFunc
}

type ConnectionValidator interface {
	ValidateWrite(ctx context.Context) (interface{}, error)
	ValidateRead(ctx context.Context) (interface{}, error)
}

func (v *auroraValidator) ValidateWrite(ctx context.Context) (interface{}, error) {
	return v.validateWriteFunc(ctx, v.store)
}

func (v *auroraValidator) ValidateRead(ctx context.Context) (interface{}, error) {
	return v.validateReadFunc(ctx, v.store)
}

func NewDefaultConnValidator(store *model.Store, writeFunc ValidateWriteFunc, readFunc ValidateReadFunc) ConnectionValidator {
	return &auroraValidator{store: store, validateWriteFunc: writeFunc, validateReadFunc: readFunc}
}

// ConnPool db conns pool interface
type ConnPool interface {
	Begin() (*sql.Tx, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)

	Close() error

	Conn(ctx context.Context) (*sql.Conn, error)
	Driver() driver.Driver

	Exec(query string, args ...any) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)

	Ping() error
	PingContext(ctx context.Context) error

	Prepare(query string) (*sql.Stmt, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)

	Query(query string, args ...any) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row

	SetConnMaxIdleTime(d time.Duration)
	SetConnMaxLifetime(d time.Duration)
	SetMaxIdleConns(n int)
	SetMaxOpenConns(n int)

	Stats() sql.DBStats
}
