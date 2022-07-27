package model

import (
	"database/sql"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
	"time"
)

type ReplicaStatus struct {
	ServerID    string    `json:"serverID"`
	SessionID   string    `json:"sessionID"`
	LastUpdated time.Time `json:"lastUpdated"`
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

var updateHealthQuery = `UPDATE canary SET id=id +1, ts = CURRENT_TIMESTAMP`

type Store struct {
	DB     *sql.DB
	RODB   *sql.DB
	Logger *zap.Logger
}

func (s *Store) GetReplicaStatus(ro bool) ([]ReplicaStatus, error) {
	rsList := []ReplicaStatus{}
	var rows *sql.Rows
	var err error
	if ro && s.RODB != nil {
		rows, err = s.RODB.Query(replicaStatusQuery)
	} else {
		if ro {
			s.Logger.Warn("using rw connection because there ro connection is not injected")
		}
		rows, err = s.DB.Query(replicaStatusQuery)
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
	var rows *sql.Rows
	var err error
	if s.RODB != nil {
		rows, err = s.RODB.Query(getLastFooQuery)
	} else {
		rows, err = s.DB.Query(getLastFooQuery)
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
	exec, err := s.DB.Exec(insertFoo)
	var affected int64
	if err != nil {
		return affected, err
	}
	affected, err = exec.RowsAffected()
	if err != nil {
		return affected, err
	}
	return affected, nil
}

func (s *Store) UpdateCanary() (int64, error) {
	exec, err := s.DB.Exec(updateHealthQuery)
	var affected int64
	if err != nil {
		return affected, err
	}
	affected, err = exec.RowsAffected()
	if err != nil {
		return affected, err
	}
	return affected, nil
}

func (s *Store) GetConnectionPoolStats() sql.DBStats {
	stat := s.DB.Stats()
	return stat
}

func (s *Store) GetROConnectionPoolStats() sql.DBStats {
	stat := s.RODB.Stats()
	return stat
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
		DB:     pool,
		Logger: logger,
	}

	if rodsn != "" {
		rodbPool, err := openPool(rodsn, pgc, logger)
		if err != nil {
			return nil, err
		}

		store.RODB = rodbPool
		logger.Info("Established RO DB Connection")
	}
	return store, nil
}

const defaultMaxConnections = 50
const defaultMinConnections = 20

func openPool(dsn string, pgc *PgConfig, logger *zap.Logger) (*sql.DB, error) {
	logger.Info("DB connection:", zap.String("host", pgc.hostURL),
		zap.Bool("Enable TLS", pgc.enableTLS),
		zap.String("user", pgc.user), zap.String("port", pgc.port),
		zap.String("database", pgc.database), zap.String("caBundlePath", pgc.caBundleFSPath))
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	config.MaxConns = defaultMaxConnections
	config.MinConns = defaultMinConnections

	dbpool, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	err = dbpool.Ping()
	if err != nil {
		return nil, err
	}
	return dbpool, nil
}
