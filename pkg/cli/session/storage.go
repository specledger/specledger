package session

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/specledger/specledger/pkg/cli/auth"
)

const (
	// StorageBucket is the Supabase Storage bucket for sessions
	StorageBucket = "sessions"
	// StorageTimeout is the timeout for storage operations
	StorageTimeout = 30 * time.Second
)

// StorageClient handles Supabase Storage operations
type StorageClient struct {
	baseURL  string
	anonKey  string
	client   *http.Client
}

// NewStorageClient creates a new storage client
func NewStorageClient() *StorageClient {
	return &StorageClient{
		baseURL: auth.GetSupabaseURL(),
		anonKey: auth.GetSupabaseAnonKey(),
		client:  &http.Client{Timeout: StorageTimeout},
	}
}

// UploadResponse represents the response from an upload operation
type UploadResponse struct {
	Key string `json:"Key"`
	ID  string `json:"Id"`
}

// Upload uploads compressed session content to Supabase Storage
func (s *StorageClient) Upload(accessToken string, storagePath string, data []byte) (*UploadResponse, error) {
	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.baseURL, StorageBucket, storagePath)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/gzip")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("apikey", s.anonKey)
	req.Header.Set("x-upsert", "true") // Replace if exists

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("upload failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp map[string]interface{}
		if err := json.Unmarshal(body, &errResp); err == nil {
			if msg, ok := errResp["message"].(string); ok {
				return nil, fmt.Errorf("upload failed (%d): %s", resp.StatusCode, msg)
			}
		}
		return nil, fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	var uploadResp UploadResponse
	if err := json.Unmarshal(body, &uploadResp); err != nil {
		return nil, fmt.Errorf("failed to parse upload response: %w", err)
	}

	return &uploadResp, nil
}

// Download downloads session content from Supabase Storage
func (s *StorageClient) Download(accessToken string, storagePath string) ([]byte, error) {
	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.baseURL, StorageBucket, storagePath)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("apikey", s.anonKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("download failed with status %d: %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

// SignedURLResponse represents the response from signed URL generation
type SignedURLResponse struct {
	SignedURL string `json:"signedURL"`
}

// GetSignedURL generates a time-limited signed URL for session content
func (s *StorageClient) GetSignedURL(accessToken string, storagePath string, expiresIn int) (*SignedURLResponse, error) {
	url := fmt.Sprintf("%s/storage/v1/object/sign/%s/%s", s.baseURL, StorageBucket, storagePath)

	reqBody := map[string]int{"expiresIn": expiresIn}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("apikey", s.anonKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get signed URL: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("failed to get signed URL (%d): %s", resp.StatusCode, string(body))
	}

	var signedResp SignedURLResponse
	if err := json.Unmarshal(body, &signedResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &signedResp, nil
}

// BuildStoragePath constructs the storage path for a session
func BuildStoragePath(projectID, featureBranch, identifier string) string {
	return fmt.Sprintf("%s/%s/%s.json.gz", projectID, featureBranch, identifier)
}
