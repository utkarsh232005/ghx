#!/bin/sh
set -e

# Configuration
OWNER="KDM-cli"
REPO="ghx"
BINARY_NAME="ghx"

# Find OS and Architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64|amd64)
        ARCH="x86_64"
        ;;
    arm64|aarch64)
        ARCH="arm64"
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

case "$OS" in
    darwin)
        OS="Darwin"
        ;;
    linux)
        OS="Linux"
        ;;
    *)
        echo "Unsupported operating system: $OS"
        exit 1
        ;;
esac

# Fetch latest release tag name from GitHub API
echo "Finding the latest release..."
LATEST_TAG=$(curl -s "https://api.github.com/repos/$OWNER/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_TAG" ]; then
    echo "Error: Could not retrieve latest release tag. Make sure a release tag has been pushed."
    exit 1
fi

echo "Latest release is $LATEST_TAG"

# Construct download URL (matching GoReleaser naming format)
TARBALL="${BINARY_NAME}_${LATEST_TAG#v}_${OS}_${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/$OWNER/$REPO/releases/download/$LATEST_TAG/$TARBALL"

# Create a temporary directory for downloading and extracting
TMP_DIR=$(mktemp -d)
clean_up() {
    rm -rf "$TMP_DIR"
}
trap clean_up EXIT

echo "Downloading $DOWNLOAD_URL..."
if ! curl -sSfL "$DOWNLOAD_URL" -o "$TMP_DIR/$TARBALL"; then
    echo "Error: Failed to download release asset."
    echo "Check if the release asset exists at: https://github.com/$OWNER/$REPO/releases/tag/$LATEST_TAG"
    exit 1
fi

echo "Extracting..."
tar -xzf "$TMP_DIR/$TARBALL" -C "$TMP_DIR"

# Determine installation path
INSTALL_DIR="/usr/local/bin"
if [ ! -w "$INSTALL_DIR" ]; then
    # Fallback to local user bin if /usr/local/bin is not writable without sudo
    INSTALL_DIR="$HOME/.local/bin"
    mkdir -p "$INSTALL_DIR"
    echo "Warning: $INSTALL_DIR is not writable. Installing to $INSTALL_DIR instead."
    echo "Make sure $INSTALL_DIR is in your PATH."
fi

echo "Installing to $INSTALL_DIR..."
mv "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
chmod +x "$INSTALL_DIR/$BINARY_NAME"

echo "🛸 $BINARY_NAME was successfully installed to $INSTALL_DIR/$BINARY_NAME"
