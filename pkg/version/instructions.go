package version

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// InstallationMethod represents how the CLI was installed
type InstallationMethod string

const (
	MethodHomebrew InstallationMethod = "homebrew"
	MethodGoInstall InstallationMethod = "go-install"
	MethodBinary    InstallationMethod = "binary"
)

// DetectInstallationMethod attempts to determine how the CLI was installed.
func DetectInstallationMethod() InstallationMethod {
	execPath, err := os.Executable()
	if err != nil {
		return MethodBinary
	}

	// Resolve symlinks
	realPath, err := filepath.EvalSymlinks(execPath)
	if err != nil {
		realPath = execPath
	}

	// Check for Homebrew
	if isHomebrewInstall(realPath) {
		return MethodHomebrew
	}

	// Check for Go install
	if isGoInstall(realPath) {
		return MethodGoInstall
	}

	return MethodBinary
}

// isHomebrewInstall checks if the binary is in a Homebrew prefix.
func isHomebrewInstall(path string) bool {
	// Common Homebrew paths
	homebrewPaths := []string{
		"/opt/homebrew/",     // Apple Silicon
		"/usr/local/Cellar/", // Intel Mac
		"/home/linuxbrew/",   // Linux
	}

	for _, prefix := range homebrewPaths {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	// Check if HOMEBREW_PREFIX is set
	if prefix := os.Getenv("HOMEBREW_PREFIX"); prefix != "" {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	return false
}

// isGoInstall checks if the binary is in GOPATH/bin.
func isGoInstall(path string) bool {
	// Check GOPATH
	if gopath := os.Getenv("GOPATH"); gopath != "" {
		goBinPath := filepath.Join(gopath, "bin")
		if strings.HasPrefix(path, goBinPath) {
			return true
		}
	}

	// Check GOBIN
	if gobin := os.Getenv("GOBIN"); gobin != "" {
		if strings.HasPrefix(path, gobin) {
			return true
		}
	}

	// Check common Go paths
	goPaths := []string{
		"/go/bin/",
		"/sdk/go/bin/",
	}

	for _, prefix := range goPaths {
		if strings.Contains(path, prefix) {
			return true
		}
	}

	return false
}

// GetUpdateInstructions returns installation-method-specific update instructions.
func GetUpdateInstructions() string {
	method := DetectInstallationMethod()
	return GetUpdateInstructionsForMethod(method)
}

// GetUpdateInstructionsForMethod returns update instructions for a specific installation method.
func GetUpdateInstructionsForMethod(method InstallationMethod) string {
	switch method {
	case MethodHomebrew:
		return getHomebrewInstructions()
	case MethodGoInstall:
		return getGoInstallInstructions()
	default:
		return getBinaryInstructions()
	}
}

func getHomebrewInstructions() string {
	return `  brew upgrade specledger`
}

func getGoInstallInstructions() string {
	return `  go install github.com/specledger/specledger/cmd/sl@latest`
}

func getBinaryInstructions() string {
	var sb strings.Builder

	sb.WriteString("  Download from:\n")
	sb.WriteString("    https://github.com/specledger/specledger/releases/latest\n")
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("  Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH))

	return sb.String()
}

// FormatUpdateMessage formats a complete update message with all instructions.
func FormatUpdateMessage(currentVersion, latestVersion string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("A new version (%s) is available!\n", latestVersion))
	sb.WriteString(fmt.Sprintf("You are currently on %s\n\n", currentVersion))
	sb.WriteString("Update with one of:\n")

	method := DetectInstallationMethod()

	// Show the detected method first
	sb.WriteString(GetUpdateInstructionsForMethod(method))
	sb.WriteString("\n")

	// If not detected, show all options
	if method == MethodBinary {
		sb.WriteString("\nOr use:\n")
		sb.WriteString("  brew upgrade specledger          # Homebrew\n")
		sb.WriteString("  go install .../cmd/sl@latest     # Go install\n")
	}

	return sb.String()
}
