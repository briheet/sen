package main

import (
	"encoding/json"
	"net/http"
)

func (a *application) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", a.home)
	mux.HandleFunc("/health", a.health)
	mux.HandleFunc("/visit", a.visit)
	mux.HandleFunc("/report", a.report)
	return logRequests(mux)
}

func (a *application) home(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"endpoints": []string{"/health", "/visit?user=alice", "/report"},
	})
}

func (a *application) health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (a *application) visit(w http.ResponseWriter, r *http.Request) {
	user := r.URL.Query().Get("user")
	if user == "" {
		user = "guest"
	}

	result, err := a.service.RecordVisit(r.Context(), user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (a *application) report(w http.ResponseWriter, r *http.Request) {
	result, err := a.service.Report(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
