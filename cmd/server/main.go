package main // This MUST be the 'main' package

import (
	"log"
	"net/http"

	// Import your internal sdk package
	"github.com/akave-ai/go-akavelink/internal/sdk"
)

func main() {
	// Initialize the Akave SDK and its IPC client.
	// This calls the InitAkaveSDK function from your internal/sdk package.
	err := sdk.InitAkaveSDK()
	if err != nil {
		log.Fatalf("Application startup failed: %v", err)
	}

	// Register handlers from your internal/sdk package.
	http.HandleFunc("/health", sdk.HealthHandler)

	// Register download handler.
	// Using a catch-all route for demonstration.
	http.HandleFunc("/", sdk.DownloadHandler)

	log.Println("Starting go-akavelink server on :8080...")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
