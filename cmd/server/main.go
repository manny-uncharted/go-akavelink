package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	akavesdk "github.com/akave-ai/go-akavelink/internal/sdk" // aliased for clarity
	"github.com/joho/godotenv"                               // Import godotenv
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
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: No .env file found or could not be loaded: %v. Relying on system environment variables.", err)
		// It's a warning, not a fatal error, because variables might be set directly in the environment.
	}

	// 0. Read the private key from the environment variable.
	privateKey := os.Getenv("AKAVE_PRIVATE_KEY")
	if privateKey == "" {
		log.Fatal("FATAL: AKAVE_PRIVATE_KEY environment variable not set.")
	}
	nodeAddress := os.Getenv("AKAVE_NODE_ADDRESS")
	if nodeAddress == "" {
		log.Fatal("FATAL: AKAVE_NODE_ADDRESS environment variable not set.")
	}
	// 1. Configure and initialize our Akave client wrapper.
	cfg := akavesdk.Config{
		NodeAddress:       nodeAddress,
		MaxConcurrency:    10,
		BlockPartSize:     1024 * 1024, // 1MB
		UseConnectionPool: true,
		PrivateKeyHex:     privateKey,
	}

	client, err := akavesdk.NewClient(cfg)
	if err != nil {
		log.Fatalf("Fatal error initializing Akave client: %v", err)
	}
	// Ensure the connection is closed when the application exits.
	defer client.Close()

	// 2. Create a new server instance with the initialized client.
	srv := &server{
		client: client,
	}

	// 3. Register the handlers. The handlers are now methods on our server struct,
	// which gives them access to the client.
	http.HandleFunc("/health", srv.healthHandler)

	log.Println("Starting go-akavelink server on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// healthHandler is now a method on the server.
func (s *server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
