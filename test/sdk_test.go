package test

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
)

// Declare variables to hold values from environment variables
var (
	testPrivateKey  string
	testNodeAddress string
)

func init() {
	// Function to find the Go module root (where go.mod is located)
	// This is generally the project root where .env would reside.
	findModuleRoot := func() (string, error) {
		_, filename, _, ok := runtime.Caller(0)
		if !ok {
			return "", nil // Or an error, depending on strictness
		}
		currentDir := filepath.Dir(filename)

		for {
			goModPath := filepath.Join(currentDir, "go.mod")
			if _, err := os.Stat(goModPath); err == nil {
				return currentDir, nil // Found go.mod, this is the root
			}
			parentDir := filepath.Dir(currentDir)
			if parentDir == currentDir { // Reached file system root
				break
			}
			currentDir = parentDir
		}
		return "", nil // Or an error if go.mod isn't found
	}

	// 1. Check for an explicit DOTENV_PATH environment variable
	dotenvPath := os.Getenv("DOTENV_PATH")

	if dotenvPath == "" {
		// 2. If DOTENV_PATH is not set, try to find the module root and load .env from there
		moduleRoot, err := findModuleRoot()
		if err != nil {
			log.Printf("Warning: Failed to find Go module root: %v. .env won't be loaded automatically.", err)
		} else if moduleRoot != "" {
			dotenvPath = filepath.Join(moduleRoot, ".env")
		} else {
			log.Printf("Warning: Could not determine Go module root. .env won't be loaded automatically.")
		}
	}

	// 3. Load the .env file if a path was determined
	if dotenvPath != "" {
		err := godotenv.Load(dotenvPath)
		if err != nil {
			log.Printf("Warning: Could not load .env file from %s: %v. Relying on system environment variables.", dotenvPath, err)
		} else {
			log.Printf("Successfully loaded .env file from %s.", dotenvPath)
		}
	} else {
		log.Println("Warning: No .env path determined. Relying solely on system environment variables.")
	}

	// Retrieve private key from environment variable (will now include values from .env if loaded)
	testPrivateKey = os.Getenv("AKAVE_PRIVATE_KEY")
	// if testPrivateKey == "" {
	// 	testPrivateKey = "e11da8d70c0ef001264b59dc2f" // Fallback mock key
	// 	log.Println("AKAVE_PRIVATE_KEY not set in environment or .env, using mock private key for tests.")
	// }

	// Retrieve node address from environment variable
	testNodeAddress = os.Getenv("AKAVE_NODE_ADDRESS")
	// if testNodeAddress == "" {
	// 	testNodeAddress = "connect.akave.ai:5500" // Fallback remote node
	// 	log.Println("AKAVE_NODE_ADDRESS not set in environment or .env, using fallback node address for tests.")
	// }
}

// ... rest of your test functions ...
