package pool

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func setupPGEnv(t *testing.T) {
	t.Setenv("PG_USER", "koko")
	t.Setenv("PG_PASSWORD", "koko")
	t.Setenv("PG_DATABASE", "postgres")
	t.Setenv("PG_HOST", "localhost")
	t.Setenv("PG_PORT", "5435")
}

func setupLogging() (*zap.Logger, error) {
	zapConfig := zap.NewDevelopmentConfig()
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	level, err := zapcore.ParseLevel("info")
	if err != nil {
		return nil, err
	}
	zapConfig.Level.SetLevel(level)
	logger, err := zapConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("create logger: %w", err)
	}
	return logger, nil
}

func TestAuroraPGPool_ValidateWrite(t *testing.T) {
	setupPGEnv(t)
	logger, err := setupLogging()
	require.NoError(t, err)
	pgc, err := loadPostgresConfig()
	require.NoError(t, err)
	dsn := getDSN(pgc)
	config, err := pgxpool.ParseConfig(dsn)
	require.NoError(t, err)
	ctx := context.Background()
	apConfig := &Config{
		PGXConfig:      config,
		QueryValidator: DefaultWriteValidator,
	}

	testPool, err := NewAuroraPool(ctx, apConfig, logger)
	exec, err := testPool.Exec(ctx, writeQuery)
	require.NoError(t, err)
	require.Equal(t, exec.RowsAffected(), int64(1))
}

func TestAuroraPGPool_ValidateRead(t *testing.T) {
	setupPGEnv(t)
	logger, err := setupLogging()
	require.NoError(t, err)
	pgc, err := loadPostgresConfig()
	require.NoError(t, err)
	dsn := getRODSN(pgc)
	config, err := pgxpool.ParseConfig(dsn)
	require.NoError(t, err)
	ctx := context.Background()
	apConfig := &Config{
		PGXConfig:      config,
		QueryValidator: DefaultWriteValidator,
	}

	testPool, err := NewAuroraPool(ctx, apConfig, logger)
	rows, err := testPool.Query(context.Background(), readerQuery)
	require.NoError(t, err)
	rows.Close()
}

func getDSN(pgc *pgConfig) string {
	var dsn string
	if !pgc.enableTLS {
		dsn = fmt.Sprintf(dsnNoTLS, pgc.user, pgc.password, pgc.hostURL, pgc.port, pgc.database)
	} else {
		dsn = fmt.Sprintf(dsnTLS, pgc.user, pgc.password, pgc.hostURL, pgc.port, pgc.database, pgc.caBundleFSPath)
	}
	return dsn
}

func getRODSN(pgc *pgConfig) string {
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

func loadPostgresConfig() (*pgConfig, error) {
	isSecure := os.Getenv("ENABLE_TLS")
	tls := false
	if isSecure == "yes" || isSecure == "true" {
		tls = true
	}

	pgc := &pgConfig{
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

type pgConfig struct {
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

func validate(pgc *pgConfig) error {
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
