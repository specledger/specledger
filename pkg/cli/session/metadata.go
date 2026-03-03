package session

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/specledger/specledger/pkg/cli/auth"
)

const (
	// MetadataTimeout is the timeout for metadata operations
	MetadataTimeout = 30 * time.Second
	// SessionsTable is the PostgREST table name
	SessionsTable = "sessions"
)

// MetadataClient handles Supabase PostgREST operations for session metadata
type MetadataClient struct {
	baseURL string
	anonKey string
	client  *http.Client
}

// NewMetadataClient creates a new metadata client
func NewMetadataClient() *MetadataClient {
	return &MetadataClient{
		baseURL: auth.GetSupabaseURL(),
		anonKey: auth.GetSupabaseAnonKey(),
		client:  &http.Client{Timeout: MetadataTimeout},
	}
}

// CreateSessionInput represents the input for creating a session
type CreateSessionInput struct {
	ProjectID     string        `json:"project_id"`
	FeatureBranch string        `json:"feature_branch"`
	CommitHash    *string       `json:"commit_hash,omitempty"`
	TaskID        *string       `json:"task_id,omitempty"`
	AuthorID      string        `json:"author_id"`
	StoragePath   string        `json:"storage_path"`
	Status        SessionStatus `json:"status"`
	SizeBytes     int64         `json:"size_bytes"`
	RawSizeBytes  int64         `json:"raw_size_bytes"`
	MessageCount  int           `json:"message_count"`
}

// Create creates a new session metadata record
func (m *MetadataClient) Create(accessToken string, input *CreateSessionInput) (*SessionMetadata, error) {
	url := fmt.Sprintf("%s/rest/v1/%s", m.baseURL, SessionsTable)

	jsonBody, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("apikey", m.anonKey)
	req.Header.Set("Prefer", "return=representation")

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("create failed: %w", err)
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
				return nil, fmt.Errorf("create failed (%d): %s", resp.StatusCode, msg)
			}
		}
		return nil, fmt.Errorf("create failed with status %d: %s", resp.StatusCode, string(body))
	}

	// PostgREST returns an array for INSERT with return=representation
	var sessions []SessionMetadata
	if err := json.Unmarshal(body, &sessions); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(sessions) == 0 {
		return nil, fmt.Errorf("no session returned from create")
	}

	return &sessions[0], nil
}

// QueryOptions represents options for querying sessions
type QueryOptions struct {
	ProjectID     string
	FeatureBranch string
	CommitHash    string
	TaskID        string
	AuthorID      string
	StartDate     *time.Time
	EndDate       *time.Time
	Limit         int
	Offset        int
	OrderBy       string
	OrderDesc     bool
}

// Query queries sessions based on options
func (m *MetadataClient) Query(accessToken string, opts *QueryOptions) ([]SessionMetadata, error) {
	reqURL := fmt.Sprintf("%s/rest/v1/%s", m.baseURL, SessionsTable)

	// Build query parameters
	params := url.Values{}
	params.Set("select", "*")

	if opts.ProjectID != "" {
		params.Set("project_id", "eq."+opts.ProjectID)
	}
	if opts.FeatureBranch != "" {
		params.Set("feature_branch", "eq."+opts.FeatureBranch)
	}
	if opts.CommitHash != "" {
		// Use like for partial commit hash (prefix match)
		// PostgREST uses % as wildcard, not *
		if len(opts.CommitHash) < 40 {
			params.Set("commit_hash", "like."+opts.CommitHash+"%")
		} else {
			params.Set("commit_hash", "eq."+opts.CommitHash)
		}
	}
	if opts.TaskID != "" {
		params.Set("task_id", "eq."+opts.TaskID)
	}
	if opts.AuthorID != "" {
		params.Set("author_id", "eq."+opts.AuthorID)
	}
	// Date range filtering using PostgREST 'and' operator
	if opts.StartDate != nil && opts.EndDate != nil {
		// Both dates: use and() for proper AND logic
		params.Set("and", fmt.Sprintf("(created_at.gte.%s,created_at.lte.%s)",
			opts.StartDate.Format(time.RFC3339), opts.EndDate.Format(time.RFC3339)))
	} else if opts.StartDate != nil {
		params.Set("created_at", "gte."+opts.StartDate.Format(time.RFC3339))
	} else if opts.EndDate != nil {
		params.Set("created_at", "lte."+opts.EndDate.Format(time.RFC3339))
	}

	// Order
	orderBy := "created_at"
	if opts.OrderBy != "" {
		orderBy = opts.OrderBy
	}
	if opts.OrderDesc {
		params.Set("order", orderBy+".desc")
	} else {
		params.Set("order", orderBy+".asc")
	}

	// Pagination
	if opts.Limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", opts.Limit))
	}
	if opts.Offset > 0 {
		params.Set("offset", fmt.Sprintf("%d", opts.Offset))
	}

	fullURL := reqURL + "?" + params.Encode()

	req, err := http.NewRequest(http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("apikey", m.anonKey)

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("query failed with status %d: %s", resp.StatusCode, string(body))
	}

	var sessions []SessionMetadata
	if err := json.Unmarshal(body, &sessions); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return sessions, nil
}

// GetByID retrieves a session by its ID
func (m *MetadataClient) GetByID(accessToken string, sessionID string) (*SessionMetadata, error) {
	reqURL := fmt.Sprintf("%s/rest/v1/%s?id=eq.%s&select=*", m.baseURL, SessionsTable, sessionID)

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("apikey", m.anonKey)

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("get failed with status %d: %s", resp.StatusCode, string(body))
	}

	var sessions []SessionMetadata
	if err := json.Unmarshal(body, &sessions); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(sessions) == 0 {
		return nil, nil
	}

	return &sessions[0], nil
}

// GetByCommitHash retrieves a session by project and commit hash
func (m *MetadataClient) GetByCommitHash(accessToken string, projectID string, commitHash string) (*SessionMetadata, error) {
	sessions, err := m.Query(accessToken, &QueryOptions{
		ProjectID:  projectID,
		CommitHash: commitHash,
		Limit:      1,
	})
	if err != nil {
		return nil, err
	}
	if len(sessions) == 0 {
		return nil, nil
	}
	return &sessions[0], nil
}

// GetByTaskID retrieves a session by project and task ID
func (m *MetadataClient) GetByTaskID(accessToken string, projectID string, taskID string) (*SessionMetadata, error) {
	sessions, err := m.Query(accessToken, &QueryOptions{
		ProjectID: projectID,
		TaskID:    taskID,
		Limit:     1,
	})
	if err != nil {
		return nil, err
	}
	if len(sessions) == 0 {
		return nil, nil
	}
	return &sessions[0], nil
}

// ListByFeature retrieves all sessions for a feature branch
func (m *MetadataClient) ListByFeature(accessToken string, projectID string, featureBranch string) ([]SessionMetadata, error) {
	return m.Query(accessToken, &QueryOptions{
		ProjectID:     projectID,
		FeatureBranch: featureBranch,
		OrderBy:       "created_at",
		OrderDesc:     true,
	})
}
