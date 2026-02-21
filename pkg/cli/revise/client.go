package revise

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/specledger/specledger/pkg/cli/auth"
)

const clientTimeout = 30 * time.Second

// apiProject is the PostgREST response for the projects table.
type apiProject struct {
	ID            string `json:"id"`
	DefaultBranch string `json:"default_branch"`
}

// apiSpec is the PostgREST response for the specs table.
type apiSpec struct {
	ID      string `json:"id"`
	SpecKey string `json:"spec_key"`
	Phase   string `json:"phase"`
}

// apiChange is the PostgREST response for the changes table.
type apiChange struct {
	ID         string `json:"id"`
	SpecID     string `json:"spec_id"`
	HeadBranch string `json:"head_branch"`
	BaseBranch string `json:"base_branch"`
	State      string `json:"state"`
}

// SpecWithCommentCount is used by the branch picker to list specs with pending review comments.
type SpecWithCommentCount struct {
	SpecKey      string
	CommentCount int
}

// ReviseClient handles all Supabase PostgREST calls for the revise command.
type ReviseClient struct {
	baseURL     string
	anonKey     string
	accessToken string
	httpClient  *http.Client
}

// NewReviseClient creates a ReviseClient using the configured Supabase URL and anon key.
// The provided accessToken is used for Authorization headers; call doWithRetry to handle 401s.
func NewReviseClient(accessToken string) *ReviseClient {
	return &ReviseClient{
		baseURL:     auth.GetSupabaseURL(),
		anonKey:     auth.GetSupabaseAnonKey(),
		accessToken: accessToken,
		httpClient:  &http.Client{Timeout: clientTimeout},
	}
}

// doWithRetry executes fn with the current access token. On a 401 or PGRST303, it calls
// auth.GetValidAccessToken() to force a token refresh, updates the stored token, and retries once.
func (c *ReviseClient) doWithRetry(fn func(token string) (*http.Response, error)) (*http.Response, error) {
	resp, err := fn(c.accessToken)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()

		newToken, err := auth.GetValidAccessToken()
		if err != nil {
			return nil, fmt.Errorf("token refresh failed: %w (run 'sl auth login')", err)
		}
		c.accessToken = newToken

		resp, err = fn(c.accessToken)
		if err != nil {
			return nil, err
		}
	}

	return resp, nil
}

// get performs an authenticated GET request to the given PostgREST path.
func (c *ReviseClient) get(token, path string) (*http.Response, error) {
	reqURL := c.baseURL + path

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("apikey", c.anonKey)
	req.Header.Set("Accept", "application/json")

	return c.httpClient.Do(req)
}

// readJSON reads and unmarshals a JSON response body.
func readJSON(resp *http.Response, dest interface{}) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp map[string]any
		if json.Unmarshal(body, &errResp) == nil {
			if msg, ok := errResp["message"].(string); ok {
				return fmt.Errorf("API error (%d): %s", resp.StatusCode, msg)
			}
		}
		return fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	if err := json.Unmarshal(body, dest); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	return nil
}

// GetProject fetches the project record for the given GitHub repo.
func (c *ReviseClient) GetProject(repoOwner, repoName string) (*apiProject, error) {
	path := fmt.Sprintf(
		"/rest/v1/projects?repo_owner=eq.%s&repo_name=eq.%s&select=id,default_branch",
		url.QueryEscape(repoOwner),
		url.QueryEscape(repoName),
	)

	resp, err := c.doWithRetry(func(token string) (*http.Response, error) {
		return c.get(token, path)
	})
	if err != nil {
		return nil, fmt.Errorf("GetProject: %w", err)
	}

	var projects []apiProject
	if err := readJSON(resp, &projects); err != nil {
		return nil, fmt.Errorf("GetProject: %w", err)
	}

	if len(projects) == 0 {
		return nil, fmt.Errorf("project not found: %s/%s", repoOwner, repoName)
	}

	return &projects[0], nil
}

