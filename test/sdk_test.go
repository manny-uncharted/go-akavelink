package test

import (
	"log"
	"os" // Keep this import if you still use it for anything else, though not strictly needed now for .env loading
	// Keep this import if you still use it for anything else, though not strictly needed now for .env loading
	"testing"

	akavesdk "github.com/akave-ai/go-akavelink/internal/sdk"
	"github.com/akave-ai/go-akavelink/internal/utils" // Import your new utils package
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Declare variables to hold values from environment variables
var (
	testPrivateKey  string
	testNodeAddress string
)

// init function runs before any tests in the package
func init() {
	// Load environment variables using the reusable function
	// This will try DOTENV_PATH first, then fallback to module root .env
	utils.LoadEnvConfig()

	// Retrieve private key from environment variable (will now include values from .env if loaded)
	testPrivateKey = os.Getenv("AKAVE_PRIVATE_KEY")
	if testPrivateKey == "" {
		// Provide a fallback mock key for tests if no real key is set.
		// For TestNewClient_Success, this will cause it to skip unless a real key is provided.
		testPrivateKey = "e11da8d70c0ef001264b59dc2f"
		log.Println("AKAVE_PRIVATE_KEY not set in environment or .env, using mock private key for tests.")
	}

	// Retrieve node address from environment variable
	testNodeAddress = os.Getenv("AKAVE_NODE_ADDRESS")
	if testNodeAddress == "" {
		// Provide a fallback for tests.
		testNodeAddress = "connect.akave.ai:5500" // Fallback to a common remote test address
		log.Println("AKAVE_NODE_ADDRESS not set in environment or .env, using fallback node address for tests.")
	}
}

// TestNewClient_Success tests successful client initialization
func TestNewClient_Success(t *testing.T) {
	// Ensure that if AKAVE_PRIVATE_KEY is not set (e.g., in a CI environment
	// where real connectivity isn't expected or configured), we skip this test.
	if os.Getenv("AKAVE_PRIVATE_KEY") == "" {
		t.Skip("AKAVE_PRIVATE_KEY environment variable not set, skipping TestNewClient_Success as it requires a real key.")
	}

	cfg := akavesdk.Config{
		NodeAddress:       testNodeAddress,
		MaxConcurrency:    1,
		BlockPartSize:     1024,
		UseConnectionPool: false,
		PrivateKeyHex:     testPrivateKey,
	}

	client, err := akavesdk.NewClient(cfg)
	require.NoError(t, err, "NewClient should not return an error with valid config")
	require.NotNil(t, client, "NewClient should return a non-nil client")

	// Ensure Close is called to release resources
	defer func() {
		err := client.Close()
		assert.NoError(t, err, "client.Close() should not return an error")
	}()

	// Add more specific assertions here if there were client internal states to check.
}

// TestNewClient_MissingPrivateKey tests client initialization without a private key
func TestNewClient_MissingPrivateKey(t *testing.T) {
	// Temporarily unset the private key for this test to ensure it fails as expected
	originalPrivateKey := os.Getenv("AKAVE_PRIVATE_KEY")
	os.Setenv("AKAVE_PRIVATE_KEY", "")
	defer os.Setenv("AKAVE_PRIVATE_KEY", originalPrivateKey) // Restore after test

	cfg := akavesdk.Config{
		NodeAddress:       testNodeAddress,
		MaxConcurrency:    1,
		BlockPartSize:     1024,
		UseConnectionPool: false,
		PrivateKeyHex:     "", // This will now be an empty string, simulating missing
	}

	client, err := akavesdk.NewClient(cfg)
	require.Error(t, err, "NewClient should return an error if private key is missing")
	assert.Contains(t, err.Error(), "private key is required", "Error message should indicate missing private key")
	assert.Nil(t, client, "NewClient should return a nil client on error")
}

// TestNewClient_SDKInitializationFailure tests a scenario where the underlying SDK fails
func TestNewClient_SDKInitializationFailure(t *testing.T) {
	// Temporarily set an invalid private key for this test
	originalPrivateKey := os.Getenv("AKAVE_PRIVATE_KEY")
	os.Setenv("AKAVE_PRIVATE_KEY", "0xinvalidkey")           // This specifically triggers the "invalid hex character" error
	defer os.Setenv("AKAVE_PRIVATE_KEY", originalPrivateKey) // Restore after test

	invalidPrivateKey := os.Getenv("AKAVE_PRIVATE_KEY") // Will now be "0xinvalidkey"

	cfg := akavesdk.Config{
		NodeAddress:       testNodeAddress,
		MaxConcurrency:    1,
		BlockPartSize:     1024,
		UseConnectionPool: false,
		PrivateKeyHex:     invalidPrivateKey,
	}

	client, err := akavesdk.NewClient(cfg)
	require.Error(t, err, "NewClient should return an error with an invalid private key (simulating SDK init failure)")
	// Updated assertion to match the actual error message from the SDK
	assert.Contains(t, err.Error(), "invalid hex character", "Error message should indicate an invalid private key")
	assert.Nil(t, client, "NewClient should return a nil client on SDK init error")
}
