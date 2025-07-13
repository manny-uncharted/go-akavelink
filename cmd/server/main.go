// cmd/server/main.go

package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	akavesdk "github.com/akave-ai/go-akavelink/internal/sdk"
	"github.com/akave-ai/go-akavelink/internal/utils"
)

// AkaveResponse is our JSON envelope.
type AkaveResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type server struct {
	client *akavesdk.Client
}

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

	bucketName := r.PathValue("bucketName")
	ctx := r.Context()

	// First, list all files in the bucket.
	log.Printf("Listing files in bucket '%s' to clear it.", bucketName)
	files, err := s.client.ListFiles(ctx, bucketName)
	if err != nil {
		s.writeErrorResponse(w, http.StatusInternalServerError, "Failed to list files for deletion: "+err.Error())
		return
	}

	// Loop through and delete each file.
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
		// Extend the pause to 1 second to allow the network to process the transaction.
		time.Sleep(1 * time.Second)
	}
	log.Printf("All files in bucket '%s' have been deleted.", bucketName)

	// Finally, delete the now-empty bucket.
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


// listFilesHandler returns a list of all files within a specific bucket.
func (s *server) listFilesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	bucketName := r.PathValue("bucketName")
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



func (s *server) bucketsHandler(w http.ResponseWriter, r *http.Request) {
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

// uploadHandler processes multipart file uploads, handling overwrites via a query parameter.
func (s *server) uploadHandler(w http.ResponseWriter, r *http.Request) {
	bucketName := r.PathValue("bucketName")
	overwrite := r.URL.Query().Get("overwrite") == "true"

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, "Failed to parse multipart form: "+err.Error())
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, "Failed to retrieve file from form")
		return
	}
	defer file.Close()

	ctx := r.Context()
	fileName := handler.Filename

	if overwrite {
		log.Printf("Overwrite flag set. Attempting to delete existing file: %s/%s", bucketName, fileName)
		if err := s.client.FileDelete(ctx, bucketName, fileName); err != nil {
			log.Printf("Error: Failed to delete existing file during overwrite: %v", err)
			s.writeErrorResponse(w, http.StatusInternalServerError, "Failed to delete existing file: "+err.Error())
			return
		}
		log.Printf("Successfully deleted existing file for overwrite.")
	}

	log.Printf("Initiating upload for '%s' to bucket '%s'", fileName, bucketName)
	fileUpload, err := s.client.CreateFileUpload(ctx, bucketName, fileName)
	if err != nil {
		// Specifically check for the FileAlreadyExists error to provide a helpful response.
		if strings.Contains(err.Error(), "FileAlreadyExists") {
			msg := "File already exists. To overwrite, add the '?overwrite=true' query parameter."
			s.writeErrorResponse(w, http.StatusConflict, msg)
			return
		}
		s.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create file upload stream: "+err.Error())
		return
	}

	log.Printf("Uploading content for file: %s", fileName)
	finalMetadata, err := s.client.Upload(ctx, fileUpload, file)
	if err != nil {
		s.writeErrorResponse(w, http.StatusInternalServerError, "Failed to upload file content: "+err.Error())
		return
	}

	s.writeSuccessResponse(w, http.StatusCreated, finalMetadata)
	log.Printf("Successfully uploaded file '%s' with Root CID: %s", finalMetadata.Name, finalMetadata.RootCID)
}

// downloadHandler streams a requested file directly to the client.
func (s *server) downloadHandler(w http.ResponseWriter, r *http.Request) {
	bucketName := r.PathValue("bucketName")
	fileName := r.PathValue("fileName")

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

	bucketName := r.PathValue("bucketName")
	fileName := r.PathValue("fileName")
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

	srv := &server{client: client}
	mux := http.NewServeMux()
	mux.HandleFunc("/health", srv.healthHandler)
	mux.HandleFunc("/buckets", srv.bucketsHandler)
	mux.HandleFunc("DELETE /buckets/delete/{bucketName}", srv.deleteBucketHandler)
	mux.HandleFunc("POST /files/upload/{bucketName}", srv.uploadHandler)
	mux.HandleFunc("GET /files/download/{bucketName}/{fileName}", srv.downloadHandler)
	mux.HandleFunc("DELETE /files/delete/{bucketName}/{fileName}", srv.fileDeleteHandler)
	mux.HandleFunc("GET /files/list/{bucketName}", srv.listFilesHandler)

	log.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func main() {
	MainFunc()
}