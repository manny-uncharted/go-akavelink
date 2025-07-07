package sdk

import (
	"fmt"

	"github.com/akave-ai/akavesdk/sdk"
)

type Config struct {
	NodeAddress       string
	MaxConcurrency    int
	BlockPartSize     int64
	UseConnectionPool bool
	PrivateKeyHex     string
}

type Client struct {
	*sdk.IPC
	sdk *sdk.SDK
}

func NewClient(cfg Config) (*Client, error) {
	// Add a check to ensure the private key is provided.
	if cfg.PrivateKeyHex == "" {
		return nil, fmt.Errorf("private key is required for IPC client but was not provided")
	}

	sdkOpts := []sdk.Option{
		sdk.WithPrivateKey(cfg.PrivateKeyHex),
	}

	// Initialize the main Akave SDK with the base config AND our new options.
	newSDK, err := sdk.New(
		cfg.NodeAddress,
		cfg.MaxConcurrency,
		cfg.BlockPartSize,
		cfg.UseConnectionPool,
		sdkOpts...,
	)
	if err != nil {
		// If New() fails, it's likely due to an invalid key or other config.
		return nil, fmt.Errorf("failed to initialize Akave SDK: %w", err)
	}

	// Now, this call will succeed because newSDK already holds the private key.
	ipc, err := newSDK.IPC()
	if err != nil {
		newSDK.Close()
		return nil, fmt.Errorf("failed to get IPC client from Akave SDK: %w", err)
	}

	return &Client{
		IPC: ipc,
		sdk: newSDK,
	}, nil
}

// Close gracefully shuts down the connection to the Akave SDK.
// It's important to call this on application shutdown to release resources.
func (c *Client) Close() error {
	fmt.Println("Closing Akave SDK connection...")
	return c.sdk.Close()
}
