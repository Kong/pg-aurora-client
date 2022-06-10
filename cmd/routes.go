package main

import (
	"github.com/gorilla/mux"
	"net/http"
)

func (ac *appContext) routes() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/pghealth", ac.getPGHealthHandler).Methods("GET")
	r.HandleFunc("/health", ac.getHealthHandler).Methods("GET")
	return r
}

func (ac *appContext) getHealthHandler(w http.ResponseWriter, r *http.Request) {
	err := ac.writeJSON(w, http.StatusOK, envelope{"status": "ok"}, nil)
	if err != nil {
		ac.logError(r, err)
	}
}

func (ac *appContext) getPGHealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UFT-8")
	// your logic here to call
	status, err := ac.Store.GetReplicaStatus()
	if err != nil {
		ac.logError(r, err)
		ac.errorResponse(w, r, http.StatusInternalServerError, "Failed to Query PG")
		return
	}
	err = ac.writeJSON(w, http.StatusOK, envelope{"replicaStatusList": status}, nil)
	if err != nil {
		ac.logError(r, err)
	}
}
