package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Credentials represents the stored authentication tokens
type Credentials struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // seconds until expiry
	CreatedAt    int64  `json:"created_at"` // unix timestamp when saved
	UserEmail    string `json:"user_email"`
	UserID       string `json:"user_id"`
}

// refreshBuffer is how far before actual expiry we consider a token expired,
// so we proactively refresh instead of racing the clock.
const refreshBuffer = 30 * time.Second

// IsExpired checks if the access token has expired (or is within the refresh buffer).
func (c *Credentials) IsExpired() bool {
	expiresAt := time.Unix(c.CreatedAt, 0).Add(time.Duration(c.ExpiresIn) * time.Second)
	return time.Now().After(expiresAt.Add(-refreshBuffer))
}

// ExpiresAt returns the expiration time
func (c *Credentials) ExpiresAt() time.Time {
	return time.Unix(c.CreatedAt, 0).Add(time.Duration(c.ExpiresIn) * time.Second)
}

// IsValid checks if credentials exist and have required fields
func (c *Credentials) IsValid() bool {
	return c.AccessToken != "" && c.RefreshToken != ""
}

// GetCredentialsPath returns the path to the credentials file
func GetCredentialsPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to HOME environment variable
		homeDir = os.Getenv("HOME")
	}
	return filepath.Join(homeDir, ".specledger", "credentials.json")
}

// LoadCredentials loads credentials from disk
func LoadCredentials() (*Credentials, error) {
	credPath := GetCredentialsPath()

	data, err := os.ReadFile(credPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No credentials stored
		}
		return nil, fmt.Errorf("failed to read credentials: %w", err)
	}

	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("failed to parse credentials: %w", err)
	}

	return &creds, nil
}

// SaveCredentials saves credentials to disk with secure permissions
func SaveCredentials(creds *Credentials) error {
	credPath := GetCredentialsPath()

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(credPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	// Write with restricted permissions (owner read/write only)
	if err := os.WriteFile(credPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write credentials: %w", err)
	}

	return nil
}

// DeleteCredentials removes the credentials file
func DeleteCredentials() error {
	credPath := GetCredentialsPath()

	err := os.Remove(credPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete credentials: %w", err)
	}

	return nil
}
