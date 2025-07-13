// Package cmd provides the HTTP server that manages AkaveLink buckets and files.
//
// It exposes RESTful endpoints for health checks, bucket management, and file operations.
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
	"github.com/gorilla/mux"

	akavesdk "github.com/akave-ai/go-akavelink/internal/sdk"
	"github.com/akave-ai/go-akavelink/internal/utils"
)

// AkaveResponse defines the standard JSON response envelope.
type AkaveResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// server encapsulates dependencies for HTTP handlers.
type server struct {
	client *akavesdk.Client
}


// healthHandler responds with a simple status OK message.
func (s *server) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// deleteBucketHandler removes all files from a bucket and then deletes the bucket itself.
func (s *server) deleteBucketHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		s.writeErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	vars := mux.Vars(r)
	if _, ok := vars["bucketName"]; !ok {
		s.writeErrorResponse(w, http.StatusBadRequest, "bucketName is required")
		return
	}
	bucketName := vars["bucketName"]
	ctx := r.Context()

	log.Printf("Listing files in bucket '%s' to clear it.", bucketName)
	files, err := s.client.ListFiles(ctx, bucketName)
	if err != nil {
		s.writeErrorResponse(w, http.StatusInternalServerError, "Failed to list files for deletion: "+err.Error())
		return
	}

	log.Printf("Found %d files to delete.", len(files))
	for _, file := range files {
		ipc, err := s.client.NewIPC()
		if err != nil {
			s.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create IPC client for deletion: "+err.Error())
			return
		}

		log.Printf("Deleting file: %s from bucket: %s", file.Name, bucketName)
		if err := ipc.FileDelete(ctx, bucketName, file.Name); err != nil {
			s.writeErrorResponse(w, http.StatusInternalServerError, "Failed to delete file '"+file.Name+"': "+err.Error())
			return
		}
		time.Sleep(1 * time.Second)
	}
	log.Printf("All files in bucket '%s' have been deleted.", bucketName)

	log.Printf("Deleting empty bucket: %s", bucketName)
	if err := s.client.DeleteBucket(ctx, bucketName); err != nil {
		s.writeErrorResponse(w, http.StatusInternalServerError, "Failed to delete empty bucket: "+err.Error())
		return
	}

	log.Printf("Successfully deleted bucket and its contents: %s", bucketName)
	s.writeSuccessResponse(w, http.StatusOK, map[string]string{
		"message": "Bucket and all its contents deleted successfully",
	})
}


// createBucketHandler provisions a new bucket.
func (s *server) createBucketHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        s.writeErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
        return
    }

	vars := mux.Vars(r)
	if _, ok := vars["bucketName"]; !ok {
		s.writeErrorResponse(w, http.StatusBadRequest, "bucketName is required")
		return
	}
	bucketName := vars["bucketName"]
	ctx := r.Context()

    log.Printf("Attempting to create bucket: %s", bucketName)
    if err := s.client.CreateBucket(ctx, bucketName); err != nil {
        log.Printf("Error: Failed to create bucket: %v", err)
        s.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create bucket: "+err.Error())
        return
    }

    log.Printf("Successfully created bucket: %s", bucketName)
    s.writeSuccessResponse(w, http.StatusCreated, map[string]string{
        "message":    "Bucket created successfully",
        "bucketName": bucketName,
    })
}


// fileInfoHandler returns metadata for a given file.
func (s *server) fileInfoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	vars := mux.Vars(r)
	if _, ok := vars["bucketName"]; !ok {
		s.writeErrorResponse(w, http.StatusBadRequest, "bucketName is required")
		return
	}
	if _, ok := vars["fileName"]; !ok {
		s.writeErrorResponse(w, http.StatusBadRequest, "fileName is required")
		return
	}

	bucketName := vars["bucketName"]
	fileName := vars["fileName"]
	ctx := r.Context()

	log.Printf("Getting info for file: %s/%s", bucketName, fileName)
	fileInfo, err := s.client.FileInfo(ctx, bucketName, fileName)
	if err != nil {
		s.writeErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve file info: "+err.Error())
		return
	}

	s.writeSuccessResponse(w, http.StatusOK, fileInfo)
}


// listFilesHandler returns all files in the specified bucket.
func (s *server) listFilesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	vars := mux.Vars(r)
	if _, ok := vars["bucketName"]; !ok {
		s.writeErrorResponse(w, http.StatusBadRequest, "bucketName is required")
		return
	}
	bucketName := vars["bucketName"]
	ctx := r.Context()

	log.Printf("Fetching files for bucket: %s", bucketName)
	files, err := s.client.ListFiles(ctx, bucketName)
	if err != nil {
		log.Printf("Error: Failed to list files: %v", err)
		s.writeErrorResponse(w, http.StatusInternalServerError, "Failed to list files: "+err.Error())
		return
	}

	log.Printf("Successfully retrieved %d files from bucket: %s", len(files), bucketName)
	s.writeSuccessResponse(w, http.StatusOK, files)
}


