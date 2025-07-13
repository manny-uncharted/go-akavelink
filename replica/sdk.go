// internal/sdk/sdk.go

package sdk

import (
    "context"
    "fmt"
    "io"

    "github.com/akave-ai/akavesdk/sdk"
)

// Config holds configuration for the Akave SDK client.
type Config struct {
    NodeAddress       string
    MaxConcurrency    int
    BlockPartSize     int64
    UseConnectionPool bool
    PrivateKeyHex     string
}

// Client wraps the official Akave SDK's IPC interface and core instance.
type Client struct {
    *sdk.IPC
    sdk *sdk.SDK
}

// NewClient constructs and returns a Client using the provided Config.
func NewClient(cfg Config) (*Client, error) {
    if cfg.PrivateKeyHex == "" {
        return nil, fmt.Errorf("private key is required for IPC client but was not provided")
    }

    sdkOpts := []sdk.Option{
        sdk.WithPrivateKey(cfg.PrivateKeyHex),
    }

    core, err := sdk.New(
        cfg.NodeAddress,
        cfg.MaxConcurrency,
        cfg.BlockPartSize,
        cfg.UseConnectionPool,
        sdkOpts...,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to initialize Akave SDK: %w", err)
    }

    ipc, err := core.IPC()
    if err != nil {
        core.Close()
        return nil, fmt.Errorf("failed to get IPC client from Akave SDK: %w", err)
    }

    return &Client{IPC: ipc, sdk: core}, nil
}

// Close gracefully shuts down the underlying SDK connection.
func (c *Client) Close() error {
    return c.sdk.Close()
}

// CreateBucket provisions a new bucket under the caller’s key.
func (c *Client) CreateBucket(ctx context.Context, bucketName string) error {
    // call through to the IPC layer
    if _, err := c.IPC.CreateBucket(ctx, bucketName); err != nil {
        return fmt.Errorf("failed to create bucket %q: %w", bucketName, err)
    }
    return nil
}

// ListBuckets returns the names of all buckets accessible to this client.
func (c *Client) ListBuckets() ([]string, error) {
    buckets, err := c.IPC.ListBuckets(context.Background())
    if err != nil {
        return nil, fmt.Errorf("failed to list buckets: %w", err)
    }

    names := make([]string, len(buckets))
    for i, b := range buckets {
        names[i] = b.Name
    }
    return names, nil
}

// CreateFileUpload opens a new upload session for the given bucket and file name.
func (c *Client) CreateFileUpload(ctx context.Context, bucket, fileName string) (*sdk.IPCFileUpload, error) {
    return c.IPC.CreateFileUpload(ctx, bucket, fileName)
}

// Upload streams the given reader into the established upload session.
func (c *Client) Upload(ctx context.Context, upload *sdk.IPCFileUpload, reader io.Reader) (sdk.IPCFileMetaV2, error) {
    return c.IPC.Upload(ctx, upload, reader)
}

// CreateFileDownload opens a download session for the specified bucket and file.
func (c *Client) CreateFileDownload(ctx context.Context, bucket, fileName string) (sdk.IPCFileDownload, error) {
    return c.IPC.CreateFileDownload(ctx, bucket, fileName)
}

// Download writes the content of the download session to the provided writer.
func (c *Client) Download(ctx context.Context, download sdk.IPCFileDownload, writer io.Writer) error {
    return c.IPC.Download(ctx, download, writer)
}
