package sdk // This remains the 'sdk' package

import (
	// Important: context must be imported for r.Context()
	"fmt" // Important: io must be imported for io.Writer in Download
	"log"
	"net/http"
	"strings"

	"github.com/akave-ai/akavesdk/sdk" // The actual Akave SDK package
)

// Global variable to hold the initialized Akave IPC client
// Exported because it's accessed by functions that will be called from 'main'
var IPCClient *sdk.IPC

// InitAkaveSDK initializes the main Akave SDK and its IPC client.
// This function should be called once at the start of your application.
// Exported (capitalized) to be callable from the 'main' package.
func InitAkaveSDK() error { // <-- Capitalized 'I'
	akaveNodeAddress := "localhost:50051"
	maxConcurrency := 10
	blockPartSize := int64(1024 * 1024)
	useConnectionPool := true

	newSDK, err := sdk.New(akaveNodeAddress, maxConcurrency, blockPartSize, useConnectionPool)
	if err != nil {
		return fmt.Errorf("failed to initialize Akave SDK: %w", err)
	}

	ipc, err := newSDK.IPC()
	if err != nil {
		return fmt.Errorf("failed to get IPC client from Akave SDK: %w", err)
	}
	IPCClient = ipc

	log.Println("Akave SDK and IPC client initialized successfully.")
	return nil
}

// HealthHandler responds with "ok" to health check requests.
// Exported (capitalized) to be callable from the 'main' package.
func HealthHandler(w http.ResponseWriter, r *http.Request) { // <-- Capitalized 'H'
	fmt.Fprintln(w, "ok")
}

// DownloadHandler handles file download requests.
// Expected URL format: /{bucketName}/files/{fileName}/download
// Exported (capitalized) to be callable from the 'main' package.
func DownloadHandler(w http.ResponseWriter, r *http.Request) { // <-- Capitalized 'D'
	if IPCClient == nil {
		http.Error(w, "Akave IPC client not initialized", http.StatusInternalServerError)
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 || parts[1] != "files" || parts[3] != "download" {
		http.NotFound(w, r)
		return
	}

	bucketName := parts[0]
	fileName := parts[2]

	ctx := r.Context()
	fileDownload, err := IPCClient.CreateFileDownload(ctx, bucketName, fileName)
	if err != nil {
		log.Printf("Error initiating file download for bucket '%s', file '%s': %v", bucketName, fileName, err)
		http.Error(w, fmt.Sprintf("failed to initiate file download: %v", err), http.StatusInternalServerError)
		return
	}

	err = IPCClient.Download(ctx, fileDownload, w)
	if err != nil {
		log.Printf("Error completing file download for bucket '%s', file '%s': %v", bucketName, fileName, err)
		http.Error(w, fmt.Sprintf("failed to complete file download: %v", err), http.StatusInternalServerError)
		return
	}
	log.Printf("Successfully downloaded file for bucket '%s', file '%s'", bucketName, fileName)
}

// No func main() here! It belongs only in package main.
