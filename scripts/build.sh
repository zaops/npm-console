#!/bin/bash

# npm-console Build Script
# Builds binaries for multiple platforms

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
APP_NAME="npm-console"
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build information
LDFLAGS="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}"

# Output directory
OUTPUT_DIR="dist"
BIN_DIR="${OUTPUT_DIR}/bin"

# Supported platforms
PLATFORMS=(
    "windows/amd64"
    "windows/386"
    "linux/amd64"
    "linux/386"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
)

echo -e "${BLUE}🚀 Building ${APP_NAME} v${VERSION}${NC}"
echo -e "${BLUE}Build Time: ${BUILD_TIME}${NC}"
echo -e "${BLUE}Git Commit: ${GIT_COMMIT}${NC}"
echo ""

# Clean previous builds
echo -e "${YELLOW}🧹 Cleaning previous builds...${NC}"
rm -rf "${OUTPUT_DIR}"
mkdir -p "${BIN_DIR}"

# Build for each platform
for platform in "${PLATFORMS[@]}"; do
    IFS='/' read -r -a platform_split <<< "$platform"
    GOOS="${platform_split[0]}"
    GOARCH="${platform_split[1]}"
    
    output_name="${APP_NAME}"
    if [ "$GOOS" = "windows" ]; then
        output_name="${output_name}.exe"
    fi
    
    output_path="${BIN_DIR}/${APP_NAME}-${GOOS}-${GOARCH}"
    if [ "$GOOS" = "windows" ]; then
        output_path="${output_path}.exe"
    fi
    
    echo -e "${YELLOW}📦 Building for ${GOOS}/${GOARCH}...${NC}"
    
    env GOOS="$GOOS" GOARCH="$GOARCH" go build \
        -ldflags="${LDFLAGS}" \
        -o "$output_path" \
        .
    
    if [ $? -eq 0 ]; then
        file_size=$(du -h "$output_path" | cut -f1)
        echo -e "${GREEN}✅ Built ${output_path} (${file_size})${NC}"
    else
        echo -e "${RED}❌ Failed to build for ${GOOS}/${GOARCH}${NC}"
        exit 1
    fi
done

echo ""
echo -e "${GREEN}🎉 Build completed successfully!${NC}"
echo -e "${BLUE}📁 Binaries are available in: ${BIN_DIR}${NC}"

# List all built binaries
echo ""
echo -e "${BLUE}📋 Built binaries:${NC}"
ls -lh "${BIN_DIR}/"

# Create checksums
echo ""
echo -e "${YELLOW}🔐 Generating checksums...${NC}"
cd "${BIN_DIR}"
sha256sum * > checksums.txt
echo -e "${GREEN}✅ Checksums saved to checksums.txt${NC}"

echo ""
echo -e "${GREEN}🚀 Build process completed!${NC}"