// viewBucketHandler lists all buckets accessible to the client.
func (s *server) viewBucketHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	buckets, err := s.client.ListBuckets()
	if err != nil {
		http.Error(w, "failed to list buckets: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AkaveResponse{Success: true, Data: buckets})
}

// uploadHandler receives a multipart file and uploads it.
func (s *server) uploadHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	if _, ok := vars["bucketName"]; !ok {
		s.writeErrorResponse(w, http.StatusBadRequest, "bucketName is required")
		return
	}

	bucketName := vars["bucketName"]
	if bucketName == "" {
		s.writeErrorResponse(w, http.StatusBadRequest, "bucketName cannot be empty")
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, "Failed to parse multipart form: "+err.Error())
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, "Failed to retrieve file from form: "+err.Error())
		return
	}
	defer file.Close()

	ctx := r.Context()
	fileName := handler.Filename

	log.Printf("Initiating upload for '%s' to bucket '%s'", fileName, bucketName)
	fileUpload, err := s.client.CreateFileUpload(ctx, bucketName, fileName)
	if err != nil {
		log.Printf("Error creating file upload stream: %v", err)
		s.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create file upload stream")
		return
	}

	log.Printf("Uploading content for file: %s", fileName)
	finalMetadata, err := s.client.Upload(ctx, fileUpload, file)
	if err != nil {
		log.Printf("Error uploading file content: %v", err)
		s.writeErrorResponse(w, http.StatusInternalServerError, "Failed to upload file content")
		return
	}

	response := AkaveResponse{
		Success: true,
		Data: map[string]interface{}{
			"message":     "File uploaded successfully",
			"rootCID":     finalMetadata.RootCID,
			"bucketName":  finalMetadata.BucketName,
			"fileName":    finalMetadata.Name,
			"size":        finalMetadata.Size,
			"encodedSize": finalMetadata.EncodedSize,
			"createdAt":   finalMetadata.CreatedAt,
			"committedAt": finalMetadata.CommittedAt,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
	log.Printf("Successfully uploaded file '%s' with Root CID: %s", finalMetadata.Name, finalMetadata.RootCID)
}

// downloadHandler streams a requested file directly to the client.
func (s *server) downloadHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	if _, ok := vars["bucketName"]; !ok {
		s.writeErrorResponse(w, http.StatusBadRequest, "bucketName is required")
		return
	}
	if _, ok := vars["fileName"]; !ok {
		s.writeErrorResponse(w, http.StatusBadRequest, "fileName is required")
		return
	}
	bucketName := vars["bucketName"]
	fileName := vars["fileName"]

	ctx := r.Context()

	log.Printf("Initiating download for %s/%s", bucketName, fileName)
	fileDownload, err := s.client.CreateFileDownload(ctx, bucketName, fileName)
	if err != nil {
		s.writeErrorResponse(w, http.StatusNotFound, "Failed to create file download: "+err.Error())
		return
	}

	// Set headers for file download.
	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set("Content-Type", "application/octet-stream")

	// The Download method streams the file content directly into the http.ResponseWriter.
	if err := s.client.Download(ctx, fileDownload, w); err != nil {
		// Cannot write a JSON error here as the headers and potentially
		// part of the body have already been sent. We can only log the error.
		log.Printf("Error during file stream: %v", err)
		return
	}
	log.Printf("Successfully downloaded: %s/%s", bucketName, fileName)
}

// fileDeleteHandler handles the deletion of a file from a bucket.
func (s *server) fileDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		s.writeErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	vars := mux.Vars(r)
	if _, ok := vars["bucketName"]; !ok {
		s.writeErrorResponse(w, http.StatusBadRequest, "bucketName is required")
		return
	}
	if _, ok := vars["fileName"]; !ok {
		s.writeErrorResponse(w, http.StatusBadRequest, "fileName is required")
		return
	}

	bucketName := vars["bucketName"]
	fileName := vars["fileName"]
	ctx := r.Context()

	log.Printf("Attempting to delete file: %s/%s", bucketName, fileName)
	if err := s.client.FileDelete(ctx, bucketName, fileName); err != nil {
		log.Printf("Error: Failed to delete file: %v", err)
		s.writeErrorResponse(w, http.StatusInternalServerError, "Failed to delete file: "+err.Error())
		return
	}

	log.Printf("Successfully deleted file: %s/%s", bucketName, fileName)
	s.writeSuccessResponse(w, http.StatusOK, map[string]string{
		"message": "File deleted successfully",
	})
}

// writeSuccessResponse is a helper to standardize successful JSON responses.
func (s *server) writeSuccessResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	response := AkaveResponse{
		Success: true,
		Data:    data,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// writeErrorResponse is a helper to standardize error JSON responses.
func (s *server) writeErrorResponse(w http.ResponseWriter, statusCode int, errorMsg string) {
	response := AkaveResponse{
		Success: false,
		Error:   errorMsg,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

func MainFunc() {
	utils.LoadEnvConfig()

	key := os.Getenv("AKAVE_PRIVATE_KEY")
	node := os.Getenv("AKAVE_NODE_ADDRESS")
	if key == "" || node == "" {
		log.Fatal("AKAVE_PRIVATE_KEY and AKAVE_NODE_ADDRESS must be set")
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
		log.Fatalf("client init error: %v", err)
	}
	defer client.Close()


	// Set up HTTP server with routes
	r := mux.NewRouter()
    srv := &server{client: client}

    r.HandleFunc("/health", srv.healthHandler).Methods("GET")
    r.HandleFunc("/buckets/{bucketName}", srv.createBucketHandler).Methods("POST")
    r.HandleFunc("/buckets/{bucketName}", srv.deleteBucketHandler).Methods("DELETE")
    r.HandleFunc("/buckets/", srv.viewBucketHandler).Methods("GET")

    r.HandleFunc("/buckets/{bucketName}/files", srv.listFilesHandler).Methods("GET")
    r.HandleFunc("/buckets/{bucketName}/files", srv.uploadHandler).Methods("POST")
    r.HandleFunc("/buckets/{bucketName}/files/{fileName}", srv.fileInfoHandler).Methods("GET")
    r.HandleFunc("/buckets/{bucketName}/files/{fileName}/download", srv.downloadHandler).Methods("GET")
    r.HandleFunc("/buckets/{bucketName}/files/{fileName}", srv.fileDeleteHandler).Methods("DELETE")

    log.Println("Server listening on :8080")
    log.Fatal(http.ListenAndServe(":8080", r))
}

func main() {
	MainFunc()
}