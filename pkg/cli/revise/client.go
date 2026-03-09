package revise

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/specledger/specledger/pkg/cli/comment"
)

type SpecWithCommentCount struct {
	SpecKey      string
	CommentCount int
}

type ReviseClient struct {
	*comment.Client
}

func NewReviseClient(accessToken string) *ReviseClient {
	return &ReviseClient{
		Client: comment.NewClient(accessToken),
	}
}

func (c *ReviseClient) ResolveComment(commentID string) error {
	reqURL := fmt.Sprintf(
		"%s/rest/v1/review_comments?id=eq.%s",
		c.BaseURL,
		url.QueryEscape(commentID),
	)

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

func (c *ReviseClient) ResolveCommentWithReplies(commentID string, replyIDs []string) error {
	allIDs := make([]string, 0, 1+len(replyIDs))
	allIDs = append(allIDs, commentID)
	allIDs = append(allIDs, replyIDs...)
	idList := "(" + strings.Join(allIDs, ",") + ")"

	reqURL := fmt.Sprintf(
		"%s/rest/v1/review_comments?id=in.%s",
		c.BaseURL,
		url.QueryEscape(idList),
	)

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

func (c *ReviseClient) ListSpecsWithComments(projectID string) ([]SpecWithCommentCount, error) {
	specsPath := fmt.Sprintf(
		"/rest/v1/specs?project_id=eq.%s&select=id,spec_key",
		url.QueryEscape(projectID),
	)

	resp, err := c.DoWithRetry(func(token string) (*http.Response, error) {
		return c.Get(token, specsPath)
	})
	if err != nil {
		return nil, fmt.Errorf("ListSpecsWithComments: fetch specs: %w", err)
	}

	type apiSpec struct {
		ID      string `json:"id"`
		SpecKey string `json:"spec_key"`
	}

	var specs []apiSpec
	if err := comment.ReadJSON(resp, &specs); err != nil {
		return nil, fmt.Errorf("ListSpecsWithComments: parse specs: %w", err)
	}

	if len(specs) == 0 {
		return nil, nil
	}

	specIDToKey := make(map[string]string, len(specs))
	specIDs := make([]string, 0, len(specs))
	for _, s := range specs {
		specIDToKey[s.ID] = s.SpecKey
		specIDs = append(specIDs, s.ID)
	}
	specIDList := "(" + strings.Join(specIDs, ",") + ")"

	changesPath := fmt.Sprintf(
		"/rest/v1/changes?spec_id=in.%s&select=id,spec_id",
		url.QueryEscape(specIDList),
	)

	resp, err = c.DoWithRetry(func(token string) (*http.Response, error) {
		return c.Get(token, changesPath)
	})
	if err != nil {
		return nil, fmt.Errorf("ListSpecsWithComments: fetch changes: %w", err)
	}

	type apiChange struct {
		ID     string `json:"id"`
		SpecID string `json:"spec_id"`
	}

	var changes []apiChange
	if err := comment.ReadJSON(resp, &changes); err != nil {
		return nil, fmt.Errorf("ListSpecsWithComments: parse changes: %w", err)
	}

	if len(changes) == 0 {
		return nil, nil
	}

	changeIDToSpecKey := make(map[string]string, len(changes))
	changeIDs := make([]string, 0, len(changes))
	for _, ch := range changes {
		changeIDToSpecKey[ch.ID] = specIDToKey[ch.SpecID]
		changeIDs = append(changeIDs, ch.ID)
	}
	changeIDList := "(" + strings.Join(changeIDs, ",") + ")"

	commentsPath := fmt.Sprintf(
		"/rest/v1/review_comments?change_id=in.%s&is_resolved=eq.false&parent_comment_id=is.null&select=id,change_id",
		url.QueryEscape(changeIDList),
	)

	resp, err = c.DoWithRetry(func(token string) (*http.Response, error) {
		return c.Get(token, commentsPath)
	})
	if err != nil {
		return nil, fmt.Errorf("ListSpecsWithComments: fetch comments: %w", err)
	}

	type commentRow struct {
		ID       string `json:"id"`
		ChangeID string `json:"change_id"`
	}

	var commentRows []commentRow
	if err := comment.ReadJSON(resp, &commentRows); err != nil {
		return nil, fmt.Errorf("ListSpecsWithComments: parse comments: %w", err)
	}

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
