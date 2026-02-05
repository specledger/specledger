#!/bin/bash
# install.sh - Installation script for SpecLedger CLI
# Supports: macOS, Linux, and requires curl or wget

set -e

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Default variables
VERSION="${VERSION:-latest}"
DOWNLOAD_URL="${DOWNLOAD_URL:-}"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
USE_SUDO="${USE_SUDO:-}"

# Detect OS
detect_os() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
        echo "darwin"
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        echo "linux"
    elif [[ "$OSTYPE" == "msys" || "$OSTYPE" == "win32" ]]; then
        echo "windows"
    else
        echo "unknown"
    fi
}

OS="$(detect_os)"

# Download function
download_file() {
    local url="$1"
    local output="$2"

    if command -v curl > /dev/null 2>&1; then
        if ! curl -fsSL "$url" -o "$output"; then
            echo -e "${RED}Error: Failed to download from $url${NC}" >&2
            exit 1
        fi
    elif command -v wget > /dev/null 2>&1; then
        if ! wget -q "$url" -O "$output"; then
            echo -e "${RED}Error: Failed to download from $url${NC}" >&2
            exit 1
        fi
    else
        echo -e "${RED}Error: Neither curl nor wget is available${NC}" >&2
        exit 1
    fi
}

# Extract OS-specific download URL
get_download_url() {
    local version="$1"
    local arch="${ARCH:-amd64}"

    case "$OS" in
        darwin)
            echo "https://github.com/specledger/specledger/releases/download/${version}/specledger_${version}_darwin_${arch}.tar.gz"
            ;;
        linux)
            echo "https://github.com/specledger/specledger/releases/download/${version}/specledger_${version}_linux_${arch}.tar.gz"
            ;;
        windows)
            echo "https://github.com/specledger/specledger/releases/download/${version}/specledger_${version}_windows_${arch}.zip"
            ;;
        *)
            echo -e "${RED}Error: Unsupported operating system: $OS${NC}" >&2
            exit 1
            ;;
    esac
}

# Create install directory if it doesn't exist
setup_install_dir() {
    if [[ ! -d "$INSTALL_DIR" ]]; then
        echo -e "${YELLOW}Creating install directory: $INSTALL_DIR${NC}"
        mkdir -p "$INSTALL_DIR"
    fi

    # Add to PATH if not already there
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        echo -e "${YELLOW}Adding $INSTALL_DIR to PATH${NC}"

        local shell_rc=""
        if [[ -n "$ZSH_VERSION" ]]; then
            shell_rc="$HOME/.zshrc"
        elif [[ -n "$BASH_VERSION" ]]; then
            shell_rc="$HOME/.bashrc"
        elif [[ -n "$FISH_VERSION" ]]; then
            shell_rc="$HOME/.config/fish/config.fish"
        fi

        if [[ -n "$shell_rc" ]]; then
            if ! grep -q "$INSTALL_DIR" "$shell_rc" 2>/dev/null; then
                echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >> "$shell_rc"
                echo -e "${GREEN}Added $INSTALL_DIR to $shell_rc${NC}"
            fi
        fi
    fi
}

# Install on macOS/Linux
install_on_unix() {
    local version="$1"
    local url="$2"
    local temp_file=$(mktemp)
    local extract_dir=$(mktemp -d)

    echo "Downloading SpecLedger $version for $OS..."

    download_file "$url" "$temp_file"

    echo "Extracting..."

    if [[ "$OS" == "windows" ]]; then
        unzip -q "$temp_file" -d "$extract_dir"
    else
        tar -xzf "$temp_file" -C "$extract_dir"
    fi

    # Find the extracted binary
    local binary_path=""
    if [[ "$OS" == "windows" ]]; then
        binary_path="$extract_dir/specledger_${version}_windows_${ARCH}/sl.exe"
    else
        binary_path="$extract_dir/sl"
    fi

    if [[ ! -f "$binary_path" ]]; then
        echo -e "${RED}Error: Binary not found at $binary_path${NC}" >&2
        rm -rf "$temp_file" "$extract_dir"
        exit 1
    fi

    # Determine if we need sudo
    if [[ -w "$INSTALL_DIR" ]]; then
        cp "$binary_path" "$INSTALL_DIR/sl"
        chmod +x "$INSTALL_DIR/sl"
    else
        if [[ -z "$USE_SUDO" ]]; then
            echo -e "${YELLOW}You may need to run with sudo for system-wide install${NC}"
            USE_SUDO="true"
        fi
        if [[ "$USE_SUDO" == "true" ]]; then
            sudo cp "$binary_path" "/usr/local/bin/sl"
            sudo chmod +x "/usr/local/bin/sl"
        else
            cp "$binary_path" "$INSTALL_DIR/sl"
            chmod +x "$INSTALL_DIR/sl"
        fi
    fi

    echo -e "${GREEN}✓ Installed SpecLedger $version to $INSTALL_DIR/sl${NC}"
    echo -e "${YELLOW}Please ensure $INSTALL_DIR is in your PATH${NC}"
    echo ""

    # Cleanup
    rm -rf "$temp_file" "$extract_dir"
}

# Install on Windows
install_on_windows() {
    local version="$1"
    local url="$2"

    echo "Installing SpecLedger $version on Windows..."

    local temp_file="$HOME/AppData/Local/Temp/specledger-install.zip"

    download_file "$url" "$temp_file"

    # Install to Program Files if possible, else AppData
    if [[ -d "C:/Program Files/SpecLedger" ]]; then
        local target_dir="C:/Program Files/SpecLedger"
    else
        local target_dir="$HOME/AppData/Local/SpecLedger"
    fi

    mkdir -p "$target_dir"
    unzip -q "$temp_file" -d "$target_dir"

    # Copy sl.exe to directory
    cp "$target_dir/specledger_${version}_windows_${ARCH}/sl.exe" "$target_dir/sl.exe"
    rm "$temp_file"

    echo -e "${GREEN}✓ Installed SpecLedger $version to $target_dir/sl.exe${NC}"
    echo -e "${YELLOW}Please add $target_dir to your PATH${NC}"
    echo ""
}

# Verify installation
verify_installation() {
    if command -v sl > /dev/null 2>&1; then
        echo "SpecLedger version:"
        sl version || true
        return 0
    else
        echo -e "${RED}Error: SpecLedger not found in PATH${NC}" >&2
        echo "Please add $INSTALL_DIR to your PATH and try again."
        return 1
    fi
}

# Main installation flow
main() {
    echo "SpecLedger Installation Script"
    echo "=============================="
    echo ""

    # Get download URL
    if [[ -z "$DOWNLOAD_URL" ]]; then
        DOWNLOAD_URL=$(get_download_url "$VERSION")
    fi

    echo "Installing SpecLedger $VERSION"
    echo "Platform: $OS"
    echo "Architecture: ${ARCH:-amd64}"
    echo "Install Directory: $INSTALL_DIR"
    echo ""

    # Setup install directory
    setup_install_dir
    echo ""

    # Install based on OS
    if [[ "$OS" == "windows" ]]; then
        install_on_windows "$VERSION" "$DOWNLOAD_URL"
    else
        install_on_unix "$VERSION" "$DOWNLOAD_URL"
    fi

    echo ""
    echo "Installation complete!"

    # Verify
    if verify_installation; then
        echo -e "${GREEN}✓ Installation verified successfully${NC}"
    fi
}

# Run main function
main "$@"
