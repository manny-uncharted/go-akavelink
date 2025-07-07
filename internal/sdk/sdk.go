package sdk // This remains the 'sdk' package

import (
	// Important: context must be imported for r.Context()
	"fmt" // Important: io must be imported for io.Writer in Download

	"github.com/akave-ai/akavesdk/sdk" // The actual Akave SDK package
)

// Config holds the configuration parameters needed to initialize the Akave SDK.
// Using a struct for configuration makes the NewClient function signature clean
// and easy to extend if more parameters are needed in the future.
type Config struct {
	NodeAddress       string
	MaxConcurrency    int
	BlockPartSize     int64
	UseConnectionPool bool
}

// Client is a wrapper around the Akave SDK. It is designed to hold the
// necessary client instances (like the IPC client) and manage the SDK's lifecycle.
type Client struct {
	// By embedding the *sdk.IPC, our Client automatically gets all the methods
	// of the IPC client, such as CreateFileDownload and Download.
	// This makes our wrapper a transparent proxy for IPC operations.
	*sdk.IPC

	// We store the parent SDK instance to manage its lifecycle,
	// specifically for closing the connection gracefully.
	sdk *sdk.SDK
}

// NewClient creates and initializes a new Akave SDK client wrapper.
// It takes a configuration struct, connects to the Akave node,
// and prepares the IPC client for use.
func NewClient(cfg Config) (*Client, error) {
	// Initialize the main Akave SDK with the provided configuration.
	newSDK, err := sdk.New(
		cfg.NodeAddress,
		cfg.MaxConcurrency,
		cfg.BlockPartSize,
		cfg.UseConnectionPool,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Akave SDK: %w", err)
	}

	// Get the specific IPC (Inter-Process Communication) client from the SDK.
	// Your application logic primarily uses this for file operations.
	ipc, err := newSDK.IPC()
	if err != nil {
		// If we can't get the IPC client, we should clean up the SDK connection.
		// We assume the SDK has a Close() method for this.
		newSDK.Close()
		return nil, fmt.Errorf("failed to get IPC client from Akave SDK: %w", err)
	}

	// Return our new wrapper client, containing both the IPC client and the main SDK instance.
	return &Client{
		IPC: ipc,
		sdk: newSDK,
	}, nil
}

// Close gracefully shuts down the connection to the Akave SDK.
// It's important to call this on application shutdown to release resources.
func (c *Client) Close() error {
	fmt.Println("Closing Akave SDK connection...")
	return c.sdk.Close() // Assuming the SDK provides a Close() method.
}