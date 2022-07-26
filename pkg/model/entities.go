package model

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
	"time"
)

const defaultMaxConnections = 50
const defaultMinConnections = 20

type ReplicaStatus struct {
	ServerID    string    `json:"serverID"`
	SessionID   string    `json:"sessionID"`
	LastUpdated time.Time `json:"lastUpdated"`
}

type Canary struct {
	ID          int64     `json:"ID"`
	LastUpdated time.Time `json:"lastUpdated"`
	DiffMS      float64   `json:"diffMS"`
}

type Foo struct {
	ID          string    `json:"id"`
	Created     time.Time `json:"created"`
	LastUpdated time.Time `json:"lastUpdated"`
}

var replicaStatusQuery = `SELECT SERVER_ID, SESSION_ID, LAST_UPDATE_TIMESTAMP FROM aurora_replica_status()
     WHERE EXTRACT(EPOCH FROM(NOW() - LAST_UPDATE_TIMESTAMP)) <= 300 OR SESSION_ID = 'MASTER_SESSION_ID'
     ORDER BY LAST_UPDATE_TIMESTAMP DESC`

var getLastFooQuery = `SELECT ID, CREATED_AT, UPDATED_AT FROM FOO ORDER BY CREATED_AT DESC LIMIT 1`

var insertFoo = `INSERT INTO foo (id) VALUES (default)`

var updateHealthQuery = `UPDATE canary SET id=id +1, ts = CURRENT_TIMESTAMP RETURNING id,ts`

var roHealthQuery = `SELECT id, ts, Extract(epoch FROM (current_timestamp - ts))*1000 AS diff_ms from canary;`

type Store struct {
	dbPool   *pgxpool.Pool
	roDBPool *pgxpool.Pool
	Logger   *zap.Logger
}

func NewStore(logger *zap.Logger) (*Store, error) {
	pgc, err := loadPostgresConfig()
	if err != nil {
		return nil, err
	}

	var dsn string
	var rodsn string
	if !pgc.enableTLS {
		dsn = fmt.Sprintf(dsnNoTLS, pgc.user, pgc.password, pgc.hostURL, pgc.port, pgc.database)
		if pgc.roHostURL != "" {
			rodsn = fmt.Sprintf(dsnNoTLS, pgc.user, pgc.password, pgc.roHostURL, pgc.port, pgc.database)
		}
	} else {
		dsn = fmt.Sprintf(dsnTLS, pgc.user, pgc.password, pgc.hostURL, pgc.port, pgc.database, pgc.caBundleFSPath)
		if pgc.roHostURL != "" {
			rodsn = fmt.Sprintf(dsnTLS, pgc.user, pgc.password, pgc.roHostURL, pgc.port, pgc.database,
				pgc.caBundleFSPath)
		}
	}
	pool, err := openPool(dsn, pgc, logger)
	if err != nil {
		return nil, err
	}
	logger.Info("Established DB Connection")
	store := &Store{
		dbPool: pool,
		Logger: logger,
	}

	if rodsn != "" {
		rodbPool, err := openPool(rodsn, pgc, logger)
		if err != nil {
			return nil, err
		}

		store.roDBPool = rodbPool
		logger.Info("Established RO DB Connection")
	}
	return store, nil
}

type PoolStats struct {
	AcquireCount    int64
	AcquireDuration time.Duration
	AcquiredConns   int32
	IdleConns       int32
	TotalConns      int32
	MaxConns        int32
}

func (s *Store) GetReplicaStatus(ro bool) ([]ReplicaStatus, error) {
	rsList := []ReplicaStatus{}
	var rows pgx.Rows
	var err error
	ctx := context.Background()
	if ro && s.roDBPool != nil {
		rows, err = s.roDBPool.Query(ctx, replicaStatusQuery)
	} else {
		if ro {
			s.Logger.Warn("using rw connection because there ro connection is not injected")
		}
		rows, err = s.dbPool.Query(ctx, replicaStatusQuery)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var rs ReplicaStatus
		err := rows.Scan(
			&rs.ServerID,
			&rs.SessionID,
			&rs.LastUpdated)
		if err != nil {
			return nil, err
		}
		rsList = append(rsList, rs)
	}
	return rsList, nil
}

