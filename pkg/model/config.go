package model

import (
	"fmt"
	"os"
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

func loadPostgresConfig() (*PgConfig, error) {
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