// GetSpec fetches the spec record for the given project and spec key (branch name).
func (c *ReviseClient) GetSpec(projectID, specKey string) (*apiSpec, error) {
	path := fmt.Sprintf(
		"/rest/v1/specs?project_id=eq.%s&spec_key=eq.%s&select=id,spec_key,phase",
		url.QueryEscape(projectID),
		url.QueryEscape(specKey),
	)

	resp, err := c.doWithRetry(func(token string) (*http.Response, error) {
		return c.get(token, path)
	})
	if err != nil {
		return nil, fmt.Errorf("GetSpec: %w", err)
	}

	var specs []apiSpec
	if err := readJSON(resp, &specs); err != nil {
		return nil, fmt.Errorf("GetSpec: %w", err)
	}

	if len(specs) == 0 {
		return nil, fmt.Errorf("spec not found: %s", specKey)
	}

	return &specs[0], nil
}

// GetChange fetches the open change record for the given spec.
func (c *ReviseClient) GetChange(specID string) (*apiChange, error) {
	path := fmt.Sprintf(
		"/rest/v1/changes?spec_id=eq.%s&select=id,spec_id,head_branch,base_branch,state",
		url.QueryEscape(specID),
	)

	resp, err := c.doWithRetry(func(token string) (*http.Response, error) {
		return c.get(token, path)
	})
	if err != nil {
		return nil, fmt.Errorf("GetChange: %w", err)
	}

	var changes []apiChange
	if err := readJSON(resp, &changes); err != nil {
		return nil, fmt.Errorf("GetChange: %w", err)
	}

	if len(changes) == 0 {
		return nil, fmt.Errorf("no change found for spec %s", specID)
	}

	// Return the most recent open change, or the first change if none are open.
	for _, ch := range changes {
		if ch.State == "open" {
			return &ch, nil
		}
	}

	return &changes[0], nil
}

// FetchComments fetches all unresolved top-level review comments for the given change.
// Results are ordered by created_at ascending (oldest first).
func (c *ReviseClient) FetchComments(changeID string) ([]ReviewComment, error) {
	path := fmt.Sprintf(
		"/rest/v1/review_comments?change_id=eq.%s&is_resolved=eq.false&parent_comment_id=is.null"+
			"&select=id,file_path,content,selected_text,line,start_line,author_name,author_email,created_at"+
			"&order=created_at.asc",
		url.QueryEscape(changeID),
	)

	resp, err := c.doWithRetry(func(token string) (*http.Response, error) {
		return c.get(token, path)
	})
	if err != nil {
		return nil, fmt.Errorf("FetchComments: %w", err)
	}

	var comments []ReviewComment
	if err := readJSON(resp, &comments); err != nil {
		return nil, fmt.Errorf("FetchComments: %w", err)
	}

	return comments, nil
}

