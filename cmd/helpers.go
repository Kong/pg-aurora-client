package main

import (
	"encoding/json"
	"net/http"
	"runtime/debug"
)

type envelope map[string]interface{}

func (ac *appContext) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	js = append(js, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)
	return nil
}

func (ac *appContext) errorResponse(w http.ResponseWriter, status int, message interface{}) {
	env := envelope{"error": message}
	err := ac.writeJSON(w, status, env, nil)
	if err != nil {
		ac.logError(err)
		w.WriteHeader(500)
	}
}

func (ac *appContext) logError(err error) {
	ac.Logger.Sugar().Errorf("%s\n%s", err.Error(), debug.Stack())
}

func (ac *appContext) logJson(message interface{}) {
	env := envelope{"payload": message}
	js, err := json.Marshal(env)
	if err != nil {
		ac.logError(err)
		return
	}
	ac.Logger.Info(string(js))
}
