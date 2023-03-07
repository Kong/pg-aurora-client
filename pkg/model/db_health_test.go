package model

import (
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"testing"
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

func TestAuroraValidator_ValidateWrite(t *testing.T) {
	setupPGEnv(t)
	logger, err := setupLogging()
	require.NoError(t, err)
	pgc, err := LoadPostgresConfig()
	require.NoError(t, err)
	s, err := NewStore(logger, pgc)
	require.NoError(t, err)
	_, err = s.UpdateCanary()
	if err != nil {
		return
	}
}
