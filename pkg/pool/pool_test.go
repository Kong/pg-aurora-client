package pool

import (
	"context"
	"fmt"
	"github.com/kong/pg-aurora-client/pkg/model"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"testing"
)

func testWriterConnValidator(ctx context.Context, store *model.Store) (interface{}, error) {
	canary, err := store.UpdatePoolHealthCheck()
	if err != nil {
		return nil, err
	}
	return canary, err
}

func testReaderConnValidator(ctx context.Context, store *model.Store) (interface{}, error) {
	canary, err := store.GetPoolHealthCheck()
	if err != nil {
		return nil, err
	}
	return canary, nil
}

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

func TestAuroraValidator_ValidateWrite(t *testing.T) {
	setupPGEnv(t)
	logger, err := setupLogging()
	require.NoError(t, err)
	s, err := model.NewStore(logger)
	require.NoError(t, err)
	connnValidator := NewDefaultConnValidator(s, testWriterConnValidator, testReaderConnValidator)
	write, err := connnValidator.ValidateWrite(context.Background())
	require.NoError(t, err)
	canary, ok := write.(*model.Canary)
	require.True(t, ok)
	logger.Info("Canary", zap.Int64("id", canary.ID), zap.Time("ts", canary.LastUpdated))
	read, err := connnValidator.ValidateRead(context.Background())
	require.NoError(t, err)
	readCanary, ok := read.(*model.Canary)
	require.True(t, ok)
	logger.Info("ReadCanary", zap.Int64("id", readCanary.ID), zap.Time("ts", readCanary.LastUpdated),
		zap.Float64("duration", readCanary.DiffMS))
}
