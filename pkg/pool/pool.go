package pool

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
	"time"
)

type ConnectionValidator interface {
	ValidateWrite(ctx context.Context) error
	ValidateRead(ctx context.Context) error
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

type PGXConnPool interface {
	Close()
	Acquire(ctx context.Context) (*pgxpool.Conn, error)
	AcquireFunc(ctx context.Context, f func(*pgxpool.Conn) error) error
	AcquireAllIdle(ctx context.Context) []*pgxpool.Conn
	Config() *pgxpool.Config
	Stat() *pgxpool.Stat
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	QueryFunc(ctx context.Context, sql string, args []interface{}, scans []interface{}, f func(pgx.QueryFuncRow) error) (pgconn.CommandTag, error)
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	Begin(ctx context.Context) (pgx.Tx, error)
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
	BeginFunc(ctx context.Context, f func(pgx.Tx) error) error
	BeginTxFunc(ctx context.Context, txOptions pgx.TxOptions, f func(pgx.Tx) error) error
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
	Ping(ctx context.Context) error
}

type ValidationFunction func(ctx context.Context, conn *pgxpool.Conn, logger *zap.Logger) bool

type AuroraPGPool struct {
	innerPool         *pgxpool.Pool
	writeValidateFunc ValidationFunction
	readValidateFunc  ValidationFunction
	logger            *zap.Logger
}

func (p *AuroraPGPool) Close() {
	p.innerPool.Close()
}

func (p *AuroraPGPool) Acquire(ctx context.Context) (*pgxpool.Conn, error) {
	return p.innerPool.Acquire(ctx)
}

func (p *AuroraPGPool) AcquireFunc(ctx context.Context, f func(*pgxpool.Conn) error) error {
	return p.innerPool.AcquireFunc(ctx, f)
}

func (p *AuroraPGPool) AcquireAllIdle(ctx context.Context) []*pgxpool.Conn {
	return p.innerPool.AcquireAllIdle(ctx)
}

func (p *AuroraPGPool) Config() *pgxpool.Config {
	return p.innerPool.Config()
}

func (p *AuroraPGPool) Stat() *pgxpool.Stat {
	return p.innerPool.Stat()
}

func (p *AuroraPGPool) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	return p.innerPool.Exec(ctx, sql, arguments...)
}

func (p *AuroraPGPool) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return p.innerPool.Query(ctx, sql, args...)
}

func (p *AuroraPGPool) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return p.innerPool.QueryRow(ctx, sql, args...)
}

func (p *AuroraPGPool) QueryFunc(ctx context.Context, sql string, args []interface{}, scans []interface{},
	f func(pgx.QueryFuncRow) error) (pgconn.CommandTag, error) {
	return p.innerPool.QueryFunc(ctx, sql, args, scans, f)
}

func (p *AuroraPGPool) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	return p.innerPool.SendBatch(ctx, b)
}

func (p *AuroraPGPool) Begin(ctx context.Context) (pgx.Tx, error) {
	return p.innerPool.Begin(ctx)
}

func (p *AuroraPGPool) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	return p.innerPool.BeginTx(ctx, txOptions)
}

func (p *AuroraPGPool) BeginFunc(ctx context.Context, f func(pgx.Tx) error) error {
	return p.innerPool.BeginFunc(ctx, f)
}

func (p *AuroraPGPool) BeginTxFunc(ctx context.Context, txOptions pgx.TxOptions, f func(pgx.Tx) error) error {
	return p.innerPool.BeginTxFunc(ctx, txOptions, f)
}

func (p *AuroraPGPool) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string,
	rowSrc pgx.CopyFromSource) (int64, error) {
	return p.innerPool.CopyFrom(ctx, tableName, columnNames, rowSrc)
}

func (p *AuroraPGPool) Ping(ctx context.Context) error {
	return p.innerPool.Ping(ctx)
}

func (p *AuroraPGPool) ValidateWrite(ctx context.Context) error {
	if p.writeValidateFunc == nil {
		return errors.New("no WriteValidateFunc set")
	}
	conn, err := p.innerPool.Acquire(ctx)
	defer conn.Release()
	if err != nil {
		return err
	}
	validated := p.writeValidateFunc(ctx, conn, p.logger)
	if !validated {
		return errors.New("write validation failed")
	}
	return nil
}

func (p *AuroraPGPool) ValidateRead(ctx context.Context) error {
	if p.readValidateFunc == nil {
		return errors.New("no ReadValidateFunc set")
	}
	conn, err := p.innerPool.Acquire(ctx)
	defer conn.Release()
	if err != nil {
		return err
	}
	validated := p.readValidateFunc(ctx, conn, p.logger)
	if !validated {
		return errors.New("read validation failed")
	}
	return nil
}

func NewAuroraPool(ctx context.Context, config *Config, logger *zap.Logger) (*AuroraPGPool, error) {
	dbpool, err := pgxpool.ConnectConfig(ctx, config.PGXConfig)
	if err != nil {
		return nil, err
	}
	err = dbpool.Ping(ctx)
	if err != nil {
		return nil, err
	}
	return &AuroraPGPool{
		innerPool:         dbpool,
		logger:            logger,
		writeValidateFunc: config.WriteValidator,
		readValidateFunc:  config.ReadValidator,
	}, nil

	// Start the validator
}
