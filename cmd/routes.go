package main

import (
	"github.com/gorilla/mux"
	"net/http"
)

func (ac *appContext) routes() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/health", ac.getHealth).Methods("GET")
	// Aurora specific
	r.HandleFunc("/replstatus", ac.getReplicationStatus).Methods("GET")
	r.HandleFunc("/ro/replstatus", ac.getROReplicationStatus).Methods("GET")

	// Generic health
	r.HandleFunc("/poolstats", ac.getConnectionPoolStats).Methods("GET")
	r.HandleFunc("/ro/poolstats", ac.getROConnectionPoolStats).Methods("GET")
	r.HandleFunc("/canary", ac.getCanary).Methods("GET")
	r.HandleFunc("/canary", ac.upsertCanary).Methods("POST")
	r.HandleFunc("/replicationcanary", ac.getReplicationCanary).Methods("GET")
	r.HandleFunc("/replicationcanary", ac.upsertReplicationCanary).Methods("POST")

	// KAdmin
	r.HandleFunc("/foo", ac.getPGFoo).Methods("GET")
	r.HandleFunc("/foo", ac.postPGFoo).Methods("POST")

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
	payload := envelope{"roreplicaStatusList": status}
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

func (ac *appContext) getConnectionPoolStats(w http.ResponseWriter, _ *http.Request) {
	stats := ac.Store.GetConnectionPoolStats(false)
	payload := envelope{"connectionPoolStats": stats}
	err := ac.writeJSON(w, http.StatusOK, payload, nil)
	if err != nil {
		ac.logError(err)
	}
	ac.logJson(payload)
}

func (ac *appContext) getROConnectionPoolStats(w http.ResponseWriter, _ *http.Request) {
	stats := ac.Store.GetConnectionPoolStats(true)
	payload := envelope{"roconnectionPoolStats": stats}
	err := ac.writeJSON(w, http.StatusOK, payload, nil)
	if err != nil {
		ac.logError(err)
	}
	ac.logJson(payload)
}

func (ac *appContext) getCanary(w http.ResponseWriter, _ *http.Request) {
	canary, err := ac.Store.GetCanary()
	if err != nil {
		ac.logError(err)
		ac.errorResponse(w, http.StatusInternalServerError, "Failed to Query PG")
		return
	}
	payload := envelope{"canary": canary}
	err = ac.writeJSON(w, http.StatusOK, payload, nil)
	if err != nil {
		ac.logError(err)
	}
	ac.logJson(payload)
}

func (ac *appContext) upsertCanary(w http.ResponseWriter, _ *http.Request) {
	canary, err := ac.Store.UpdateCanary()
	if err != nil {
		ac.logError(err)
		ac.errorResponse(w, http.StatusInternalServerError, "Failed to Query PG")
		return
	}
	payload := envelope{"canary": canary}
	err = ac.writeJSON(w, http.StatusOK, payload, nil)
	if err != nil {
		ac.logError(err)
	}
	ac.logJson(payload)
}

func (ac *appContext) getReplicationCanary(w http.ResponseWriter, _ *http.Request) {
	canary, err := ac.Store.GetReplicationCanary()
	if err != nil {
		ac.logError(err)
		ac.errorResponse(w, http.StatusInternalServerError, "Failed to Query PG")
		return
	}
	payload := envelope{"replicationCanary": canary}
	err = ac.writeJSON(w, http.StatusOK, payload, nil)
	if err != nil {
		ac.logError(err)
	}
	ac.logJson(payload)
}

func (ac *appContext) upsertReplicationCanary(w http.ResponseWriter, _ *http.Request) {
	canary, err := ac.Store.UpdateReplicationCanary()
	if err != nil {
		ac.logError(err)
		ac.errorResponse(w, http.StatusInternalServerError, "Failed to Query PG")
		return
	}
	payload := envelope{"replicationCanary": canary}
	err = ac.writeJSON(w, http.StatusOK, payload, nil)
	if err != nil {
		ac.logError(err)
	}
	ac.logJson(payload)
}
