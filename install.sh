#!/bin/bash
#
# Install script for kvf - https://github.com/oxio/kvf
# Auto-detects OS and architecture and downloads the correct binary
#
# Usage:
#   curl -sL https://raw.githubusercontent.com/oxio/kvf/main/install.sh | bash
#   curl -sL https://raw.githubusercontent.com/oxio/kvf/main/install.sh | bash -s -- v1.0.0
#   ./install.sh [VERSION]
#

set -e

REPO="oxio/kvf"
BINARY_NAME="kvf"
INSTALL_DIR="/usr/local/bin"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print colored output
info() { echo -e "${GREEN}[INFO]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }
step() { echo -e "${BLUE}[...]${NC} $1"; }

# Show usage information
usage() {
    echo "Usage: $0 [OPTIONS] [VERSION]"
    echo ""
    echo "Install kvf - Key Value File CLI tool"
    echo ""
    echo "Arguments:"
    echo "  VERSION       Specific version to install (e.g., v1.0.0)"
    echo "                If not specified, installs the latest version"
    echo ""
    echo "Options:"
    echo "  -h, --help    Show this help message"
    echo "  -f, --force   Force reinstall even if already up-to-date"
    echo ""
    echo "Examples:"
    echo "  $0                    # Install latest version"
    echo "  $0 v1.0.0             # Install specific version"
    echo "  $0 --force            # Force reinstall latest"
    echo "  $0 v1.0.0 --force     # Force reinstall v1.0.0"
    echo ""
    echo "One-liner usage:"
    echo "  curl -sL https://raw.githubusercontent.com/oxio/kvf/main/install.sh | bash"
    echo "  curl -sL https://raw.githubusercontent.com/oxio/kvf/main/install.sh | bash -s -- v1.0.0"
}

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
    local version
    if command -v curl &> /dev/null; then
        version=$(curl -sL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    elif command -v wget &> /dev/null; then
        version=$(wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    else
        error "Neither curl nor wget is available. Please install one of them."
    fi
    echo "$version"
}

# Validate that a version exists
validate_version() {
    local version="$1"
    local release_url="https://github.com/${REPO}/releases/tag/${version}"
    
    if command -v curl &> /dev/null; then
        if ! curl -sL -o /dev/null -w "%{http_code}" "$release_url" | grep -q "200"; then
            return 1
        fi
    elif command -v wget &> /dev/null; then
        if ! wget -q --spider "$release_url"; then
            return 1
        fi
    fi
    return 0
}

# Get currently installed version
get_installed_version() {
    local binary_path="$1"
    if [ -x "$binary_path" ]; then
        # Try to get version from the binary
        # The binary may output version via --version or version subcommand
        local version
        version=$("$binary_path" --version 2>/dev/null | head -1 | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)
        if [ -z "$version" ]; then
            version=$("$binary_path" version 2>/dev/null | head -1 | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)
        fi
        if [ -n "$version" ]; then
            echo "v${version}"
        fi
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

# Find the existing binary location
find_existing_binary() {
    if command -v "$BINARY_NAME" &> /dev/null; then
        command -v "$BINARY_NAME"
    fi
}

# Determine where to install the binary
determine_install_dir() {
    local existing_binary="$1"
    
    # If binary exists, use its directory
    if [ -n "$existing_binary" ]; then
        dirname "$existing_binary"
        return
    fi
    
    # Otherwise, try default locations
    if [ -w "$INSTALL_DIR" ]; then
        echo "$INSTALL_DIR"
    elif [ -w "${HOME}/.local/bin" ] || mkdir -p "${HOME}/.local/bin" 2>/dev/null; then
        echo "${HOME}/.local/bin"
    else
        echo "$INSTALL_DIR"  # Will need sudo
    fi
}

# Parse command line arguments
parse_args() {
    FORCE=false
    REQUESTED_VERSION=""
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                usage
                exit 0
                ;;
            -f|--force)
                FORCE=true
                shift
                ;;
            v*)
                # Version argument (starts with 'v')
                REQUESTED_VERSION="$1"
                shift
                ;;
            *)
                # Unknown argument - could be version without 'v' prefix
                if [[ $1 =~ ^[0-9] ]]; then
                    REQUESTED_VERSION="v$1"
                else
                    warn "Unknown argument: $1"
                    usage
                    exit 1
                fi
                shift
                ;;
        esac
    done
}

