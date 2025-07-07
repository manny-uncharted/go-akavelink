#!/bin/bash

# ---
# This script sets up the necessary private key file, exports environment
# variables, and runs the go-akavelink server.
# ---

# Set color codes for output
BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}### Akave Project Setup & Run ###${NC}\n"

# --- Step 1: Get Private Key ---
echo -e "${YELLOW}Step 1: Securely storing your private key.${NC}"
read -sp "Please enter your 64-character hex private key: " user_private_key
echo # Newline after hidden input

# Validate that the key was entered
if [ -z "$user_private_key" ]; then
    echo -e "\n${RED}Error: No private key entered. Exiting.${NC}"
    exit 1
fi

if ! [[ $user_private_key =~ ^[a-fA-F0-9]{64}$ ]]; then
    echo -e "\n${RED}Error: Invalid key format. Please provide a 64-character hex string. Exiting.${NC}"
    exit 1
fi

# --- Step 2: Store Key in a Secure File ---
KEY_DIR="$HOME/.keys"
KEY_FILE="$KEY_DIR/akave.key"

echo "-> Creating directory ${KEY_DIR} (if it doesn't exist)..."
mkdir -p "$KEY_DIR"

echo "-> Saving your key to ${KEY_FILE}..."
echo -n "$user_private_key" > "$KEY_FILE"

echo "-> Setting secure permissions (read/write for user only)..."
chmod 600 "$KEY_FILE"
echo -e "${GREEN}Private key successfully stored and secured.${NC}\n"

# --- Step 3: Export Environment Variables ---
echo -e "${YELLOW}Step 2: Exporting environment variables for this session.${NC}"

echo "-> Exporting AKAVE_PRIVATE_KEY from file..."
export AKAVE_PRIVATE_KEY=$(cat "$KEY_FILE")

echo "-> Exporting AKAVE_NODE_ADDRESS..."
export AKAVE_NODE_ADDRESS="connect.akave.ai:5500"

echo -e "${GREEN}Environment variables are set and now you're good to go..${NC}\n"
