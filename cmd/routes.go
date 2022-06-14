package main

import (
	"github.com/gorilla/mux"
	"net/http"
)

func (ac *appContext) routes() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/pghealth", ac.getReplicationStatus).Methods("GET")
	r.HandleFunc("/replstatusro", ac.getROReplicationStatus).Methods("GET")
	r.HandleFunc("/foo", ac.getPGFoo).Methods("GET")
	r.HandleFunc("/foo", ac.postPGFoo).Methods("POST")
	r.HandleFunc("/health", ac.getHealth).Methods("GET")
	return r
}

func (ac *appContext) getHealth(w http.ResponseWriter, _ *http.Request) {
	err := ac.writeJSON(w, http.StatusOK, envelope{"status": "ok"}, nil)
	if err != nil {
		ac.logError(err)
	}
}

func (ac *appContext) getReplicationStatus(w http.ResponseWriter, _ *http.Request) {
	status, err := ac.Store.GetReplicaStatus(false)
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

func (ac *appContext) getROReplicationStatus(w http.ResponseWriter, _ *http.Request) {
	status, err := ac.Store.GetReplicaStatus(true)
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

func (ac *appContext) getPGFoo(w http.ResponseWriter, _ *http.Request) {
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

func (ac *appContext) postPGFoo(w http.ResponseWriter, _ *http.Request) {
	foo, err := ac.Store.InsertFoo()
	if err != nil {
		ac.logError(err)
		ac.errorResponse(w, http.StatusInternalServerError, "Failed to Query PG")
		return
	}
	payload := envelope{"rowsInserted": foo}
	err = ac.writeJSON(w, http.StatusOK, payload, nil)
	if err != nil {
		ac.logError(err)
	}
	ac.logJson(payload)
}
