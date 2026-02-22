#!/bin/bash
#
# Install script for kvf - https://github.com/oxio/kvf
# Auto-detects OS and architecture and downloads the correct binary
#

set -e

REPO="oxio/kvf"
BINARY_NAME="kvf"
INSTALL_DIR="/usr/local/bin"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Print colored output
info() { echo -e "${GREEN}[INFO]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }

# Detect operating system
detect_os() {
    case "$(uname -s)" in
        Darwin*)    echo "darwin" ;;
        Linux*)     echo "linux" ;;
        CYGWIN*|MINGW*|MSYS*)    echo "windows" ;;
        *)          error "Unsupported OS: $(uname -s)" ;;
    esac
}

# Detect architecture
detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64)  echo "amd64" ;;
        arm64|aarch64) echo "arm64" ;;
        armv7l|armv7)  echo "arm" ;;
        *)             error "Unsupported architecture: $(uname -m)" ;;
    esac
}

# Get latest release version from GitHub API
get_latest_version() {
    if command -v curl &> /dev/null; then
        curl -sL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
    elif command -v wget &> /dev/null; then
        wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
    else
        error "Neither curl nor wget is available. Please install one of them."
    fi
}

# Download file using curl or wget
download_file() {
    local url="$1"
    local output="$2"
    
    if command -v curl &> /dev/null; then
        curl -sL "$url" -o "$output"
    elif command -v wget &> /dev/null; then
        wget -q "$url" -O "$output"
    else
        error "Neither curl nor wget is available. Please install one of them."
    fi
}

# Main installation logic
main() {
    # Detect OS and architecture
    OS=$(detect_os)
    ARCH=$(detect_arch)
    
    # For Windows, use .exe extension
    if [ "$OS" = "windows" ]; then
        BINARY="${BINARY_NAME}-${OS}-${ARCH}.exe"
    else
        BINARY="${BINARY_NAME}-${OS}-${ARCH}"
    fi
    
    info "Detected OS: $OS"
    info "Detected architecture: $ARCH"
    
    # Get latest version
    VERSION=$(get_latest_version)
    if [ -z "$VERSION" ]; then
        error "Could not determine latest version. Please check your internet connection."
    fi
    info "Latest version: $VERSION"
    
    # Construct download URL
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY}"
    
    info "Downloading: $DOWNLOAD_URL"
    
    # Create temporary directory
    TMP_DIR=$(mktemp -d)
    TMP_FILE="${TMP_DIR}/${BINARY_NAME}"
    
    # Download the binary
    download_file "$DOWNLOAD_URL" "$TMP_FILE"
    
    # Check if download was successful
    if [ ! -s "$TMP_FILE" ]; then
        rm -rf "$TMP_DIR"
        error "Download failed or file is empty. Binary might not be available for your platform."
    fi
    
    # Make executable
    chmod +x "$TMP_FILE"
    
    # Determine install directory (use local bin if no sudo access)
    if [ -w "$INSTALL_DIR" ]; then
        DEST="${INSTALL_DIR}/${BINARY_NAME}"
        mv "$TMP_FILE" "$DEST"
    elif command -v sudo &> /dev/null; then
        DEST="${INSTALL_DIR}/${BINARY_NAME}"
        sudo mv "$TMP_FILE" "$DEST"
    else
        # Fallback to user's local bin
        LOCAL_BIN="${HOME}/.local/bin"
        mkdir -p "$LOCAL_BIN"
        DEST="${LOCAL_BIN}/${BINARY_NAME}"
        mv "$TMP_FILE" "$DEST"
        warn "Could not write to ${INSTALL_DIR}. Installed to ${DEST}"
        warn "Make sure ${LOCAL_BIN} is in your PATH."
    fi
    
    # Cleanup
    rm -rf "$TMP_DIR"
    
    info "Successfully installed ${BINARY_NAME} to ${DEST}"
    info "Version: $VERSION"
    
    # Verify installation
    if command -v "$BINARY_NAME" &> /dev/null; then
        info "Verification: $(${BINARY_NAME} --version 2>/dev/null || echo 'installed')"
    else
        warn "${BINARY_NAME} is not in your PATH. Add the install directory to your PATH."
    fi
}

# Run main function
main "$@"
