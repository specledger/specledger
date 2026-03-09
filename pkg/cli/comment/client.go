package comment

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

type apiProject struct {
	ID            string `json:"id"`
	DefaultBranch string `json:"default_branch"`
}

type apiSpec struct {
	ID      string `json:"id"`
	SpecKey string `json:"spec_key"`
	Phase   string `json:"phase"`
}

type apiChange struct {
	ID         string `json:"id"`
	SpecID     string `json:"spec_id"`
	HeadBranch string `json:"head_branch"`
	BaseBranch string `json:"base_branch"`
	State      string `json:"state"`
}

type Client struct {
	BaseURL     string
	AnonKey     string
	accessToken string
	HTTPClient  *http.Client
}

func NewClient(accessToken string) *Client {
	return &Client{
		BaseURL:     auth.GetSupabaseURL(),
		AnonKey:     auth.GetSupabaseAnonKey(),
		accessToken: accessToken,
		HTTPClient:  &http.Client{Timeout: clientTimeout},
	}
}

func (c *Client) DoWithRetry(fn func(token string) (*http.Response, error)) (*http.Response, error) {
	resp, err := fn(c.accessToken)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()

		newToken, err := auth.ForceRefreshAccessToken()
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

func (c *Client) Get(token, path string) (*http.Response, error) {
	reqURL := c.BaseURL + path

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("apikey", c.AnonKey)
	req.Header.Set("Accept", "application/json")

	return c.HTTPClient.Do(req)
}

func ReadJSON(resp *http.Response, dest interface{}) error {
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

func (c *Client) GetProject(repoOwner, repoName string) (*apiProject, error) {
	path := fmt.Sprintf(
		"/rest/v1/projects?repo_owner=eq.%s&repo_name=eq.%s&select=id,default_branch",
		url.QueryEscape(repoOwner),
		url.QueryEscape(repoName),
	)

	resp, err := c.DoWithRetry(func(token string) (*http.Response, error) {
		return c.Get(token, path)
	})
	if err != nil {
		return nil, fmt.Errorf("GetProject: %w", err)
	}

	var projects []apiProject
	if err := ReadJSON(resp, &projects); err != nil {
		return nil, fmt.Errorf("GetProject: %w", err)
	}

	if len(projects) == 0 {
		return nil, fmt.Errorf("project not found: %s/%s", repoOwner, repoName)
	}

	return &projects[0], nil
}

func (c *Client) GetSpec(projectID, specKey string) (*apiSpec, error) {
	path := fmt.Sprintf(
		"/rest/v1/specs?project_id=eq.%s&spec_key=eq.%s&select=id,spec_key,phase",
		url.QueryEscape(projectID),
		url.QueryEscape(specKey),
	)

	resp, err := c.DoWithRetry(func(token string) (*http.Response, error) {
		return c.Get(token, path)
	})
	if err != nil {
		return nil, fmt.Errorf("GetSpec: %w", err)
	}

	var specs []apiSpec
	if err := ReadJSON(resp, &specs); err != nil {
		return nil, fmt.Errorf("GetSpec: %w", err)
	}

	if len(specs) == 0 {
		return nil, fmt.Errorf("spec not found: %s", specKey)
	}

	return &specs[0], nil
}

func (c *Client) GetChange(specID string) (*apiChange, error) {
	path := fmt.Sprintf(
		"/rest/v1/changes?spec_id=eq.%s&select=id,spec_id,head_branch,base_branch,state",
		url.QueryEscape(specID),
	)

	resp, err := c.DoWithRetry(func(token string) (*http.Response, error) {
		return c.Get(token, path)
	})
	if err != nil {
		return nil, fmt.Errorf("GetChange: %w", err)
	}

	var changes []apiChange
	if err := ReadJSON(resp, &changes); err != nil {
		return nil, fmt.Errorf("GetChange: %w", err)
	}

	if len(changes) == 0 {
		return nil, fmt.Errorf("no change found for spec %s", specID)
	}

	for _, ch := range changes {
		if ch.State == "open" {
			return &ch, nil
		}
	}

	return &changes[0], nil
}

func (c *Client) FetchComments(changeID string) ([]ReviewComment, error) {
	path := fmt.Sprintf(
		"/rest/v1/review_comments?change_id=eq.%s&is_resolved=eq.false&parent_comment_id=is.null"+
			"&select=id,file_path,content,selected_text,line,start_line,author_name,author_email,created_at"+
			"&order=created_at.asc",
		url.QueryEscape(changeID),
	)

	resp, err := c.DoWithRetry(func(token string) (*http.Response, error) {
		return c.Get(token, path)
	})
	if err != nil {
		return nil, fmt.Errorf("FetchComments: %w", err)
	}

	var comments []ReviewComment
	if err := ReadJSON(resp, &comments); err != nil {
		return nil, fmt.Errorf("FetchComments: %w", err)
	}

	return comments, nil
}

func (c *Client) FetchCommentByID(commentID string) (*ReviewComment, error) {
	path := fmt.Sprintf(
		"/rest/v1/review_comments?id=eq.%s"+
			"&select=id,file_path,content,selected_text,line,start_line,author_name,author_email,is_resolved,created_at",
		url.QueryEscape(commentID),
	)

	resp, err := c.DoWithRetry(func(token string) (*http.Response, error) {
		return c.Get(token, path)
	})
	if err != nil {
		return nil, fmt.Errorf("FetchCommentByID: %w", err)
	}

	var comments []ReviewComment
	if err := ReadJSON(resp, &comments); err != nil {
		return nil, fmt.Errorf("FetchCommentByID: %w", err)
	}

	if len(comments) == 0 {
		return nil, fmt.Errorf("comment not found: %s", commentID)
	}

	return &comments[0], nil
}

func (c *Client) FetchReplies(changeID string) ([]ReviewComment, error) {
	path := fmt.Sprintf(
		"/rest/v1/review_comments?change_id=eq.%s&is_resolved=eq.false&parent_comment_id=not.is.null"+
			"&select=id,file_path,content,selected_text,parent_comment_id,author_name,author_email,created_at"+
			"&order=created_at.asc",
		url.QueryEscape(changeID),
	)

	resp, err := c.DoWithRetry(func(token string) (*http.Response, error) {
		return c.Get(token, path)
	})
	if err != nil {
		return nil, fmt.Errorf("FetchReplies: %w", err)
	}

	var replies []ReviewComment
	if err := ReadJSON(resp, &replies); err != nil {
		return nil, fmt.Errorf("FetchReplies: %w", err)
	}

	return replies, nil
}

func (c *Client) FetchRepliesByParentID(parentID string) ([]ReviewComment, error) {
	path := fmt.Sprintf(
		"/rest/v1/review_comments?parent_comment_id=eq.%s"+
			"&select=id,content,author_name,author_email,created_at"+
			"&order=created_at.asc",
		url.QueryEscape(parentID),
	)

	resp, err := c.DoWithRetry(func(token string) (*http.Response, error) {
		return c.Get(token, path)
	})
	if err != nil {
		return nil, fmt.Errorf("FetchRepliesByParentID: %w", err)
	}

	var replies []ReviewComment
	if err := ReadJSON(resp, &replies); err != nil {
		return nil, fmt.Errorf("FetchRepliesByParentID: %w", err)
	}

	return replies, nil
}

func (c *Client) FetchResolvedComments(changeID string) ([]ReviewComment, error) {
	path := fmt.Sprintf(
		"/rest/v1/review_comments?change_id=eq.%s&is_resolved=eq.true&parent_comment_id=is.null"+
			"&select=id,file_path,content,selected_text,line,start_line,author_name,author_email,created_at"+
			"&order=created_at.asc",
		url.QueryEscape(changeID),
	)

	resp, err := c.DoWithRetry(func(token string) (*http.Response, error) {
		return c.Get(token, path)
	})
	if err != nil {
		return nil, fmt.Errorf("FetchResolvedComments: %w", err)
	}

	var comments []ReviewComment
	if err := ReadJSON(resp, &comments); err != nil {
		return nil, fmt.Errorf("FetchResolvedComments: %w", err)
	}

	return comments, nil
}

func BuildReplyMap(replies []ReviewComment) ReplyMap {
	m := make(ReplyMap)
	for _, r := range replies {
		m[r.ParentCommentID] = append(m[r.ParentCommentID], r)
	}
	return m
}

func (c *Client) CreateReply(parentID, content string) (*ThreadReply, error) {
	reqURL := fmt.Sprintf("%s/rest/v1/review_comments", c.BaseURL)

	body := map[string]interface{}{
		"parent_comment_id": parentID,
		"content":           content,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("CreateReply: failed to marshal request: %w", err)
	}

	postFn := func(token string) (*http.Response, error) {
		req, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("apikey", c.AnonKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Prefer", "return=representation")

		return c.HTTPClient.Do(req)
	}

	resp, err := c.DoWithRetry(postFn)
	if err != nil {
		return nil, fmt.Errorf("CreateReply: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("CreateReply: parent comment not found: %s", parentID)
	}

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("CreateReply: API error (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	var replies []struct {
		ID        string `json:"id"`
		Content   string `json:"content"`
		CreatedAt string `json:"created_at"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&replies); err != nil {
		return nil, fmt.Errorf("CreateReply: failed to parse response: %w", err)
	}

	if len(replies) == 0 {
		return nil, fmt.Errorf("CreateReply: no reply returned from API")
	}

	return &ThreadReply{
		ID:        replies[0].ID,
		ParentID:  parentID,
		Content:   replies[0].Content,
		CreatedAt: replies[0].CreatedAt,
	}, nil
}

func (c *Client) ResolveComment(commentID string) error {
	reqURL := fmt.Sprintf("%s/rest/v1/review_comments?id=eq.%s", c.BaseURL, url.QueryEscape(commentID))

	body := []byte(`{"is_resolved":true}`)

	patchFn := func(token string) (*http.Response, error) {
		req, err := http.NewRequest(http.MethodPatch, reqURL, bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("apikey", c.AnonKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Prefer", "return=minimal")

		return c.HTTPClient.Do(req)
	}

	resp, err := c.DoWithRetry(patchFn)
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

func (c *Client) ResolveCommentWithReplies(commentID string, replyIDs []string) error {
	allIDs := make([]string, 0, 1+len(replyIDs))
	allIDs = append(allIDs, commentID)
	allIDs = append(allIDs, replyIDs...)
	idList := "(" + strings.Join(allIDs, ",") + ")"

	reqURL := fmt.Sprintf("%s/rest/v1/review_comments?id=in.%s", c.BaseURL, url.QueryEscape(idList))

	body := []byte(`{"is_resolved":true}`)

	patchFn := func(token string) (*http.Response, error) {
		req, err := http.NewRequest(http.MethodPatch, reqURL, bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("apikey", c.AnonKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Prefer", "return=minimal")

		return c.HTTPClient.Do(req)
	}

	resp, err := c.DoWithRetry(patchFn)
	if err != nil {
		return fmt.Errorf("ResolveCommentWithReplies: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ResolveCommentWithReplies: API error (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}
