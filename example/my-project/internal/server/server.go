package server

import (
	"fmt"
	"log"
	"net/http"
)

// Start initializes and starts the HTTP server
func Start() {
	fmt.Println("Starting HTTP server on :8080")
	
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/health", handleHealth)
	
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to myapp server!")
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status": "healthy"}`)
} 