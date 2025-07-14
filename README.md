
# go-akavelink

🚀 A Go-based HTTP server that wraps the Akave SDK, exposing Akave APIs over REST. The previous version of this repository was a CLI wrapper around the Akave SDK; refer to [akavelink](https://github.com/akave-ai/akavelink).

## Project Goals

* Provide a production-ready HTTP layer around the Akave SDK.
* Replace dependency on CLI-based wrappers.
* Facilitate integration of Akave storage into other systems via simple REST APIs.

---

## Dev Setup

Follow these steps to set up and run `go-akavelink` locally:

1.  **Clone the Repository:**

    ```bash
    git clone [https://github.com/akave-ai/go-akavelink](https://github.com/akave-ai/go-akavelink)
    cd go-akavelink
    ```

2.  **Obtain Akave Tokens and Private Key:**

    * Access the Akave Faucet: [https://faucet.akave.ai/](https://faucet.akave.ai/)
    * Add the Akave network to a wallet.
    * Claim tokens.
    * Obtain the private key from the wallet.

3.  **Configure Environment Variables:**
    Create a `.env` file in the root of the `go-akavelink` directory with the following content, replacing `YOUR_PRIVATE_KEY_HERE` with the obtained private key:

    ```
    AKAVE_PRIVATE_KEY="YOUR_PRIVATE_KEY_HERE"
    AKAVE_NODE_ADDRESS="connect.akave.ai:5500"
    ```

4.  **Run Setup Script (Recommended):**

    The `scripts/` directory contains helper scripts (`setup.sh` and `setup.bat`) to automate the environment variable export process.

    **For macOS/Linux:**
    Navigate to the `scripts` directory and grant execute permissions, then run the script:

    ```bash
    chmod +x scripts/setup.sh
    ./scripts/setup.sh
    ```

    This script exports the variables from the `.env` file into the current terminal session. To verify, run:

    ```bash
    echo $AKAVE_PRIVATE_KEY
    echo $AKAVE_NODE_ADDRESS
    ```

    These variables will persist for the current terminal session. For permanent environment variables, consider adding them to a shell's configuration file (e.g., `~/.bashrc`, `~/.zshrc`, or `~/.profile`).

    **For Windows PowerShell:**
    PowerShell execution policy might need adjustment to run scripts. Open PowerShell as an administrator and run:

    ```powershell
    Set-ExecutionPolicy RemoteSigned -Scope CurrentUser
    ```

    Then, it is recommended to use Windows Subsystem for Linux (WSL) or Git Bash to run the `.sh` script for a more native experience.

    Alternatively, to manually export variables in PowerShell:
    ```powershell
    Get-Content .env | ForEach-Object {
        $line = $_.Trim()
        if (-not ([string]::IsNullOrEmpty($line)) -and -not $line.StartsWith("#")) {
            $parts = $line.Split('=', 2)
            if ($parts.Length -eq 2) {
                $varName = $parts[0]
                $varValue = $parts[1].Trim('"')
                [System.Environment]::SetEnvironmentVariable($varName, $varValue, "Process")
                Write-Host "Exported variable: $varName"
            }
        }
    }
    ```

    To verify variables are loaded in the current PowerShell session, run:

    ```powershell
    Get-Item Env:AKAVE_PRIVATE_KEY
    Get-Item Env:AKAVE_NODE_ADDRESS
    ```

    These variables will persist for the current PowerShell session. For permanent environment variables, refer to Windows documentation on setting system or user-specific environment variables.

5.  **Install Go Modules:**
    Before running the server, ensure all Go modules are tidy and downloaded:

    ```bash
    go mod tidy
    ```

6.  **Run the Server:**

    ```bash
    go run ./cmd/server
    ```

    Output similar to the following should appear:

    ```
    2025/07/07 03:17:14 Starting go-akavelink server on :8080...
    ```

7.  **Verify Installation:**
    Visit `http://localhost:8080/health` in a web browser to verify that the server is running correctly.

---

## Project Structure

```markdown

go-akavelink/
├── cmd/                \# Main entrypoint for executables
│   └── server/
│       └── main.go     \# Starts the HTTP server
├── internal/           \# Internal logic, not intended for external consumption
│   └── sdk/            \# Wrapper around the Akave SDK
├── pkg/                \# Public packages (if needed)
├── docs/               \# Architecture, design, and other documentation
├── scripts/            \# Helper scripts (e.g., setup.sh for environment variables)
│   ├── setup.sh
│   └── setup.bat
├── go.mod              \# Go module definition file
├── README.md           \# Project overview and setup instructions
└── CONTRIBUTING.md     \# Guide for project contributors

```

---

## Contributing

This repository is open to contributions! See [`CONTRIBUTING.md`](./CONTRIBUTING.md).

* Check the [issue tracker](https://github.com/akave-ai/go-akavelink/issues) for `good first issue` and `help wanted` labels.
* Follow the pull request checklist and formatting conventions.
