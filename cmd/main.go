package main

import (
	"database/sql"
	"fmt"
	"github.com/kong/pg-aurora-client/pkg/model"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"log"
	"net/http"
)

type appContext struct {
	Store  *model.Store
	Logger *zap.Logger
}

const defaultMaxConn = 50
const defaultMaxIdleConns = 20

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
	db, err := openDB(dsn, pgc, logger)
	if err != nil {
		logger.Error("DB Connection failed", zap.Error(err))
	}
	defer db.Close()
	logger.Info("Established DB Connection")

	db.SetMaxOpenConns(defaultMaxConn)
	db.SetMaxIdleConns(defaultMaxIdleConns)

	ac := &appContext{
		Store:  &model.Store{DB: db, Logger: logger},
		Logger: logger,
	}
	if rodsn != "" {
		rodb, err := openDB(rodsn, pgc, logger)
		if err != nil {
			logger.Error("DB RO Connection failed", zap.Error(err))
		}
		defer rodb.Close()
		rodb.SetMaxOpenConns(defaultMaxConn)
		rodb.SetMaxIdleConns(defaultMaxIdleConns)
		ac.Store.RODB = rodb
		logger.Info("Established RO DB Connection")
	}
	ac.Logger.Info("Application is running on : 8080 .....")
	http.ListenAndServe("0.0.0.0:8080", ac.routes())
}

func openDB(dsn string, pgc *pgConfig, logger *zap.Logger) (*sql.DB, error) {
	logger.Info("DB connection:", zap.String("host", pgc.hostURL),
		zap.Bool("Enable TLS", pgc.enableTLS),
		zap.String("user", pgc.user), zap.String("port", pgc.port),
		zap.String("database", pgc.database), zap.String("caBundlePath", pgc.caBundleFSPath))
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