func (s *Store) GetMostRecentFoo() (*Foo, error) {
	var foo Foo
	var rows pgx.Rows
	var err error
	ctx := context.Background()
	if s.roDBPool != nil {
		rows, err = s.roDBPool.Query(ctx, getLastFooQuery)
	} else {
		rows, err = s.dbPool.Query(ctx, getLastFooQuery)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if rows.Next() {
		err := rows.Scan(
			&foo.ID,
			&foo.Created,
			&foo.LastUpdated)
		if err != nil {
			return nil, err
		}
	}
	return &foo, nil
}

func (s *Store) InsertFoo() (int64, error) {
	ctx := context.Background()
	exec, err := s.dbPool.Exec(ctx, insertFoo)
	var affected int64
	if err != nil {
		return affected, err
	}
	affected = exec.RowsAffected()
	return affected, nil
}

func (s *Store) GetConnectionPoolStats() *PoolStats {
	stat := s.dbPool.Stat()
	poolstats := &PoolStats{
		AcquireCount:    stat.AcquireCount(),
		AcquireDuration: stat.AcquireDuration(),
		AcquiredConns:   stat.AcquiredConns(),
		IdleConns:       stat.IdleConns(),
		TotalConns:      stat.TotalConns(),
		MaxConns:        stat.MaxConns(),
	}
	return poolstats
}

func (s *Store) GetROConnectionPoolStats() *PoolStats {
	stat := s.roDBPool.Stat()
	poolstats := &PoolStats{
		AcquireCount:    stat.AcquireCount(),
		AcquireDuration: stat.AcquireDuration(),
		AcquiredConns:   stat.AcquiredConns(),
		IdleConns:       stat.IdleConns(),
		TotalConns:      stat.TotalConns(),
		MaxConns:        stat.MaxConns(),
	}
	return poolstats
}

func (s *Store) UpdatePoolHealthCheck() (*Canary, error) {
	var canary Canary
	var rows pgx.Rows
	var err error
	ctx := context.Background()
	rows, err = s.dbPool.Query(ctx, updateHealthQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if rows.Next() {
		err := rows.Scan(
			&canary.ID,
			&canary.LastUpdated)
		if err != nil {
			return nil, err
		}
	}
	return &canary, nil
}

func (s *Store) GetPoolHealthCheck() (*Canary, error) {
	var canary Canary
	var rows pgx.Rows
	var err error
	ctx := context.Background()
	if s.roDBPool != nil {
		rows, err = s.roDBPool.Query(ctx, roHealthQuery)
	} else {
		rows, err = s.dbPool.Query(ctx, roHealthQuery)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if rows.Next() {
		err := rows.Scan(
			&canary.ID,
			&canary.LastUpdated,
			&canary.DiffMS)
		if err != nil {
			return nil, err
		}
	}
	return &canary, nil
}

func openPool(dsn string, pgc *PgConfig, logger *zap.Logger) (*pgxpool.Pool, error) {
	logger.Info("DB connection:", zap.String("host", pgc.hostURL),
		zap.Bool("Enable TLS", pgc.enableTLS),
		zap.String("user", pgc.user), zap.String("port", pgc.port),
		zap.String("database", pgc.database), zap.String("caBundlePath", pgc.caBundleFSPath))
	ctx := context.Background()
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	config.MaxConns = defaultMaxConnections
	config.MinConns = defaultMinConnections

	dbpool, err := pgxpool.ConnectConfig(ctx, config)
	if err != nil {
		return nil, err
	}
	err = dbpool.Ping(ctx)
	if err != nil {
		return nil, err
	}
	return dbpool, nil
}

func (s *Store) Close() {
	if s.dbPool != nil {
		s.dbPool.Close()
	}

	if s.roDBPool != nil {
		s.roDBPool.Close()
	}

}
