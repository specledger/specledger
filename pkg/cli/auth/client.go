package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	// ProductionAPIURL is the production API base URL
	ProductionAPIURL = "https://app.specledger.io"
	// DevAPIURL is the development API base URL
	DevAPIURL = "http://localhost:3000"

	// SupabaseURL is the Supabase project URL
	SupabaseURL = "https://iituikpbiesgofuraclk.supabase.co"
	// SupabaseAnonKey is the public anon key for Supabase (safe to expose)
	SupabaseAnonKey = "sb_publishable_KpaZ2lKPu6eJ5WLqheu9_A_J9dYhGQb"
)

// GetSupabaseURL returns the Supabase project URL
func GetSupabaseURL() string {
	if envURL := os.Getenv("SPECLEDGER_SUPABASE_URL"); envURL != "" {
		return envURL
	}
	return SupabaseURL
}

// GetSupabaseAnonKey returns the Supabase anon key
func GetSupabaseAnonKey() string {
	if envKey := os.Getenv("SPECLEDGER_SUPABASE_ANON_KEY"); envKey != "" {
		return envKey
	}
	return SupabaseAnonKey
}

// getAPIURL returns the API URL based on environment
func getAPIURL() string {
	if envURL := os.Getenv("SPECLEDGER_API_URL"); envURL != "" {
		return envURL
	}
	if env := os.Getenv("SPECLEDGER_ENV"); env == "dev" || env == "development" {
		return DevAPIURL
	}
	return ProductionAPIURL
}

// RefreshTokenResponse represents the response from token refresh
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	Email        string `json:"email"`
	UserID       string `json:"user_id"`
	Error        string `json:"error"`
	ErrorMessage string `json:"error_description"`
}

// RefreshAccessToken uses the refresh token to get a new access token
func RefreshAccessToken(refreshToken string) (*Credentials, error) {
	apiURL := getAPIURL()
	endpoint := fmt.Sprintf("%s/api/cli/refresh", apiURL)

	// Prepare request body
	reqBody := map[string]string{
		"refresh_token": refreshToken,
	}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP request
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var tokenResp RefreshTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if tokenResp.Error != "" {
		errMsg := tokenResp.Error
		if tokenResp.ErrorMessage != "" {
			errMsg = tokenResp.ErrorMessage
		}
		return nil, fmt.Errorf("refresh failed: %s", errMsg)
	}

	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("no access token in response")
	}

	expiresIn := tokenResp.ExpiresIn
	if expiresIn == 0 {
		expiresIn = 3600 // Default 1 hour
	}

	return &Credentials{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresIn:    expiresIn,
		CreatedAt:    time.Now().Unix(),
		UserEmail:    tokenResp.Email,
		UserID:       tokenResp.UserID,
	}, nil
}

// GetValidAccessToken returns a valid access token, refreshing if needed
func GetValidAccessToken() (string, error) {
	creds, err := LoadCredentials()
	if err != nil {
		return "", fmt.Errorf("failed to load credentials: %w", err)
	}

	if creds == nil {
		return "", fmt.Errorf("not authenticated, please run 'sl auth login'")
	}

	// If token is still valid, return it
	if !creds.IsExpired() {
		return creds.AccessToken, nil
	}

	// Token expired, try to refresh
	if creds.RefreshToken == "" {
		return "", fmt.Errorf("token expired and no refresh token available, please run 'sl auth login'")
	}

	newCreds, err := RefreshAccessToken(creds.RefreshToken)
	if err != nil {
		return "", fmt.Errorf("failed to refresh token: %w (please run 'sl auth login')", err)
	}

	// Preserve user info if not returned from refresh
	if newCreds.UserEmail == "" {
		newCreds.UserEmail = creds.UserEmail
	}
	if newCreds.UserID == "" {
		newCreds.UserID = creds.UserID
	}

	// Save the new credentials
	if err := SaveCredentials(newCreds); err != nil {
		return "", fmt.Errorf("failed to save refreshed credentials: %w", err)
	}

	return newCreds.AccessToken, nil
}
