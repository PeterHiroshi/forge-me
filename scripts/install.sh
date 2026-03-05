#!/bin/bash
set -e

# cfmon installer for macOS and Linux
# Usage: curl -sSL https://raw.githubusercontent.com/PeterHiroshi/cfmon/main/scripts/install.sh | bash

REPO="PeterHiroshi/cfmon"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="cfmon"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$OS" in
  darwin)
    OS="Darwin"
    ;;
  linux)
    OS="Linux"
    ;;
  *)
    echo "Unsupported OS: $OS"
    exit 1
    ;;
esac

case "$ARCH" in
  x86_64)
    ARCH="x86_64"
    ;;
  aarch64|arm64)
    ARCH="arm64"
    ;;
  *)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

# Get latest release version
VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$VERSION" ]; then
  echo "Failed to get latest version"
  exit 1
fi

echo "Installing cfmon $VERSION for $OS $ARCH..."

# Download URL
ARCHIVE="${BINARY_NAME}_${VERSION}_${OS}_${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$VERSION/$ARCHIVE"
CHECKSUM_URL="https://github.com/$REPO/releases/download/$VERSION/checksums.txt"

# Create temp directory
TMP_DIR=$(mktemp -d)
cd "$TMP_DIR"

# Download archive and checksums
echo "Downloading $ARCHIVE..."
curl -sL "$DOWNLOAD_URL" -o "$ARCHIVE"
curl -sL "$CHECKSUM_URL" -o checksums.txt

# Verify checksum
echo "Verifying checksum..."
if command -v sha256sum >/dev/null 2>&1; then
  grep "$ARCHIVE" checksums.txt | sha256sum -c -
elif command -v shasum >/dev/null 2>&1; then
  grep "$ARCHIVE" checksums.txt | shasum -a 256 -c -
else
  echo "Warning: Neither sha256sum nor shasum found, skipping checksum verification"
fi

# Extract
echo "Extracting..."
tar -xzf "$ARCHIVE"

# Install
echo "Installing to $INSTALL_DIR..."
if [ -w "$INSTALL_DIR" ]; then
  mv "$BINARY_NAME" "$INSTALL_DIR/"
else
  sudo mv "$BINARY_NAME" "$INSTALL_DIR/"
fi

# Cleanup
cd -
rm -rf "$TMP_DIR"

# Verify installation
if command -v "$BINARY_NAME" >/dev/null 2>&1; then
  echo "✓ cfmon installed successfully!"
  echo "Run 'cfmon --help' to get started"
else
  echo "Installation complete, but $BINARY_NAME is not in PATH"
  echo "Add $INSTALL_DIR to your PATH or move the binary to a directory in your PATH"
fi

echo ""
echo "Alternative: Install via Homebrew (if available):"
echo "  brew tap PeterHiroshi/cfmon"
echo "  brew install cfmon"
