package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: go run backend.go <port> <message>")
		os.Exit(1)
	}

	port := os.Args[1]
	message := os.Args[2]
	listenAddr := ":" + port

	handler := func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Backend [%s] received request: %s %s from %s", port, r.Method, r.URL.Path, r.RemoteAddr)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Message from backend on port %s: %s\nPath: %s\n", port, message, r.URL.Path)
	}

	http.HandleFunc("/", handler)

	log.Printf("Starting backend server on port %s with message: '%s'", port, message)
	err := http.ListenAndServe(listenAddr, nil)
	if err != nil {
		log.Fatalf("Backend server on port %s failed: %v", port, err)
	}
}
