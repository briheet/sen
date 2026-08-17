package main

import (
	"fmt"
	"net/http"
)

const address = "127.0.0.1:18080"

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = writer.Write([]byte("ok"))
	})
	mux.HandleFunc("/work", work)
	_ = http.ListenAndServe(address, mux)
}

func work(writer http.ResponseWriter, _ *http.Request) {
	data := make([]byte, 128<<10)
	for index := range data {
		data[index] = byte(index)
	}
	_, _ = fmt.Fprintf(writer, "%d\n", fibonacci(24)+int(data[len(data)-1]))
}

func fibonacci(value int) int {
	if value < 2 {
		return value
	}
	return fibonacci(value-1) + fibonacci(value-2)
}
