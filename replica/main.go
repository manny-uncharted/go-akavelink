// cmd/server/main.go

package main

import (
    "encoding/json"
    "log"
    "net/http"
    "os"
    "strings"

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

func (s *server) uploadHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // extract bucketName from /files/upload/{bucketName}
    parts := strings.Split(r.URL.Path, "/")
    if len(parts) < 4 || parts[3] == "" {
        http.Error(w, "bucketName missing in path", http.StatusBadRequest)
        return
    }
    bucketName := parts[3]

    if err := r.ParseMultipartForm(32 << 20); err != nil {
        http.Error(w, "failed to parse form: "+err.Error(), http.StatusBadRequest)
        return
    }

    file, handler, err := r.FormFile("file")
    if err != nil {
        http.Error(w, "file retrieval error: "+err.Error(), http.StatusBadRequest)
        return
    }
    defer file.Close()

    ctx := r.Context()
    // Attempt to initialize upload stream
    uploadStream, err := s.client.CreateFileUpload(ctx, bucketName, handler.Filename)
    if err != nil {
        // if bucket doesn't exist, create it and retry
        if strings.Contains(err.Error(), "BucketNonexists") {
            if err2 := s.client.CreateBucket(ctx, bucketName); err2 != nil {
                http.Error(w, "bucket creation failed: "+err2.Error(), http.StatusInternalServerError)
                return
            }
            // retry upload stream initialization
            uploadStream, err = s.client.CreateFileUpload(ctx, bucketName, handler.Filename)
        }
        if err != nil {
            http.Error(w, "upload init failed: "+err.Error(), http.StatusInternalServerError)
            return
        }
    }

    // Now stream the file
    meta, err := s.client.Upload(ctx, uploadStream, file)
    if err != nil {
        http.Error(w, "upload failed: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    resp := map[string]interface{}{
        "message":     "File uploaded successfully",
        "rootCID":     meta.RootCID,
        "bucketName":  meta.BucketName,
        "fileName":    meta.Name,
        "size":        meta.Size,
        "encodedSize": meta.EncodedSize,
        "committedAt": meta.CommittedAt,
    }
    json.NewEncoder(w).Encode(AkaveResponse{Success: true, Data: resp})
}


func (s *server) downloadHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // extract bucketName & fileName from /files/download/{bucketName}/{fileName}
    parts := strings.Split(r.URL.Path, "/")
    if len(parts) < 5 {
        http.Error(w, "path must be /files/download/{bucket}/{file}", http.StatusBadRequest)
        return
    }
    bucketName := parts[3]
    fileName := parts[4]

    dlStream, err := s.client.CreateFileDownload(r.Context(), bucketName, fileName)
    if err != nil {
        http.Error(w, "download init failed: "+err.Error(), http.StatusInternalServerError)
        return
    }

    if err := s.client.Download(r.Context(), dlStream, w); err != nil {
        log.Printf("download error: %v", err)
    }
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
    mux.HandleFunc("/files/upload/", srv.uploadHandler)
    mux.HandleFunc("/files/download/", srv.downloadHandler)

    log.Println("Server listening on :8080")
    log.Fatal(http.ListenAndServe(":8080", mux))
}

func main() {
    MainFunc()
}
