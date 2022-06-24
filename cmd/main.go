package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/kong/pg-aurora-client/pkg/model"
	"go.uber.org/zap"
	"log"
	"net/http"
)

type appContext struct {
	Store  *model.Store
	Logger *zap.Logger
}

func main() {
	pgc, err := loadPostgresConfig()
	if err != nil {
		log.Fatal(err)
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
	logger, err := SetupLogging("info")
	if err != nil {
		log.Fatal(err)
	}
	pool, err := openPool(dsn, pgc, logger)
	if err != nil {
		logger.Error("DB Connection failed", zap.Error(err))
	}
	defer pool.Close()
	logger.Info("Established DB Connection")

	ac := &appContext{
		Store:  &model.Store{DBPool: pool, Logger: logger},
		Logger: logger,
	}
	if rodsn != "" {
		rodbPool, err := openPool(rodsn, pgc, logger)
		if err != nil {
			logger.Error("DB RO Connection failed", zap.Error(err))
		}
		defer rodbPool.Close()
		ac.Store.RODBPool = rodbPool
		logger.Info("Established RO DB Connection")
	}
	ac.Logger.Info("Application is running on : 8080 .....")
	http.ListenAndServe("0.0.0.0:8080", ac.routes())
}

//var tenantID = "001c5e3c-6086-4c66-b0de-8b5eba9fa655"

func openPool(dsn string, pgc *pgConfig, logger *zap.Logger) (*pgxpool.Pool, error) {
	logger.Info("DB connection:", zap.String("host", pgc.hostURL),
		zap.Bool("Enable TLS", pgc.enableTLS),
		zap.String("user", pgc.user), zap.String("port", pgc.port),
		zap.String("database", pgc.database), zap.String("caBundlePath", pgc.caBundleFSPath))
	ctx := context.Background()
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	//config.BeforeAcquire = func(ctx context.Context, c *pgx.Conn) bool {
	//	logger.Info("Before acquire..")
	//	_, err := c.Exec(ctx, fmt.Sprintf("SELECT set_tenant('%s')", tenantID))
	//	if err != nil {
	//		logger.Error("Setting tenant id failed", zap.Error(err))
	//		return false
	//	}
	//	return true
	//}
	//config.AfterRelease = func(conn *pgx.Conn) bool {
	//	logger.Info("After release..")
	//	_, err := conn.Exec(ctx, "SELECT unset_tenant()")
	//	if err != nil {
	//		logger.Error("Unsetting tenant id failed", zap.Error(err))
	//		return false
	//	}
	//	return true
	//}

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
