// Package version provides CLI version information and version checking utilities.
package version

import (
	"time"
)

// Build-time variables set by GoReleaser via ldflags
var (
	// Version is the semantic version of the CLI
	Version = "dev"
	// Commit is the git commit hash
	Commit = "unknown"
	// Date is the build date
	Date = "unknown"
	// BuildType is the build type (development, production)
	BuildType = "development"
)

// VersionInfo represents version information from GitHub Releases API.
type VersionInfo struct {
	CurrentVersion  string    `json:"current_version"`  // Installed CLI version (e.g., "1.2.0")
	LatestVersion   string    `json:"latest_version"`   // Latest available version (e.g., "1.3.0")
	LatestURL       string    `json:"latest_url"`       // URL to release page
	UpdateAvailable bool      `json:"update_available"` // true if LatestVersion > CurrentVersion
	CheckedAt       time.Time `json:"checked_at"`       // When the check was performed
	Error           string    `json:"error,omitempty"`  // Error message if check failed
}

// GetVersion returns the current CLI version.
func GetVersion() string {
	return Version
}

// GetCommit returns the git commit hash.
func GetCommit() string {
	return Commit
}

// GetBuildDate returns the build date.
func GetBuildDate() string {
	return Date
}

// GetBuildType returns the build type.
func GetBuildType() string {
	return BuildType
}

// IsDevBuild returns true if this is a development build.
func IsDevBuild() bool {
	return Version == "dev" || BuildType == "development"
}
