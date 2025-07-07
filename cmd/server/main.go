// cmd/server/main.go
// cmd/server/main.go
package main // Keep this as 'main' for the executable entry point // Keep this as 'main' for the executable entry point

import (
	"log"
	"net/http"
	"os"

	akavesdk "github.com/akave-ai/go-akavelink/internal/sdk"
	"github.com/akave-ai/go-akavelink/internal/utils" // Import your new utils package
)


// server holds the application's dependencies, like our Akave client.
type server struct {
	client *akavesdk.Client
}

// MainFunc contains the core logic of your application,
// making it testable by allowing external calls.
func MainFunc() {
	// Load environment variables using the reusable function
	utils.LoadEnvConfig()

	// 0. Read the private key from the environment variable.
	privateKey := os.Getenv("AKAVE_PRIVATE_KEY")
	if privateKey == "" {
		log.Fatal("FATAL: AKAVE_PRIVATE_KEY environment variable not set. This is required.")
	}
	nodeAddress := os.Getenv("AKAVE_NODE_ADDRESS")
	if nodeAddress == "" {
		log.Fatal("FATAL: AKAVE_NODE_ADDRESS environment variable not set. This is required.")
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
	defer func() {
		log.Println("Closing Akave SDK connection...")
		if closeErr := client.Close(); closeErr != nil {
			log.Printf("Error closing Akave SDK connection: %v", closeErr)
		} else {
			log.Println("Akave SDK connection closed successfully.")
		}
	}()

	// 2. Create a new server instance with the initialized client.
	srv := &server{
		client: client,
	}

	// 3. Register the handlers.
	mux := http.NewServeMux()
	mux.HandleFunc("/health", srv.healthHandler)


	log.Println("Starting go-akavelink server on :8080...")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}

}

// The actual main entry point for the executable.
func main() {
	MainFunc()
}


// healthHandler is a method on the server.
func (s *server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}