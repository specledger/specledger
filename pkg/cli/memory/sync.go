package memory

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
	syncTimeout    = 30 * time.Second
	knowledgeTable = "knowledge_entries"
)

// SyncClient handles Supabase PostgREST operations for knowledge entries.
type SyncClient struct {
	baseURL string
	anonKey string
	client  *http.Client
}

// NewSyncClient creates a new sync client for cloud operations.
func NewSyncClient() *SyncClient {
	return &SyncClient{
		baseURL: auth.GetSupabaseURL(),
		anonKey: auth.GetSupabaseAnonKey(),
		client:  &http.Client{Timeout: syncTimeout},
	}
}

// CloudEntry represents a knowledge entry in cloud format (flat scores).
type CloudEntry struct {
	ID               string    `json:"id"`
	ProjectID        string    `json:"project_id"`
	Title            string    `json:"title"`
	Description      string    `json:"description"`
	Tags             []string  `json:"tags"`
	SourceSessionID  string    `json:"source_session_id,omitempty"`
	SourceBranch     string    `json:"source_branch,omitempty"`
	ScoreRecurrence  float64   `json:"score_recurrence"`
	ScoreImpact      float64   `json:"score_impact"`
	ScoreSpecificity float64   `json:"score_specificity"`
	CompositeScore   float64   `json:"composite_score"`
	Status           string    `json:"status"`
	RecurrenceCount  int       `json:"recurrence_count"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// toCloudEntry converts a local KnowledgeEntry to a CloudEntry.
func toCloudEntry(entry *KnowledgeEntry, projectID string) *CloudEntry {
	return &CloudEntry{
		ID:               entry.ID,
		ProjectID:        projectID,
		Title:            entry.Title,
		Description:      entry.Description,
		Tags:             entry.Tags,
		SourceSessionID:  entry.SourceSessionID,
		SourceBranch:     entry.SourceBranch,
		ScoreRecurrence:  entry.Scores.Recurrence,
		ScoreImpact:      entry.Scores.Impact,
		ScoreSpecificity: entry.Scores.Specificity,
		CompositeScore:   entry.Scores.Composite,
		Status:           string(entry.Status),
		RecurrenceCount:  entry.RecurrenceCount,
		CreatedAt:        entry.CreatedAt,
		UpdatedAt:        entry.UpdatedAt,
	}
}

// fromCloudEntry converts a CloudEntry to a local KnowledgeEntry.
func fromCloudEntry(ce *CloudEntry) *KnowledgeEntry {
	return &KnowledgeEntry{
		ID:              ce.ID,
		Title:           ce.Title,
		Description:     ce.Description,
		Tags:            ce.Tags,
		SourceSessionID: ce.SourceSessionID,
		SourceBranch:    ce.SourceBranch,
		Scores: Score{
			Recurrence:  ce.ScoreRecurrence,
			Impact:      ce.ScoreImpact,
			Specificity: ce.ScoreSpecificity,
			Composite:   ce.CompositeScore,
		},
		Status:          EntryStatus(ce.Status),
		RecurrenceCount: ce.RecurrenceCount,
		CreatedAt:       ce.CreatedAt,
		UpdatedAt:       ce.UpdatedAt,
	}
}

// SyncResult contains the results of a sync operation.
type SyncResult struct {
	Pushed  int
	Pulled  int
	Errors  int
	Skipped int
}

// Push uploads local promoted entries to cloud.
func (c *SyncClient) Push(accessToken, projectID string, entries []*KnowledgeEntry) (*SyncResult, error) {
	result := &SyncResult{}

	for _, entry := range entries {
		ce := toCloudEntry(entry, projectID)
		jsonBody, err := json.Marshal(ce)
		if err != nil {
			result.Errors++
			continue
		}

		reqURL := fmt.Sprintf("%s/rest/v1/%s", c.baseURL, knowledgeTable)

		req, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewReader(jsonBody))
		if err != nil {
			result.Errors++
			continue
		}

		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("apikey", c.anonKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Prefer", "resolution=merge-duplicates,return=minimal")

		resp, err := c.client.Do(req)
		if err != nil {
			result.Errors++
			continue
		}
		resp.Body.Close()

		if resp.StatusCode >= 400 {
			result.Errors++
			continue
		}

		result.Pushed++
	}

	return result, nil
}

// Pull downloads entries from cloud and returns them.
func (c *SyncClient) Pull(accessToken, projectID string) ([]*KnowledgeEntry, error) {
	params := url.Values{}
	params.Set("select", "*")
	params.Set("project_id", "eq."+projectID)
	params.Set("order", "composite_score.desc")

	reqURL := fmt.Sprintf("%s/rest/v1/%s?%s", c.baseURL, knowledgeTable, params.Encode())

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("apikey", c.anonKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	var cloudEntries []CloudEntry
	if err := json.Unmarshal(body, &cloudEntries); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var entries []*KnowledgeEntry
	for i := range cloudEntries {
		entries = append(entries, fromCloudEntry(&cloudEntries[i]))
	}

	return entries, nil
}
