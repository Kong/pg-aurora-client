package model

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/kong/pg-aurora-client/pkg/pool"
	"go.uber.org/zap"
	"os"
	"time"
)

type PgConfig struct {
	user           string
	database       string
	password       string
	hostURL        string
	roHostURL      string
	port           string
	enableTLS      bool
	caBundleFSPath string
}

var dsnNoTLS = "postgres://%s:%s@%s:%s/%s?sslmode=disable"

var dsnTLS = "postgres://%s:%s@%s:%s/%s?sslmode=verify-ca&sslrootcert=%s"

const caBundleFSPath = "/config/ca_certs/aws-postgres-cabundle-secret"

func validate(pgc *PgConfig) error {
	if pgc.user == "" {
		return fmt.Errorf("env variable PG_USER cannot be empty")
	}
	if pgc.password == "" {
		return fmt.Errorf("env variable PG_PASSWORD cannot be empty")
	}
	if pgc.hostURL == "" {
		return fmt.Errorf("env variable PG_HOST cannot be empty")
	}
	if pgc.port == "" {
		return fmt.Errorf("env variable PG_PORT cannot be empty")
	}
	if pgc.database == "" {
		return fmt.Errorf("env variable PG_DATABASE cannot be empty")
	}
	return nil
}

func LoadPostgresConfig() (*PgConfig, error) {
	isSecure := os.Getenv("ENABLE_TLS")
	var tls = false
	if isSecure == "yes" || isSecure == "true" {
		tls = true
	}

	pgc := &PgConfig{
		user:           os.Getenv("PG_USER"),
		password:       os.Getenv("PG_PASSWORD"),
		hostURL:        os.Getenv("PG_HOST"),
		roHostURL:      os.Getenv("PG_RO_HOST"),
		port:           os.Getenv("PG_PORT"),
		database:       os.Getenv("PG_DATABASE"),
		enableTLS:      tls,
		caBundleFSPath: caBundleFSPath,
	}

	if err := validate(pgc); err != nil {
		return nil, err
	}
	return pgc, nil
}

func getDSN(pgc *PgConfig) string {
	var dsn string
	if !pgc.enableTLS {
		dsn = fmt.Sprintf(dsnNoTLS, pgc.user, pgc.password, pgc.hostURL, pgc.port, pgc.database)
	} else {
		dsn = fmt.Sprintf(dsnTLS, pgc.user, pgc.password, pgc.hostURL, pgc.port, pgc.database, pgc.caBundleFSPath)
	}
	return dsn
}

func getRODSN(pgc *PgConfig) string {
	var dsn string
	if !pgc.enableTLS {
		if pgc.roHostURL == "" {
			dsn = fmt.Sprintf(dsnNoTLS, pgc.user, pgc.password, pgc.hostURL, pgc.port, pgc.database)
		} else {
			dsn = fmt.Sprintf(dsnNoTLS, pgc.user, pgc.password, pgc.roHostURL, pgc.port, pgc.database)
		}
	} else {
		if pgc.roHostURL == "" {
			dsn = fmt.Sprintf(dsnTLS, pgc.user, pgc.password, pgc.hostURL, pgc.port, pgc.database, pgc.caBundleFSPath)
		} else {
			dsn = fmt.Sprintf(dsnTLS, pgc.user, pgc.password, pgc.roHostURL, pgc.port, pgc.database, pgc.caBundleFSPath)
		}
	}
	return dsn
}

func openPool(dsn string, pgc *PgConfig, logger *zap.Logger, validator pool.ValidationFunction) (pool.PGXConnPool, error) {
	logger.Debug("DB connection:", zap.String("host", pgc.hostURL),
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
	// Intentionally not being aggressive since we have 2 background check threads
	config.HealthCheckPeriod = time.Minute * 5
	apConfig := &pool.Config{
		PGXConfig:      config,
		QueryValidator: validator,
	}

	dbpool, err := pool.NewAuroraPool(ctx, apConfig, logger)
	if err != nil {
		return nil, err
	}
	return dbpool, nil
}
