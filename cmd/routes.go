package main

import (
	"github.com/gorilla/mux"
	"net/http"
)

func (ac *appContext) routes() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/pghealth", ac.getPGHealthHandler).Methods("GET")
	r.HandleFunc("/foo", ac.getPGFooHandler).Methods("GET")
	r.HandleFunc("/health", ac.getHealthHandler).Methods("GET")
	return r
}

func (ac *appContext) getHealthHandler(w http.ResponseWriter, _ *http.Request) {
	err := ac.writeJSON(w, http.StatusOK, envelope{"status": "ok"}, nil)
	if err != nil {
		ac.logError(err)
	}
}

func (ac *appContext) getPGHealthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UFT-8")
	// your logic here to call
	status, err := ac.Store.GetReplicaStatus()
	if err != nil {
		ac.logError(err)
		ac.errorResponse(w, http.StatusInternalServerError, "Failed to Query PG")
		return
	}
	payload := envelope{"replicaStatusList": status}
	err = ac.writeJSON(w, http.StatusOK, payload, nil)
	if err != nil {
		ac.logError(err)
	}
	ac.logJson(payload)
}

func (ac *appContext) getPGFooHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UFT-8")
	// your logic here to call
	foo, err := ac.Store.GetMostRecentFoo()
	if err != nil {
		ac.logError(err)
		ac.errorResponse(w, http.StatusInternalServerError, "Failed to Query PG")
		return
	}
	payload := envelope{"foo": foo}
	err = ac.writeJSON(w, http.StatusOK, payload, nil)
	if err != nil {
		ac.logError(err)
	}
	ac.logJson(payload)
}
