# Quickstart: Testing Release Delivery

This guide covers how to test the release delivery system for SpecLedger on macOS.

## Prerequisites

- Go 1.24+ installed
- git installed
- GitHub CLI (gh) - optional, for creating releases
- Docker - optional, for testing Linux builds later

## Local Development Setup

### 1. Build the Binary

```bash
# Build for your current platform
go build -o bin/sl cmd/main.go

# Build for both macOS architectures
GOARCH=amd64 go build -o bin/sl-amd64 cmd/main.go
GOARCH=arm64 go build -o bin/sl-arm64 cmd/main.go

# Test the binary
./bin/sl --help
./bin/sl --version  # After implementing version flag
```

### 2. Test Go Install Method

```bash
# Install from local module
go install ./cmd/sl@latest

# Verify installation
sl --help
sl --version
```

### 3. Create Test Archives (Simulating GoReleaser)

```bash
# Create darwin_amd64 archive
mkdir -p dist/darwin_amd64
cp bin/sl-amd64 dist/darwin_amd64/sl
chmod +x dist/darwin_amd64/sl
cd dist/darwin_amd64 && tar -czf ../../specledger_test_darwin_amd64.tar.gz sl && cd ../../..

# Create darwin_arm64 archive
mkdir -p dist/darwin_arm64
cp bin/sl-arm64 dist/darwin_arm64/sl
chmod +x dist/darwin_arm64/sl
cd dist/darwin_arm64 && tar -czf ../../specledger_test_darwin_arm64.tar.gz sl && cd ../../..

# Create checksums
cd dist
shasum -a 256 *.tar.gz > checksums.txt
cat checksums.txt
cd ..
```

### 4. Test Install Script Locally

```bash
# Set VERSION to point to local file (modify script for testing)
# Or test by copying script and binary to test directory

mkdir -p /tmp/sl-test
cp scripts/install.sh /tmp/sl-test/
cp specledger_test_darwin_amd64.tar.gz /tmp/sl-test/

# Edit install.sh to use local file, or serve via local HTTP server
python3 -m http.server 8080 --directory /tmp/sl-test

# In another terminal, test install (adjust URL as needed)
VERSION=test DOWNLOAD_URL="http://localhost:8080/specledger_test_darwin_amd64.tar.gz" bash /tmp/sl-test/install.sh
```

## GoReleaser Dry-Run

### 1. Install GoReleaser

```bash
# macOS via Homebrew
brew install goreleaser

# Or via Go
go install github.com/goreleaser/goreleaser/v2@latest
```

### 2. Run Dry-Run

```bash
# Test release without publishing
goreleaser release --snapshot --clean

# Check the output
ls -la dist/
```

Expected output:
```
dist/
├── checksums.txt
├── specledger_1.0.0_darwin_amd64.tar.gz
└── specledger_1.0.0_darwin_arm64.tar.gz
```

### 3. Verify Archives

```bash
# Check archive contents
tar -tzf dist/specledger_1.0.0_darwin_amd64.tar.gz
# Should show: sl

# Extract and test
mkdir -p test-extract
tar -xzf dist/specledger_1.0.0_darwin_amd64.tar.gz -C test-extract
./test-extract/sl --version
```

### 4. Verify Checksums

```bash
cd dist
shasum -a 256 -c checksums.txt
```

## Creating a Test Release

### Option 1: Using Git Tag (Triggers GitHub Actions)

```bash
# Create and push a test tag (e.g., v1.0.3-test)
git tag v1.0.3-test
git push origin v1.0.3-test

# Monitor the GitHub Actions workflow
# Visit: https://github.com/specledger/specledger/actions
```

### Option 2: Manual GitHub Release

```bash
# If gh CLI is installed
gh release create v1.0.3-test \
  --title "Test Release v1.0.3" \
  --notes "Testing release delivery"
```

Then manually upload the binaries from `dist/`.

## Testing Homebrew Installation

### 1. Create Homebrew Tap Repository

```bash
# Using gh CLI
gh repo create specledger/homebrew-specledger --public --description "Homebrew tap for SpecLedger CLI"

# Or create manually on GitHub
```

### 2. Initialize Tap Repository

```bash
# Clone the tap repository
git clone git@github.com:specledger/homebrew-specledger.git
cd homebrew-specledger

# Create README
cat > README.md << 'EOF'
# homebrew-specledger

Homebrew tap for [SpecLedger](https://github.com/specledger/specledger).

## Installation

```bash
brew tap specledger/homebrew-specledger
brew install specledger
```
EOF

git add README.md
git commit -m "Initial commit"
git push origin main
```

### 3. Test GoReleaser Homebrew Upload

```bash
# Run GoReleaser with skip_upload: false
# Update .goreleaser.yaml to remove skip_upload

goreleaser release --snapshot --clean
```

Check if a formula file is created and can be uploaded to the tap.

### 4. Test Installation

```bash
# Tap the repository (from local path for testing)
brew tap specledger/homebrew-specledger

# Install
brew install specledger

# Verify
sl --version

# Upgrade
brew upgrade specledger
```

## Testing on Apple Silicon

If you have access to Apple Silicon hardware:

```bash
# Verify architecture
uname -m
# Should output: arm64

# Test install script
ARCH=arm64 bash scripts/install.sh

# Verify binary architecture
file $(which sl)
# Should output: ARM64 executable
```

## Troubleshooting

### Install Script Fails

```bash
# Enable debug mode
bash -x scripts/install.sh

# Check for ARCH variable
echo "ARCH: ${ARCH:-amd64}"
```

### GoReleaser Fails

```bash
# Check Go version
go version

# Verify .goreleaser.yaml syntax
goreleaser check

# Run with verbose output
goreleaser release --snapshot --clean --verbose
```

### Checksum Verification Fails

```bash
# Manually verify checksum
shasum -a 256 dist/specledger_*.tar.gz

# Compare with checksums.txt
cat dist/checksums.txt
```

## Cleanup

```bash
# Remove test installations
rm -rf ~/.local/bin/sl
rm -rf /usr/local/bin/sl  # If installed with sudo

# Untap homebrew (if testing)
brew untap specledger/homebrew-specledger

# Remove test tags (local)
git tag -d v1.0.3-test

# Remove test tags (remote)
gh release delete v1.0.3-test --yes
git push origin :refs/tags/v1.0.3-test
```
