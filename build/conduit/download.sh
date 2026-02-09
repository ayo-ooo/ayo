#!/bin/bash
# Downloads Conduit binary for the current or specified platform
# Usage: ./download.sh [--platform linux-amd64|linux-arm64|darwin-amd64|darwin-arm64] [--output DIR]

set -e

# Conduit version to download
CONDUIT_VERSION="0.8.0"
CONDUIT_REPO="https://gitlab.com/famedly/conduit/-/releases"

# Default output directory
OUTPUT_DIR="${AYO_DATA_DIR:-$HOME/.local/share/ayo}/bin"

# Detect platform
detect_platform() {
    local os=$(uname -s | tr '[:upper:]' '[:lower:]')
    local arch=$(uname -m)
    
    case "$os" in
        linux) os="linux" ;;
        darwin) os="darwin" ;;
        *) echo "Unsupported OS: $os" >&2; exit 1 ;;
    esac
    
    case "$arch" in
        x86_64|amd64) arch="amd64" ;;
        aarch64|arm64) arch="arm64" ;;
        *) echo "Unsupported architecture: $arch" >&2; exit 1 ;;
    esac
    
    echo "${os}-${arch}"
}

# Parse arguments
PLATFORM=""
while [[ $# -gt 0 ]]; do
    case $1 in
        --platform)
            PLATFORM="$2"
            shift 2
            ;;
        --output)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        *)
            echo "Unknown option: $1" >&2
            exit 1
            ;;
    esac
done

if [ -z "$PLATFORM" ]; then
    PLATFORM=$(detect_platform)
fi

echo "Platform: $PLATFORM"
echo "Output: $OUTPUT_DIR"

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Build download URL based on platform
# Note: Conduit releases may have different naming conventions
# This is a placeholder - adjust based on actual Conduit release URLs
case "$PLATFORM" in
    linux-amd64)
        BINARY_NAME="conduit-x86_64-unknown-linux-gnu"
        ;;
    linux-arm64)
        BINARY_NAME="conduit-aarch64-unknown-linux-gnu"
        ;;
    darwin-amd64)
        BINARY_NAME="conduit-x86_64-apple-darwin"
        ;;
    darwin-arm64)
        BINARY_NAME="conduit-aarch64-apple-darwin"
        ;;
    *)
        echo "Unsupported platform: $PLATFORM" >&2
        exit 1
        ;;
esac

# For now, we'll use a mock download since Conduit's exact release URLs vary
# In production, this would download from the actual release
DOWNLOAD_URL="${CONDUIT_REPO}/v${CONDUIT_VERSION}/downloads/${BINARY_NAME}"

TEMP_FILE=$(mktemp)
trap "rm -f $TEMP_FILE" EXIT

echo "Downloading Conduit v${CONDUIT_VERSION} for ${PLATFORM}..."

# Check if curl or wget is available
if command -v curl &> /dev/null; then
    if ! curl -fsSL "$DOWNLOAD_URL" -o "$TEMP_FILE" 2>/dev/null; then
        echo "Note: Could not download from official source."
        echo "For development, please manually install Conduit or use cargo:"
        echo "  cargo install conduit"
        echo "  # or download from https://gitlab.com/famedly/conduit/-/releases"
        exit 1
    fi
elif command -v wget &> /dev/null; then
    if ! wget -q "$DOWNLOAD_URL" -O "$TEMP_FILE" 2>/dev/null; then
        echo "Note: Could not download from official source."
        echo "For development, please manually install Conduit or use cargo:"
        echo "  cargo install conduit"
        exit 1
    fi
else
    echo "Neither curl nor wget found. Please install one of them." >&2
    exit 1
fi

# Verify the download (placeholder for checksum verification)
if [ ! -s "$TEMP_FILE" ]; then
    echo "Download failed or file is empty" >&2
    exit 1
fi

# Move to final location
mv "$TEMP_FILE" "$OUTPUT_DIR/conduit"
chmod +x "$OUTPUT_DIR/conduit"

echo "Conduit installed to $OUTPUT_DIR/conduit"

# Verify it works
if "$OUTPUT_DIR/conduit" --version &> /dev/null; then
    echo "Verification: $("$OUTPUT_DIR/conduit" --version)"
else
    echo "Warning: Could not verify Conduit installation"
fi
