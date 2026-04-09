package skills

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	defaultSearchURL = "https://skills.sh"
	defaultAuditURL  = "https://add-skill.vercel.sh"
	defaultGitHubURL = "https://api.github.com"
	defaultRawGHURL  = "https://raw.githubusercontent.com"

	clientTimeout = 10 * time.Second
)

// Client is an HTTP client for the skills.sh and GitHub APIs.
type Client struct {
	SearchURL  string // skills.sh base URL
	AuditURL   string // add-skill.vercel.sh base URL
	GitHubURL  string // api.github.com base URL
	RawGHURL   string // raw.githubusercontent.com base URL
	HTTPClient *http.Client
}

// NewClient creates a Client with base URLs from env vars (for testability).
func NewClient() *Client {
	return &Client{
		SearchURL:  envOrDefault("SKILLS_API_URL", defaultSearchURL),
		AuditURL:   envOrDefault("SKILLS_AUDIT_URL", defaultAuditURL),
		GitHubURL:  envOrDefault("GITHUB_API_URL", defaultGitHubURL),
		RawGHURL:   envOrDefault("GITHUB_RAW_URL", defaultRawGHURL),
		HTTPClient: &http.Client{Timeout: clientTimeout},
	}
}

// SkillSearchResult is a single result from the search API.
type SkillSearchResult struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Source   string `json:"source"`
	Installs int    `json:"installs"`
}

type searchResponse struct {
	Skills []SkillSearchResult `json:"skills"`
}

// Search queries the skills.sh search API.
func (c *Client) Search(query string, limit int) ([]SkillSearchResult, error) {
	if limit <= 0 {
		limit = 10
	}

	reqURL := fmt.Sprintf("%s/api/search?q=%s&limit=%d",
		c.SearchURL,
		url.QueryEscape(query),
		limit,
	)

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("sl skill search failed: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sl skill search failed: skills.sh API unreachable\n→ Check your internet connection and try again.\n→ skills.sh status: https://skills.sh")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sl skill search failed (%d): unexpected response from skills.sh\n→ Try again later or check https://skills.sh", resp.StatusCode)
	}

	var result searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("sl skill search failed: invalid response from skills.sh: %w", err)
	}

	return result.Skills, nil
}

// PartnerAudit is security audit data for a single partner.
type PartnerAudit struct {
	Risk       string    `json:"risk"`
	Alerts     int       `json:"alerts"`
	Score      int       `json:"score"`
	AnalyzedAt time.Time `json:"analyzedAt"`
}

// SkillAuditResult is audit results for a single skill across all partners.
type SkillAuditResult struct {
	Slug   string        `json:"slug,omitempty"`
	ATH    *PartnerAudit `json:"ath,omitempty"`
	Socket *PartnerAudit `json:"socket,omitempty"`
	Snyk   *PartnerAudit `json:"snyk,omitempty"`
}

// FetchAudit fetches security audit data for one or more skills.
func (c *Client) FetchAudit(source string, slugs []string) (map[string]*SkillAuditResult, error) {
	reqURL := fmt.Sprintf("%s/audit?source=%s&skills=%s",
		c.AuditURL,
		url.QueryEscape(source),
		url.QueryEscape(strings.Join(slugs, ",")),
	)

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("audit request failed: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("audit API unreachable\n→ Check your internet connection and try again")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("audit failed (%d): unexpected response\n→ Try again later", resp.StatusCode)
	}

	// The API returns map[slug]{ ath: {}, socket: {}, snyk: {} }
	var raw map[string]json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("audit failed: invalid response: %w", err)
	}

	results := make(map[string]*SkillAuditResult, len(raw))
	for slug, data := range raw {
		var result SkillAuditResult
		if err := json.Unmarshal(data, &result); err != nil {
			continue
		}
		result.Slug = slug
		results[slug] = &result
	}

	return results, nil
}

// FetchSkillContent fetches the raw SKILL.md content from GitHub.
func (c *Client) FetchSkillContent(owner, repo, ref, skillPath string) ([]byte, error) {
	reqURL := fmt.Sprintf("%s/%s/%s/%s/%s",
		c.RawGHURL, owner, repo, ref, skillPath)

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch skill content: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch skill content: GitHub unreachable\n→ Check your internet connection and try again")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("skill not found at %s/%s/%s\n→ Verify the repository and skill path are correct", owner, repo, skillPath)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch skill content (%d)\n→ Try again later", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read skill content: %w", err)
	}

	return data, nil
}

// GitHubTreeEntry is a single entry from the GitHub Trees API.
type GitHubTreeEntry struct {
	Path string `json:"path"`
	Mode string `json:"mode"`
	Type string `json:"type"`
	SHA  string `json:"sha"`
	Size int    `json:"size"`
}

type githubTreeResponse struct {
	SHA  string            `json:"sha"`
	Tree []GitHubTreeEntry `json:"tree"`
}

// defaultRefFallbacks is the ordered list of refs to try when no explicit ref is given.
// Matches the skills.sh TS CLI behavior (blob.ts).
var defaultRefFallbacks = []string{"HEAD", "main", "master"}

// FetchRepoTree fetches the full recursive tree for a repository.
// When ref is empty, it tries HEAD, main, master in order (matching skills.sh behavior).
// Returns the tree entries and the ref that succeeded.
func (c *Client) FetchRepoTree(owner, repo, ref string) ([]GitHubTreeEntry, string, error) {
	refs := []string{ref}
	if ref == "" {
		refs = defaultRefFallbacks
	}

	var lastErr error
	for _, r := range refs {
		tree, err := c.fetchRepoTreeOnce(owner, repo, r)
		if err == nil {
			return tree, r, nil
		}
		lastErr = err
	}

	// When auto-resolving, clarify that the repo may exist but we couldn't find the branch
	if ref == "" {
		return nil, "", fmt.Errorf("could not resolve default branch for %s/%s (tried HEAD, main, master)\n→ Specify a branch: sl skill add owner/repo#branch-name\n→ Or verify the repository is public and accessible", owner, repo)
	}
	return nil, "", lastErr
}

func (c *Client) fetchRepoTreeOnce(owner, repo, ref string) ([]GitHubTreeEntry, error) {
	reqURL := fmt.Sprintf("%s/repos/%s/%s/git/trees/%s?recursive=1",
		c.GitHubURL, owner, repo, ref)

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("GitHub Trees API request failed: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	// Use GitHub token if available for rate limit headroom
	if token := githubToken(); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GitHub API unreachable\n→ Check your internet connection and try again")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("repository %q not found (404)\n→ Verify the repository exists and is public.\n→ For private repos, set GITHUB_TOKEN.", owner+"/"+repo)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub Trees API failed (%d)\n→ Try again later. If rate-limited, set GITHUB_TOKEN.", resp.StatusCode)
	}

	var result githubTreeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("GitHub Trees API: invalid response: %w", err)
	}

	return result.Tree, nil
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func githubToken() string {
	if t := os.Getenv("GITHUB_TOKEN"); t != "" {
		return t
	}
	return os.Getenv("GH_TOKEN")
}