// ResolveComment marks a review comment as resolved via PATCH.
func (c *ReviseClient) ResolveComment(commentID string) error {
	reqURL := fmt.Sprintf(
		"%s/rest/v1/review_comments?id=eq.%s",
		c.baseURL,
		url.QueryEscape(commentID),
	)

	body := []byte(`{"is_resolved":true}`)

	patchFn := func(token string) (*http.Response, error) {
		req, err := http.NewRequest(http.MethodPatch, reqURL, bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("apikey", c.anonKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Prefer", "return=minimal")

		return c.httpClient.Do(req)
	}

	resp, err := c.doWithRetry(patchFn)
	if err != nil {
		return fmt.Errorf("ResolveComment: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ResolveComment: API error (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// ListSpecsWithComments returns all specs for the project that have unresolved review comments.
//
// Implementation: 3 sequential calls + client-side aggregation.
//   1. GET /rest/v1/specs?project_id=eq.{pid}&select=id,spec_key
//   2. GET /rest/v1/changes?spec_id=in.({spec_ids})&select=id,spec_id
//   3. GET /rest/v1/review_comments?change_id=in.({change_ids})&is_resolved=eq.false&select=id,change_id
//
// Client-side: group comments by change → map change to spec → count per spec_key.
// See contracts/postgrest-api.md §6 for the rationale.
func (c *ReviseClient) ListSpecsWithComments(projectID string) ([]SpecWithCommentCount, error) {
	// Step 1: Get all specs for this project
	specsPath := fmt.Sprintf(
		"/rest/v1/specs?project_id=eq.%s&select=id,spec_key",
		url.QueryEscape(projectID),
	)

	resp, err := c.doWithRetry(func(token string) (*http.Response, error) {
		return c.get(token, specsPath)
	})
	if err != nil {
		return nil, fmt.Errorf("ListSpecsWithComments: fetch specs: %w", err)
	}

	var specs []apiSpec
	if err := readJSON(resp, &specs); err != nil {
		return nil, fmt.Errorf("ListSpecsWithComments: parse specs: %w", err)
	}

	if len(specs) == 0 {
		return nil, nil
	}

	// Build spec_id → spec_key map and comma-separated id list for the next query
	specIDToKey := make(map[string]string, len(specs))
	specIDs := make([]string, 0, len(specs))
	for _, s := range specs {
		specIDToKey[s.ID] = s.SpecKey
		specIDs = append(specIDs, s.ID)
	}
	specIDList := "(" + strings.Join(specIDs, ",") + ")"

	// Step 2: Get all changes for these specs
	changesPath := fmt.Sprintf(
		"/rest/v1/changes?spec_id=in.%s&select=id,spec_id",
		url.QueryEscape(specIDList),
	)

	resp, err = c.doWithRetry(func(token string) (*http.Response, error) {
		return c.get(token, changesPath)
	})
	if err != nil {
		return nil, fmt.Errorf("ListSpecsWithComments: fetch changes: %w", err)
	}

	var changes []apiChange
	if err := readJSON(resp, &changes); err != nil {
		return nil, fmt.Errorf("ListSpecsWithComments: parse changes: %w", err)
	}

	if len(changes) == 0 {
		return nil, nil
	}

	// Build change_id → spec_key map
	changeIDToSpecKey := make(map[string]string, len(changes))
	changeIDs := make([]string, 0, len(changes))
	for _, ch := range changes {
		changeIDToSpecKey[ch.ID] = specIDToKey[ch.SpecID]
		changeIDs = append(changeIDs, ch.ID)
	}
	changeIDList := "(" + strings.Join(changeIDs, ",") + ")"

	// Step 3: Count unresolved top-level comments per change
	commentsPath := fmt.Sprintf(
		"/rest/v1/review_comments?change_id=in.%s&is_resolved=eq.false&parent_comment_id=is.null&select=id,change_id",
		url.QueryEscape(changeIDList),
	)

	resp, err = c.doWithRetry(func(token string) (*http.Response, error) {
		return c.get(token, commentsPath)
	})
	if err != nil {
		return nil, fmt.Errorf("ListSpecsWithComments: fetch comments: %w", err)
	}

	type commentRow struct {
		ID       string `json:"id"`
		ChangeID string `json:"change_id"`
	}

	var commentRows []commentRow
	if err := readJSON(resp, &commentRows); err != nil {
		return nil, fmt.Errorf("ListSpecsWithComments: parse comments: %w", err)
	}

	// Aggregate: count comments per spec_key
	specKeyCount := make(map[string]int)
	for _, row := range commentRows {
		specKey := changeIDToSpecKey[row.ChangeID]
		if specKey != "" {
			specKeyCount[specKey]++
		}
	}

	result := make([]SpecWithCommentCount, 0, len(specKeyCount))
	for specKey, count := range specKeyCount {
		if count > 0 {
			result = append(result, SpecWithCommentCount{SpecKey: specKey, CommentCount: count})
		}
	}

	return result, nil
}
