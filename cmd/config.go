package main

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

// Logger is setup on startup by cmd package.
var Logger *zap.Logger
var zapConfig zap.Config

// SetupLogging configure parent logger with logLevel.
func SetupLogging(logLevel string) (*zap.Logger, error) {
	zapConfig = zap.NewProductionConfig()
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

// SetLevel updates the level for the global logger config.
// All child loggers generated with the config are updated.
func SetLevel(level string) error {
	parsedLevel, err := zapcore.ParseLevel(level)
	if err != nil {
		return fmt.Errorf("set log level: %w", err)
	}
	zapConfig.Level.SetLevel(parsedLevel)
	return nil
}

func loadPostgresConfig() (*pgConfig, error) {
	isSecure := os.Getenv("ENABLE_TLS")
	var tls = false
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

func validate(pgc *pgConfig) error {
	if pgc.user == "" {
		return fmt.Errorf("PG_USER cannot be empty")
	}
	if pgc.password == "" {
		return fmt.Errorf("PG_PASSWORD cannot be empty")
	}
	if pgc.hostURL == "" {
		return fmt.Errorf("PG_HOST cannot be empty")
	}
	if pgc.port == "" {
		return fmt.Errorf("PG_PORT cannot be empty")
	}
	if pgc.database == "" {
		return fmt.Errorf("PG_DATABASE cannot be empty")
	}
	if pgc.enableTLS && pgc.caBundleFSPath == "" {
		return fmt.Errorf("ENABLE_TLS requires a valid PG_CA_BUNDLE_FS_PATH")
	}
	return nil

}
