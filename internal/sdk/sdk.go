// Package sdk provides a high-level client for interacting with the AkaveLink IPC API.
// It encapsulates bucket and file operations and manages the underlying SDK connection.
package sdk

import (
	"fmt"

	"github.com/akave-ai/akavesdk/sdk"
)

// Config defines parameters for initializing the IPC client.
type Config struct {
	// NodeAddress is the gRPC endpoint for the AkaveLink node, e.g., "localhost:9090".
	NodeAddress       string
	// MaxConcurrency controls parallelism for upload/download streams.
	MaxConcurrency    int
	// BlockPartSize specifies the size of each chunk in bytes.
	BlockPartSize     int64
	// UseConnectionPool toggles the SDK's connection pool.
	UseConnectionPool bool
	// PrivateKeyHex is the hex-encoded private key used for signing transactions.
	PrivateKeyHex     string
}

// Client wraps the AkaveLink IPC API and manages its SDK lifecycle.
type Client struct {
	*sdk.IPC
	core *sdk.SDK
}

// NewClient initializes the AkaveLink SDK and returns a configured IPC client.
// It returns an error if the private key is missing or the SDK cannot initialize.
func NewClient(cfg Config) (*Client, error) {
	if cfg.PrivateKeyHex == "" {
		return nil, fmt.Errorf("configuration error: missing PrivateKeyHex")
	}

	opts := []sdk.Option{
		sdk.WithPrivateKey(cfg.PrivateKeyHex),
	}

	core, err := sdk.New(
		cfg.NodeAddress,
		cfg.MaxConcurrency,
		cfg.BlockPartSize,
		cfg.UseConnectionPool,
		opts..., 
	)
	if err != nil {
		return nil, fmt.Errorf("SDK initialization failed: %w", err)
	}

	ipcClient, err := core.IPC()
	if err != nil {
		core.Close()
		return nil, fmt.Errorf("failed to obtain IPC interface: %w", err)
	}

	return &Client{IPC: ipcClient, core: core}, nil
}

// NewIPC returns a fresh IPC interface instance with an updated transaction nonce.
func (c *Client) NewIPC() (*sdk.IPC, error) {
	return c.core.IPC()
}

// Close terminates all underlying SDK connections.
func (c *Client) Close() error {
	return c.core.Close()
}
