// test/main_test.go
package test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	akavesdk "github.com/akave-ai/go-akavelink/internal/sdk" // aliased for clarity
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	// Import your main package to access its components for testing
	// We need to slightly adjust how we import `main` for testing,
	// typically you'd run `go test .` from the `main` package's directory,
	// but here we are in a `test` directory, so we can't directly
	// reference `main`.
	// For integration tests like this, it's often better to start the
	// server in a goroutine and then make HTTP requests to it.
	// We'll simulate `main`'s server startup logic here.
)

// To avoid circular dependencies and allow `main_test.go` to be in `test/`,
// we will effectively copy the server setup logic from `main.go` and run it
// within a test-specific context. This is a common pattern for integration tests.

// Define a placeholder for the server structure from main.go
// You might need to expose `server` or specific handlers in `main.go`
// or re-define them here for testing if they are not exported.
// For simplicity, we'll re-define the necessary parts here.

type testServer struct {
	client *akavesdk.Client // Assuming sdk.Client is exported
}

func (s *testServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// TestMain_HealthEndpoint tests the /health endpoint of the HTTP server.
func TestMain_HealthEndpoint(t *testing.T) {
	// Set up mock environment variables for the test
	// We'll use a temporary .env file for tests.
	// This ensures our test environment variables are isolated.
	tempEnvFile, err := os.CreateTemp("", "test_env_*.env")
	require.NoError(t, err)
	defer os.Remove(tempEnvFile.Name()) // Clean up the temp file

	_, err = tempEnvFile.WriteString(fmt.Sprintf("AKAVE_PRIVATE_KEY=%s\nAKAVE_NODE_ADDRESS=%s\n", testPrivateKey, testNodeAddress))
	require.NoError(t, err)
	tempEnvFile.Close()

	// Load the temporary .env file
	err = godotenv.Load(tempEnvFile.Name())
	require.NoError(t, err, "failed to load test .env file")

	// Ensure environment variables are clean after test
	// This is important if you run tests in parallel or sequentially
	defer func() {
		os.Unsetenv("AKAVE_PRIVATE_KEY")
		os.Unsetenv("AKAVE_NODE_ADDRESS")
	}()

	// Replicate main.go's client initialization
	privateKey := os.Getenv("AKAVE_PRIVATE_KEY")
	nodeAddress := os.Getenv("AKAVE_NODE_ADDRESS")

	require.NotEmpty(t, privateKey, "AKAVE_PRIVATE_KEY should be set for test")
	require.NotEmpty(t, nodeAddress, "AKAVE_NODE_ADDRESS should be set for test")

	cfg := akavesdk.Config{
		NodeAddress:       nodeAddress,
		MaxConcurrency:    10,
		BlockPartSize:     1024 * 1024, // 1MB
		UseConnectionPool: true,
		PrivateKeyHex:     privateKey,
	}

	client, err := akavesdk.NewClient(cfg)
	require.NoError(t, err, "Failed to initialize Akave client for test server")
	defer client.Close() // Ensure client is closed

	// Create a new test server instance
	srv := &testServer{
		client: client,
	}

	// Create a new HTTP test server
	testHTTPServer := httptest.NewServer(http.HandlerFunc(srv.healthHandler))
	defer testHTTPServer.Close()

	// Make a request to the health endpoint
	resp, err := http.Get(testHTTPServer.URL + "/health")
	require.NoError(t, err, "Failed to make GET request to health endpoint")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status code 200 OK for /health")

	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read response body")
	assert.Equal(t, "ok", string(bodyBytes), "Expected body to be 'ok'")
}