package main

import (
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
	logger, err := SetupLogging("info")
	if err != nil {
		log.Fatal(err)
	}
	store, err := model.NewStore(logger)
	if err != nil {
		logger.Fatal("Failed setting up store", zap.Error(err))
	}

	ac := &appContext{Store: store, Logger: logger}
	defer store.Close()

	ac.Logger.Info("Application is running on : 8080 .....")
	http.ListenAndServe("0.0.0.0:8080", ac.routes())
}
