package version

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// SelfUpdate downloads and installs the latest version from GitHub Releases.
func SelfUpdate(ctx context.Context) error {
	// Get latest version info
	info := CheckLatestVersion(ctx)
	if info.Error != "" {
		return fmt.Errorf("failed to check latest version: %s", info.Error)
	}

	if !info.UpdateAvailable && GetVersion() != "dev" {
		fmt.Println("Already up to date!")
		return nil
	}

	fmt.Printf("Updating from %s to %s...\n", GetVersion(), info.LatestVersion)

	// Determine download URL based on platform
	downloadURL, err := getDownloadURL(info.LatestVersion)
	if err != nil {
		return err
	}

	fmt.Printf("Downloading from %s\n", downloadURL)

	// Download the binary
	tmpFile, err := downloadFile(ctx, downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer os.Remove(tmpFile)

	// Make it executable
	if err := os.Chmod(tmpFile, 0755); err != nil {
		return fmt.Errorf("failed to make binary executable: %w", err)
	}

	// Get current binary path
	currentBinary, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current binary path: %w", err)
	}

	// Resolve symlinks
	realBinary, err := filepath.EvalSymlinks(currentBinary)
	if err != nil {
		realBinary = currentBinary
	}

	// Replace the binary
	if err := replaceBinary(tmpFile, realBinary); err != nil {
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	fmt.Printf("Successfully updated to %s!\n", info.LatestVersion)
	return nil
}

// getDownloadURL returns the download URL for the current platform.
func getDownloadURL(version string) (string, error) {
	// Map GOOS to release asset naming
	osName := runtime.GOOS
	arch := runtime.GOARCH

	// Handle architecture naming
	if arch == "amd64" {
		arch = "x86_64"
	} else if arch == "arm64" {
		if osName == "darwin" {
			arch = "arm64"
		} else {
			arch = "arm64"
		}
	}

	// Asset naming pattern: sl_{version}_{os}_{arch}.tar.gz or sl_{version}_{os}_{arch}
	var assetName string
	if osName == "windows" {
		assetName = fmt.Sprintf("sl_%s_%s_%s.zip", version, osName, arch)
	} else {
		assetName = fmt.Sprintf("sl_%s_%s_%s.tar.gz", version, osName, arch)
	}

	return fmt.Sprintf("https://github.com/specledger/specledger/releases/download/v%s/%s", version, assetName), nil
}

// downloadFile downloads a file to a temporary location.
func downloadFile(ctx context.Context, url string) (string, error) {
	client := &http.Client{
		Timeout: CheckTimeout * 6, // 30 seconds for download
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", fmt.Sprintf("SpecLedger-CLI/%s", GetVersion()))

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// Create temp file
	tmpFile, err := os.CreateTemp("", "sl-update-*")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	// Write response body to temp file
	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
}

// replaceBinary replaces the current binary with the new one.
func replaceBinary(newBinary, currentBinary string) error {
	// First, try to extract if it's a tar.gz
	if strings.HasSuffix(newBinary, ".tar.gz") {
		// Extract the binary from tar.gz
		extractedBinary, err := extractTarGz(newBinary)
		if err != nil {
			return fmt.Errorf("failed to extract archive: %w", err)
		}
		defer os.Remove(extractedBinary)
		newBinary = extractedBinary
	}

	// Backup the current binary
	backup := currentBinary + ".bak"
	if err := copyFile(currentBinary, backup); err != nil {
		// Non-fatal - might be first install
	}

	// Copy new binary to current location
	if err := copyFile(newBinary, currentBinary); err != nil {
		// Try to restore backup
		if _, restoreErr := os.Stat(backup); restoreErr == nil {
			copyFile(backup, currentBinary)
		}
		return err
	}

	// Remove backup on success
	os.Remove(backup)

	return nil
}

// extractTarGz extracts a tar.gz file and returns the path to the sl binary.
func extractTarGz(archivePath string) (string, error) {
	// Create temp directory for extraction
	tmpDir, err := os.MkdirTemp("", "sl-extract-*")
	if err != nil {
		return "", err
	}

	// Use tar command (available on all Unix-like systems)
	cmd := exec.Command("tar", "-xzf", archivePath, "-C", tmpDir)
	if err := cmd.Run(); err != nil {
		return "", err
	}

	// Find the sl binary
	slPath := filepath.Join(tmpDir, "sl")
	if _, err := os.Stat(slPath); err != nil {
		return "", fmt.Errorf("binary not found in archive")
	}

	return slPath, nil
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Get source file info for permissions
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
