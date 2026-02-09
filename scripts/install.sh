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

# Detect architecture
detect_arch() {
    local arch=$(uname -m)
    case "$arch" in
        x86_64)
            echo "amd64"
            ;;
        aarch64|arm64)
            echo "arm64"
            ;;
        armv7l)
            echo "arm"
            ;;
        i386|i686)
            echo "386"
            ;;
        *)
            echo "unknown"
            ;;
    esac
}

OS="$(detect_os)"
ARCH="${ARCH:-$(detect_arch)}"

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
    local arch="$2"

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

# Get checksum URL
get_checksum_url() {
    local version="$1"
    echo "https://github.com/specledger/specledger/releases/download/${version}/checksums.txt"
}

# Verify checksum
verify_checksum() {
    local archive_file="$1"
    local checksum_file="$2"
    local archive_name=$(basename "$archive_file")

    if [[ ! -f "$checksum_file" ]]; then
        echo -e "${YELLOW}Warning: Checksum file not found, skipping verification${NC}" >&2
        return 0
    fi

    echo "Verifying checksum..."

    # Extract the expected checksum for our archive
    local expected_checksum=$(grep "$archive_name" "$checksum_file" | awk '{print $1}')

    if [[ -z "$expected_checksum" ]]; then
        echo -e "${YELLOW}Warning: No checksum found for $archive_name, skipping verification${NC}" >&2
        return 0
    fi

    # Calculate actual checksum
    if command -v shasum > /dev/null 2>&1; then
        local actual_checksum=$(shasum -a 256 "$archive_file" | awk '{print $1}')
        if [[ "$actual_checksum" == "$expected_checksum" ]]; then
            echo -e "${GREEN}✓ Checksum verified${NC}"
            return 0
        else
            echo -e "${RED}Error: Checksum mismatch!${NC}" >&2
            echo "Expected: $expected_checksum" >&2
            echo "Actual:   $actual_checksum" >&2
            rm -f "$archive_file"
            exit 1
        fi
    else
        echo -e "${YELLOW}shasum not available, skipping checksum verification${NC}" >&2
        return 0
    fi
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
    local checksum_file=$(mktemp)
    local extract_dir=$(mktemp -d)

    echo "Downloading SpecLedger $version for $OS ($ARCH)..."

    # Download archive
    download_file "$url" "$temp_file"

    # Download checksums
    download_file "$(get_checksum_url "$version")" "$checksum_file"

    # Verify checksum
    verify_checksum "$temp_file" "$checksum_file"

    echo "Extracting..."
    tar -xzf "$temp_file" -C "$extract_dir"

    # Find the extracted binary - GoReleaser puts 'sl' at the root
    local binary_path="$extract_dir/sl"

    if [[ ! -f "$binary_path" ]]; then
        echo -e "${RED}Error: Binary not found at $binary_path${NC}" >&2
        echo "Looking for files in extract directory:" >&2
        ls -la "$extract_dir" >&2
        rm -rf "$temp_file" "$checksum_file" "$extract_dir"
        exit 1
    fi

    # Install binary
    if [[ -w "$INSTALL_DIR" ]]; then
        cp "$binary_path" "$INSTALL_DIR/sl"
        chmod +x "$INSTALL_DIR/sl"
    else
        if [[ -z "$USE_SUDO" ]]; then
            echo -e "${YELLOW}Install directory not writable. You may need to run with sudo for system-wide install${NC}"
        fi

        # Try with sudo if available and directory is not writable
        if command -v sudo > /dev/null 2>&1 && [[ "$USE_SUDO" != "false" ]]; then
            if [[ "$INSTALL_DIR" == "/usr/local/bin" ]]; then
                sudo cp "$binary_path" "$INSTALL_DIR/sl"
                sudo chmod +x "$INSTALL_DIR/sl"
                echo -e "${GREEN}✓ Installed SpecLedger $version to $INSTALL_DIR/sl${NC}"
            else
                echo -e "${RED}Error: Cannot write to $INSTALL_DIR and sudo install only supports /usr/local/bin${NC}" >&2
                echo "Please set INSTALL_DIR to a writable location or run with sudo." >&2
                rm -rf "$temp_file" "$checksum_file" "$extract_dir"
                exit 1
            fi
        else
            echo -e "${RED}Error: Cannot write to $INSTALL_DIR${NC}" >&2
            rm -rf "$temp_file" "$checksum_file" "$extract_dir"
            exit 1
        fi
    fi

    # Cleanup
    rm -rf "$temp_file" "$checksum_file" "$extract_dir"
}

# Install on Windows
install_on_windows() {
    local version="$1"
    local url="$2"

    echo "Installing SpecLedger $version on Windows..."

    local temp_file="$HOME/AppData/Local/Temp/specledger-install.zip"
    local extract_dir="$HOME/AppData/Local/Temp/specledger-install"

    download_file "$url" "$temp_file"

    mkdir -p "$extract_dir"
    unzip -q "$temp_file" -d "$extract_dir"

    # Find the extracted binary
    local binary_path="$extract_dir/sl.exe"

    if [[ ! -f "$binary_path" ]]; then
        echo -e "${RED}Error: Binary not found at $binary_path${NC}" >&2
        rm -f "$temp_file"
        exit 1
    fi

    # Install to Program Files if possible, else AppData
    if [[ -w "C:/Program Files" ]]; then
        local target_dir="C:/Program Files/SpecLedger"
    else
        local target_dir="$HOME/AppData/Local/SpecLedger"
    fi

    mkdir -p "$target_dir"
    cp "$binary_path" "$target_dir/sl.exe"
    rm -f "$temp_file"
    rm -rf "$extract_dir"

    echo -e "${GREEN}✓ Installed SpecLedger $version to $target_dir/sl.exe${NC}"
    echo -e "${YELLOW}Please add $target_dir to your PATH${NC}"
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

    # Auto-detect architecture if not set
    if [[ -z "$ARCH" || "$ARCH" == "unknown" ]]; then
        ARCH=$(detect_arch)
        if [[ "$ARCH" == "unknown" ]]; then
            echo -e "${RED}Error: Unable to detect system architecture${NC}" >&2
            exit 1
        fi
    fi

    # Get download URL
    if [[ -z "$DOWNLOAD_URL" ]]; then
        DOWNLOAD_URL=$(get_download_url "$VERSION" "$ARCH")
    fi

    echo "Installing SpecLedger $VERSION"
    echo "Platform: $OS"
    echo "Architecture: $ARCH"
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
        echo ""
        echo "To get started, run:"
        echo "  sl --help"
    else
        echo ""
        echo -e "${YELLOW}Note: $INSTALL_DIR may not be in your PATH yet.${NC}"
        echo "Start a new shell or add $INSTALL_DIR to your PATH."
    fi
}

# Run main function
main "$@"
