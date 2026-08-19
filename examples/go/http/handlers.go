package main

import (
	"crypto/sha256"
	"fmt"
	"net/http"
)

func hello(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		name = "world"
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = fmt.Fprintln(w, greeting(name))
}

func greeting(name string) string {
	return "Hello, " + name + "!"
}

func health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func work(w http.ResponseWriter, r *http.Request) {
	digest := workload(r.URL.Query().Get("name"))
	_, _ = fmt.Fprintf(w, "%x\n", digest[:4])
}

// workload keeps the request stack observable during the recording.
//
//go:noinline
func workload(input string) [32]byte {
	digest := sha256.Sum256([]byte(input))
	for range 10_000 {
		digest = sha256.Sum256(digest[:])
	}
	return digest
}
