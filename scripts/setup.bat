@echo off
rem ---
rem This script sets up the necessary private key file, exports environment
rem variables for the current session, and runs the go-akavelink server.
rem ---

title Akave Project Setup

echo ### Akave Project Setup ###
echo.

rem --- Step 1: Get Private Key ---
echo Step 1: Storing your private key.
echo IMPORTANT: Your key will be visible as you type.
set /p "user_private_key=Please enter your 64-character hex private key: "

rem Validate that the key was entered
if not defined user_private_key (
    echo.
    echo Error: No private key entered. Exiting.
    pause
    exit /b 1
)

rem Clear the screen to hide the key from view
cls

echo ### Akave Project Setup ###
echo.
echo Step 1: Storing your private key. ... Done.
echo.

rem --- Step 2: Store Key in a Secure File ---
set "KEY_DIR=%USERPROFILE%\.keys"
set "KEY_FILE=%KEY_DIR%\akave.key"

echo Step 2: Saving key to a secure file.
echo -> Creating directory %KEY_DIR% (if it doesn't exist)...
if not exist "%KEY_DIR%" mkdir "%KEY_DIR%"

echo -> Saving your key to %KEY_FILE%...
echo|set /p="!user_private_key!" > "%KEY_FILE%"

echo -> Setting secure permissions (user access only)...
rem Grant the current user full access and remove inherited permissions
icacls "%KEY_FILE%" /grant "%USERNAME%":(F) /inheritance:r > nul

echo Private key successfully stored and secured.
echo.

rem --- Step 3: Set Environment Variables ---
echo Step 3: Setting environment variables for this session.

echo -> Setting AKAVE_PRIVATE_KEY from file...
set /p AKAVE_PRIVATE_KEY=<"%KEY_FILE%"

echo -> Setting AKAVE_NODE_ADDRESS...
set "AKAVE_NODE_ADDRESS=connect.akave.ai:5500"

echo Environment variables are set. You're good to go.
echo.