# Main installation logic
main() {
    parse_args "$@"
    
    # Detect OS and architecture
    OS=$(detect_os)
    ARCH=$(detect_arch)
    
    # For Windows, use .exe extension
    if [ "$OS" = "windows" ]; then
        BINARY="${BINARY_NAME}-${OS}-${ARCH}.exe"
    else
        BINARY="${BINARY_NAME}-${OS}-${ARCH}"
    fi
    
    step "Detected OS: $OS"
    step "Detected architecture: $ARCH"
    
    # Find existing installation
    EXISTING_BINARY=$(find_existing_binary)
    INSTALLED_VERSION=$(get_installed_version "$EXISTING_BINARY")
    
    # Determine version to install
    if [ -n "$REQUESTED_VERSION" ]; then
        TARGET_VERSION="$REQUESTED_VERSION"
        step "Validating version ${TARGET_VERSION}..."
        if ! validate_version "$TARGET_VERSION"; then
            error "Version ${TARGET_VERSION} not found. Check available versions at:"
            echo "  https://github.com/${REPO}/releases"
        fi
    else
        step "Checking for latest version..."
        TARGET_VERSION=$(get_latest_version)
        if [ -z "$TARGET_VERSION" ]; then
            error "Could not determine latest version. Please check your internet connection."
        fi
    fi
    
    # Check if we need to install
    if [ "$FORCE" = false ] && [ -n "$INSTALLED_VERSION" ]; then
        if [ "$INSTALLED_VERSION" = "$TARGET_VERSION" ]; then
            info "Already up to date: ${BINARY_NAME} ${INSTALLED_VERSION}"
            echo ""
            echo "Current installation: ${EXISTING_BINARY}"
            echo "To reinstall, run: $0 --force"
            exit 0
        else
            info "Update available: ${INSTALLED_VERSION} → ${TARGET_VERSION}"
        fi
    elif [ "$FORCE" = true ] && [ -n "$INSTALLED_VERSION" ]; then
        if [ "$INSTALLED_VERSION" = "$TARGET_VERSION" ]; then
            info "Force reinstalling ${BINARY_NAME} ${TARGET_VERSION}"
        else
            info "Downgrading: ${INSTALLED_VERSION} → ${TARGET_VERSION}"
        fi
    else
        info "Installing ${BINARY_NAME} ${TARGET_VERSION}"
    fi
    
    # Construct download URL
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${TARGET_VERSION}/${BINARY}"
    
    step "Downloading: $DOWNLOAD_URL"
    
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
    
    # Determine install directory
    INSTALL_DEST=$(determine_install_dir "$EXISTING_BINARY")
    DEST="${INSTALL_DEST}/${BINARY_NAME}"
    
    # Install the binary
    if [ -w "$INSTALL_DEST" ]; then
        mv "$TMP_FILE" "$DEST"
    elif command -v sudo &> /dev/null; then
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
    
    # Success message
    echo ""
    if [ -n "$INSTALLED_VERSION" ]; then
        if [ "$INSTALLED_VERSION" = "$TARGET_VERSION" ]; then
            info "Successfully reinstalled ${BINARY_NAME} ${TARGET_VERSION}"
        else
            info "Successfully updated ${BINARY_NAME} from ${INSTALLED_VERSION} to ${TARGET_VERSION}"
        fi
    else
        info "Successfully installed ${BINARY_NAME} ${TARGET_VERSION}"
    fi
    echo ""
    echo "Binary location: ${DEST}"
    
    # Verify installation
    if command -v "$BINARY_NAME" &> /dev/null; then
        INSTALLED_VERSION=$(get_installed_version "$DEST")
        if [ -n "$INSTALLED_VERSION" ]; then
            echo "Version: ${INSTALLED_VERSION}"
        fi
    else
        warn "${BINARY_NAME} is not in your PATH. Add ${INSTALL_DEST} to your PATH."
    fi
}

# Run main function
main "$@"
