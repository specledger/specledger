package version

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	// GitHubAPIURL is the GitHub Releases API endpoint
	GitHubAPIURL = "https://api.github.com/repos/specledger/specledger/releases/latest"

	// CheckTimeout is the timeout for version check requests
	CheckTimeout = 5 * time.Second
)

// GitHubRelease represents the GitHub Releases API response
type GitHubRelease struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	HTMLURL     string    `json:"html_url"`
	PublishedAt time.Time `json:"published_at"`
}

// CheckLatestVersion queries GitHub Releases API for the latest version.
// Returns VersionInfo with the latest version details, or an error message if check fails.
func CheckLatestVersion(ctx context.Context) *VersionInfo {
	info := &VersionInfo{
		CurrentVersion: GetVersion(),
		CheckedAt:      time.Now(),
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: CheckTimeout,
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, GitHubAPIURL, nil)
	if err != nil {
		info.Error = fmt.Sprintf("failed to create request: %v", err)
		return info
	}

	// Set headers
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", fmt.Sprintf("SpecLedger-CLI/%s", GetVersion()))

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		info.Error = fmt.Sprintf("network error: %v", err)
		return info
	}
	defer resp.Body.Close()

	// Check for rate limit or other errors
	if resp.StatusCode == http.StatusForbidden {
		info.Error = "rate limited"
		return info
	}

	if resp.StatusCode == http.StatusNotFound {
		info.Error = "release not found"
		return info
	}

	if resp.StatusCode != http.StatusOK {
		info.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
		return info
	}

	// Parse response
	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		info.Error = fmt.Sprintf("parse error: %v", err)
		return info
	}

	// Extract version from tag (remove 'v' prefix)
	info.LatestVersion = strings.TrimPrefix(release.TagName, "v")
	info.LatestURL = release.HTMLURL

	// Compare versions
	info.UpdateAvailable = isNewerVersion(info.CurrentVersion, info.LatestVersion)

	return info
}

// isNewerVersion compares two semantic versions and returns true if latest > current.
// Handles versions like "1.2.3", "1.2.3-beta", "dev".
func isNewerVersion(current, latest string) bool {
	// Dev builds are always considered outdated
	if current == "dev" || current == "" {
		return latest != "" && latest != "dev"
	}

	// Can't compare if latest is empty
	if latest == "" || latest == "dev" {
		return false
	}

	// Simple semver comparison (major.minor.patch)
	currentParts := parseSemver(current)
	latestParts := parseSemver(latest)

	for i := 0; i < 3; i++ {
		if latestParts[i] > currentParts[i] {
			return true
		}
		if latestParts[i] < currentParts[i] {
			return false
		}
	}

	return false
}

// parseSemver parses a semantic version string into [major, minor, patch].
// Returns [0, 0, 0] for invalid versions.
func parseSemver(v string) [3]int {
	var parts [3]int

	// Remove any pre-release suffix (e.g., "-beta.1")
	if idx := strings.Index(v, "-"); idx >= 0 {
		v = v[:idx]
	}

	// Split by "."
	segments := strings.SplitN(v, ".", 3)
	for i, seg := range segments {
		if i >= 3 {
			break
		}
		// Parse as integer, ignoring errors
		var val int
		if _, err := fmt.Sscanf(seg, "%d", &val); err == nil {
			parts[i] = val
		}
	}

	return parts
}
