package main

import (
	"fmt"
	"log"
	"net/http"
	"encoding/json"
	"os"


	akavesdk "github.com/akave-ai/go-akavelink/internal/sdk" // aliased for clarity
	
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

	mux := http.NewServeMux()
	mux.HandleFunc("/health", srv.healthHandler)

	// --- NEW: Register routes with methods and wildcards ---
	mux.HandleFunc("POST /files/upload/{bucketName}", srv.uploadHandler)
	mux.HandleFunc("GET /files/download/{bucketName}/{fileName}", srv.downloadHandler)

	log.Println("Starting go-akavelink server on :8080...")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// uploadHandler is updated to get the bucket name from the path value.
func (s *server) uploadHandler(w http.ResponseWriter, r *http.Request) {
	// --- NEW: Get bucketName from the URL path ---
	bucketName := r.PathValue("bucketName")

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Failed to parse multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to retrieve file from form: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	ctx := r.Context()
	fileName := handler.Filename

	log.Printf("Initiating upload for '%s' to bucket '%s'", fileName, bucketName)
	fileUpload, err := s.client.CreateFileUpload(ctx, bucketName, fileName)
	if err != nil {
		log.Printf("Error creating file upload stream: %v", err)
		http.Error(w, "Failed to create file upload stream", http.StatusInternalServerError)
		return
	}

	log.Printf("Uploading content for file: %s", fileName)
	finalMetadata, err := s.client.Upload(ctx, fileUpload, file)
	if err != nil {
		log.Printf("Error uploading file content: %v", err)
		http.Error(w, "Failed to upload file content", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := map[string]interface{}{
		"message":     "File uploaded successfully",
		"rootCID":     finalMetadata.RootCID,
		"bucketName":  finalMetadata.BucketName,
		"fileName":    finalMetadata.Name,
		"size":        finalMetadata.Size,
		"encodedSize": finalMetadata.EncodedSize,
		"committedAt": finalMetadata.CommittedAt,
	}
	json.NewEncoder(w).Encode(response)
	log.Printf("Successfully uploaded file '%s' with Root CID: %s", finalMetadata.Name, finalMetadata.RootCID)
}

// downloadHandler is updated to get parameters from the path value.
func (s *server) downloadHandler(w http.ResponseWriter, r *http.Request) {
	// --- NEW: Get parameters directly from the URL path ---
	bucketName := r.PathValue("bucketName")
	fileName := r.PathValue("fileName")

	ctx := r.Context()
	fileDownload, err := s.client.CreateFileDownload(ctx, bucketName, fileName)
	if err != nil {
		log.Printf("Error: Failed to create file download: %v", err)
		http.Error(w, "failed to create file download", http.StatusInternalServerError)
		return
	}

	if err := s.client.Download(ctx, fileDownload, w); err != nil {
		log.Printf("Error: Failed to complete file download: %v", err)
		return
	}
	log.Printf("Successfully downloaded: %s/%s", bucketName, fileName)
}

func (s *server) healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "ok")
}