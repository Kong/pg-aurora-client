package main

import (
	"github.com/kong/pg-aurora-client/pkg/metrics"
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
	pgc, err := model.LoadPostgresConfig()
	if err != nil {
		log.Fatal(err)
	}
	logger, err := SetupLogging("info")
	if err != nil {
		log.Fatal(err)
	}
	s, err := model.NewStore(logger, pgc)
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()

	ac := &appContext{
		Store:  s,
		Logger: logger,
	}
	// Initialize the metrics
	err = metrics.InitMetricsClient(logger, "datadog")
	if err != nil {
		logger.Error("Failed to initialize metrics", zap.Error(err))
	}
	ac.Logger.Info("Application is running on : 8080 .....")
	http.ListenAndServe("0.0.0.0:8080", ac.routes())
}
