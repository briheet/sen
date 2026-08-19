package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	address := flag.String("address", ":8080", "HTTP listen address")
	flag.Parse()

	log.Printf("HTTP server listening on %s", *address)
	if err := http.ListenAndServe(*address, routes()); err != nil {
		log.Fatal(err)
	}
}

func routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", hello)
	mux.HandleFunc("/health", health)
	mux.HandleFunc("/work", work)
	return logRequests(mux)
}
