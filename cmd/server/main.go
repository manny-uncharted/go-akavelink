package main

import (
	"fmt"
	"log"
	"net/http"
	"encoding/json"
	"strings"
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

	// 3. Register the handlers. The handlers are now methods on our server struct,
	// which gives them access to the client.
	http.HandleFunc("/health", srv.healthHandler)

	log.Println("Starting go-akavelink server on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}


// routeHandler directs traffic based on method and URL pattern.
func (s *server) routeHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	// Routing for /:bucket_id/files/...
	if len(parts) >= 3 && parts[1] == "files" {
		if r.Method == http.MethodPost && len(parts) == 3 && parts[2] == "upload" {
			s.uploadHandler(w, r)
			return
		}
		if r.Method == http.MethodGet && len(parts) == 4 && parts[3] == "download" {
			s.downloadHandler(w, r)
			return
		}
	}

	http.NotFound(w, r)
}

// downloadHandler is also a method on the server, giving it access to s.client.
func (s *server) downloadHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 || parts[1] != "files" || parts[3] != "download" {
		http.NotFound(w, r)
		return
	}

	bucketName := parts[0]
	fileName := parts[2]

	ctx := r.Context()
	// Use the client from the server struct, not a global variable.
	fileDownload, err := s.client.CreateFileDownload(ctx, bucketName, fileName)
	if err != nil {
		log.Printf("Error: Failed to create file download: %v", err)
		http.Error(w, "failed to create file download", http.StatusInternalServerError)
		return
	}

	// The Download method is available directly on our client because we embedded *sdk.IPC.
	if err := s.client.Download(ctx, fileDownload, w); err != nil {
		log.Printf("Error: Failed to complete file download: %v", err)
		// Note: Can't write a new HTTP error if the download has already started writing to the response.
		return
	}

	log.Printf("Successfully downloaded: %s/%s", bucketName, fileName)
}

// uploadHandler handles multipart file uploads.
func (s *server) uploadHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	bucketName := parts[0]

	// 1. Parse the multipart form.
	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB max memory
		http.Error(w, "Failed to parse multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 2. Get the file from the form data.
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to retrieve file from form: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	ctx := r.Context()
	fileName := handler.Filename

	// 3. Initiate the file upload stream. This step is correct.
	log.Printf("Initiating upload for '%s' to bucket '%s'", fileName, bucketName)
	fileUpload, err := s.client.CreateFileUpload(ctx, bucketName, fileName)
	if err != nil {
		log.Printf("Error creating file upload stream: %v", err)
		http.Error(w, "Failed to create file upload stream", http.StatusInternalServerError)
		return
	}

	// 4. Upload the file content using the stream.
	// This function is long-running. It chunks, uploads, and commits the file.
	log.Printf("Uploading content for file: %s", fileName)
	finalMetadata, err := s.client.Upload(ctx, fileUpload, file)
	if err != nil {
		log.Printf("Error uploading file content: %v", err)
		http.Error(w, "Failed to upload file content", http.StatusInternalServerError)
		return
	}

	// 5. Respond with the final metadata returned by the Upload function.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	
    // Create a new map for a clean JSON response
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


// healthHandler is now a method on the server.
func (s *server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

