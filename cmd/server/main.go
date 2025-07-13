// Package cmd provides the HTTP server that manages AkaveLink buckets and files.
//
// It exposes RESTful endpoints for health checks, bucket management, and file operations.
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	akavesdk "github.com/akave-ai/go-akavelink/internal/sdk"
	"github.com/akave-ai/go-akavelink/internal/utils"
)


// server encapsulates dependencies for HTTP handlers.
type server struct {
	client *akavesdk.Client
}

// healthHandler responds with a simple status OK message.
func (s *server) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}


// main initializes the server and routes.
func main() {
	utils.LoadEnvConfig()

	key := os.Getenv("AKAVE_PRIVATE_KEY")
	node := os.Getenv("AKAVE_NODE_ADDRESS")
	if key == "" || node == "" {
		log.Fatal("AKAVE_PRIVATE_KEY and AKAVE_NODE_ADDRESS are required")
	}

	cfg := akavesdk.Config{
		NodeAddress:       node,
		MaxConcurrency:    10,
		BlockPartSize:     1 << 20,
		UseConnectionPool: true,
		PrivateKeyHex:     key,
	}
	client, err := akavesdk.NewClient(cfg)
	if err != nil {
		log.Fatalf("client initialization failed: %v", err)
	}
	defer client.Close()

	r := mux.NewRouter()
	srv := &server{client: client}

	r.HandleFunc("/health", srv.healthHandler).Methods("GET")

	log.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
