#!/bin/bash

# npm-console Installation Script
# Downloads and installs the latest version of npm-console

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO="npm-console/npm-console"
BINARY_NAME="npm-console"
INSTALL_DIR="/usr/local/bin"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    i386|i686)
        ARCH="386"
        ;;
    *)
        echo -e "${RED}❌ Unsupported architecture: $ARCH${NC}"
        exit 1
        ;;
esac

case $OS in
    linux|darwin)
        ;;
    *)
        echo -e "${RED}❌ Unsupported operating system: $OS${NC}"
        exit 1
        ;;
esac

echo -e "${BLUE}🚀 npm-console Installation Script${NC}"
echo -e "${BLUE}===================================${NC}"
echo ""
echo -e "${YELLOW}Detected platform: ${OS}/${ARCH}${NC}"

# Check if running as root for system-wide installation
if [ "$EUID" -ne 0 ] && [ "$INSTALL_DIR" = "/usr/local/bin" ]; then
    echo -e "${YELLOW}⚠️  Installing to user directory instead of system-wide${NC}"
    INSTALL_DIR="$HOME/.local/bin"
    mkdir -p "$INSTALL_DIR"
    
    # Add to PATH if not already there
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        echo -e "${YELLOW}📝 Adding $INSTALL_DIR to PATH${NC}"
        echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$HOME/.bashrc"
        echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$HOME/.zshrc" 2>/dev/null || true
        echo -e "${YELLOW}⚠️  Please restart your shell or run: source ~/.bashrc${NC}"
    fi
fi

# Get latest release information
echo -e "${YELLOW}🔍 Fetching latest release information...${NC}"
LATEST_RELEASE=$(curl -s "https://api.github.com/repos/$REPO/releases/latest")

if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Failed to fetch release information${NC}"
    exit 1
fi

VERSION=$(echo "$LATEST_RELEASE" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
DOWNLOAD_URL=$(echo "$LATEST_RELEASE" | grep "browser_download_url.*${BINARY_NAME}-${OS}-${ARCH}" | cut -d '"' -f 4)

if [ -z "$VERSION" ] || [ -z "$DOWNLOAD_URL" ]; then
    echo -e "${RED}❌ Could not find release for ${OS}/${ARCH}${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Found version: $VERSION${NC}"
echo -e "${YELLOW}📥 Download URL: $DOWNLOAD_URL${NC}"

# Create temporary directory
TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

# Download binary
echo -e "${YELLOW}⬇️  Downloading npm-console...${NC}"
curl -L -o "$TMP_DIR/$BINARY_NAME" "$DOWNLOAD_URL"

if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Failed to download npm-console${NC}"
    exit 1
fi

# Make executable
chmod +x "$TMP_DIR/$BINARY_NAME"

# Verify binary works
echo -e "${YELLOW}🔍 Verifying binary...${NC}"
if ! "$TMP_DIR/$BINARY_NAME" version >/dev/null 2>&1; then
    echo -e "${RED}❌ Downloaded binary is not working${NC}"
    exit 1
fi

# Install binary
echo -e "${YELLOW}📦 Installing to $INSTALL_DIR...${NC}"
if [ -w "$INSTALL_DIR" ]; then
    cp "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/"
else
    sudo cp "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/"
fi

# Verify installation
if command -v "$BINARY_NAME" >/dev/null 2>&1; then
    INSTALLED_VERSION=$($BINARY_NAME version --short)
    echo -e "${GREEN}✅ npm-console $INSTALLED_VERSION installed successfully!${NC}"
else
    echo -e "${YELLOW}⚠️  npm-console installed but not in PATH${NC}"
    echo -e "${YELLOW}   Binary location: $INSTALL_DIR/$BINARY_NAME${NC}"
fi

echo ""
echo -e "${BLUE}🎉 Installation completed!${NC}"
echo ""
echo -e "${YELLOW}Quick start:${NC}"
echo -e "  ${BINARY_NAME} --help          # Show help"
echo -e "  ${BINARY_NAME} cache list      # List cache information"
echo -e "  ${BINARY_NAME} packages list   # List installed packages"
echo -e "  ${BINARY_NAME} web             # Start web interface"
echo ""
echo -e "${BLUE}For more information, visit: https://github.com/$REPO${NC}"
