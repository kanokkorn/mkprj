#!/bin/sh
set -e

# --- Configuration ---
OWNER="kanokkorn"
REPO="mkprj"
BINARY_NAME="mkprj"
VERSION="v0.1.0"
INSTALL_DIR="/usr/local/bin"

# --- Detect OS and Arch ---
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# --- Determine Download URL ---
# Matches: mkprj-v0.1.0-linux-amd64.tar.gz
FILE_NAME="mkprj-${VERSION}-${OS}-${ARCH}.tar.gz"
URL="https://github.com/$OWNER/$REPO/releases/download/${VERSION}/${FILE_NAME}"

echo "Downloading $FILE_NAME..."

# --- Installation ---
# Create a temporary directory for extraction
TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

curl -sL "$URL" -o "$TMP_DIR/$FILE_NAME"

echo "Extracting..."
tar -xzf "$TMP_DIR/$FILE_NAME" -C "$TMP_DIR"

# Find the binary in the extracted files (assuming it's named mkgen or mkprj)
# If your binary inside the tar is named 'mkprj', change BINARY_NAME at the top
SOURCE_BINARY=$(find "$TMP_DIR" -type f -name "$BINARY_NAME" | head -n 1)

if [ -z "$SOURCE_BINARY" ]; then
    echo "Error: Could not find binary '$BINARY_NAME' in the archive."
    exit 1
fi

if [ -w "$INSTALL_DIR" ]; then
    mv "$SOURCE_BINARY" "$INSTALL_DIR/$BINARY_NAME"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
else
    echo "Requesting sudo permissions to install to $INSTALL_DIR"
    sudo mv "$SOURCE_BINARY" "$INSTALL_DIR/$BINARY_NAME"
    sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
fi

echo "Successfully installed $(basename "$SOURCE_BINARY") to $INSTALL_DIR"
