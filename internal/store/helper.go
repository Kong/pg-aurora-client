package store

import (
	"context"
	"fmt"
	_ "github.com/golang-migrate/migrate/v4/database/pgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kong/pg-aurora-client/pkg/pool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	defaultMaxConnections = 50
	defaultMinConnections = 20
)

var writeQuery = `UPDATE canary SET id=id +1, ts = CURRENT_TIMESTAMP`

func writeValidator(ctx context.Context, conn *pgxpool.Conn, logger *zap.Logger) bool {
	exec, err := conn.Exec(ctx, writeQuery)
	if err != nil {
		logger.Error("write validation failed", zap.Error(err))
		return false
	}

	logger.Info("healthcheck write canary", zap.Int64("rowsUpdated", exec.RowsAffected()))
	return true
}

func SetupTestDatabase() (testcontainers.Container, *pool.AuroraPGPool, error) {
	containerReq := testcontainers.ContainerRequest{
		Image:        "postgres:latest",
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForListeningPort("5432/tcp"),
		Env: map[string]string{
			"POSTGRES_DB":       "koko",
			"POSTGRES_PASSWORD": "koko",
			"POSTGRES_USER":     "koko",
		},
	}
	dbContainer, err := testcontainers.GenericContainer(
		context.Background(),
		testcontainers.GenericContainerRequest{
			ContainerRequest: containerReq,
			Started:          true,
		})
	if err != nil {
		return nil, nil, err
	}
	port, err := dbContainer.MappedPort(context.Background(), "5432")
	if err != nil {
		return nil, nil, err
	}
	host, err := dbContainer.Host(context.Background())
	if err != nil {
		return nil, nil, err
	}

	dbURI := fmt.Sprintf("postgres://koko:koko@%v:%v/koko?sslmode=disable", host, port.Port())
	err = MigrateDb(dbURI)
	if err != nil {
		return nil, nil, err
	}

	// change this to AuroraPool
	config, err := pgxpool.ParseConfig(dbURI)
	if err != nil {
		return nil, nil, err
	}

	config.MaxConns = defaultMaxConnections
	config.MinConns = defaultMinConnections
	apConfig := &pool.Config{
		PGXConfig:      config,
		QueryValidator: writeValidator,
	}

	logger, err := SetupLogging("info")
	if err != nil {
		return nil, nil, err
	}

	connPool, err := pool.NewAuroraPool(context.Background(), apConfig, logger)
	if err != nil {
		return nil, nil, err
	}

	//connPool, err := pgxpool.New(context.Background(), dbURI)
	//if err != nil {
	//	return nil, nil, err
	//}

	return dbContainer, connPool, err
}

var Logger *zap.Logger
var zapConfig zap.Config

// SetupLogging configure parent logger with logLevel.
func SetupLogging(logLevel string) (*zap.Logger, error) {
	zapConfig := zap.NewDevelopmentConfig()
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	level, err := zapcore.ParseLevel(logLevel)
	if err != nil {
		return nil, err
	}
	zapConfig.Level.SetLevel(level)
	logger, err := zapConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("create logger: %w", err)
	}
	Logger = logger
	return logger, nil
}
