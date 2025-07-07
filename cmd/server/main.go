package main

import (
	"fmt"
	"log"
	"net/http"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "ok")
}


// server holds the application's dependencies, like our Akave client.
// This makes it easy to pass dependencies to our handlers.
type server struct {
	client *akavesdk.Client
}

func main() {
	http.HandleFunc("/health", healthHandler)

	log.Println("Starting go-akavelink server on :8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
