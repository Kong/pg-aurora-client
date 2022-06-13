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

func main() {
	pgc, err := loadPostgresConfig()
	if err != nil {
		log.Fatal(err)
	}
	var dsn string
	if !pgc.enableTLS {
		dsn = fmt.Sprintf(dsnNoTLS, pgc.user, pgc.password, pgc.hostURL, pgc.port, pgc.database)
	} else {
		dsn = fmt.Sprintf(dsnTLS, pgc.user, pgc.password, pgc.hostURL, pgc.port, pgc.database, pgc.caBundleFSPath)
	}
	logger, err := SetupLogging("debug")
	if err != nil {
		log.Fatal(err)
	}
	db, err := openDB(dsn, pgc, logger)
	if err != nil {
		logger.Info("DB Connection failed", zap.Error(err))
	}
	defer db.Close()
	ac := &appContext{
		Store:  &model.Store{DB: db},
		Logger: logger,
	}
	ac.Logger.Info("Application is running on : 8080 .....")
	http.ListenAndServe("0.0.0.0:8080", ac.routes())
}

func openDB(dsn string, pgc *pgConfig, logger *zap.Logger) (*sql.DB, error) {
	logger.Info("DB connection:", zap.String("host", pgc.hostURL),
		zap.Bool("Enable TLS", pgc.enableTLS),
		zap.String("user", pgc.user), zap.String("port", pgc.port),
		zap.String("caBundlePath", pgc.caBundleFSPath))
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
