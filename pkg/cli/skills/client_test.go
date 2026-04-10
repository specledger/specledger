package skills

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestClientSearch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/search" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		q := r.URL.Query().Get("q")
		if q != "commit" {
			t.Errorf("query = %q, want %q", q, "commit")
		}
		resp := searchResponse{
			Skills: []SkillSearchResult{
				{ID: "vercel-labs/agent-skills/creating-pr", Name: "creating-pr", Source: "vercel-labs/agent-skills", Installs: 12300},
				{ID: "vercel-labs/agent-skills/commit", Name: "commit", Source: "vercel-labs/agent-skills", Installs: 8100},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := &Client{
		SearchURL:  srv.URL,
		HTTPClient: srv.Client(),
	}

	results, err := c.Search("commit", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
	if results[0].Name != "creating-pr" {
		t.Errorf("results[0].Name = %q, want %q", results[0].Name, "creating-pr")
	}
	if results[0].Installs != 12300 {
		t.Errorf("results[0].Installs = %d, want %d", results[0].Installs, 12300)
	}
}

func TestClientSearch_NoResults(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(searchResponse{Skills: []SkillSearchResult{}})
	}))
	defer srv.Close()

	c := &Client{SearchURL: srv.URL, HTTPClient: srv.Client()}
	results, err := c.Search("xyznonexistent", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("len(results) = %d, want 0", len(results))
	}
}

func TestClientSearch_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := &Client{SearchURL: srv.URL, HTTPClient: srv.Client()}
	_, err := c.Search("test", 10)
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestClientFetchAudit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/audit" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		resp := map[string]SkillAuditResult{
			"creating-pr": {
				ATH:    &PartnerAudit{Risk: "safe", Alerts: 0, Score: 100},
				Socket: &PartnerAudit{Risk: "low", Alerts: 0, Score: 95},
				Snyk:   &PartnerAudit{Risk: "safe", Alerts: 0, Score: 100},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := &Client{AuditURL: srv.URL, HTTPClient: srv.Client()}
	results, err := c.FetchAudit("vercel-labs/agent-skills", []string{"creating-pr"})
	if err != nil {
		t.Fatalf("FetchAudit: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	audit, ok := results["creating-pr"]
	if !ok {
		t.Fatal("missing 'creating-pr' audit result")
	}
	if audit.ATH == nil || audit.ATH.Risk != "safe" {
		t.Errorf("ATH.Risk = %v, want safe", audit.ATH)
	}
	if audit.Socket == nil || audit.Socket.Score != 95 {
		t.Errorf("Socket.Score = %v, want 95", audit.Socket)
	}
}

func TestClientFetchSkillContent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		want := "/vercel-labs/agent-skills/main/skills/creating-pr/SKILL.md"
		if r.URL.Path != want {
			t.Errorf("path = %q, want %q", r.URL.Path, want)
		}
		_, _ = w.Write([]byte("---\nname: creating-pr\n---\nSkill content"))
	}))
	defer srv.Close()

	c := &Client{RawGHURL: srv.URL, HTTPClient: srv.Client()}
	data, err := c.FetchSkillContent("vercel-labs", "agent-skills", "main", "skills/creating-pr/SKILL.md")
	if err != nil {
		t.Fatalf("FetchSkillContent: %v", err)
	}
	if !contains(string(data), "creating-pr") {
		t.Errorf("content does not contain 'creating-pr': %s", string(data))
	}
}

func TestClientFetchSkillContent_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := &Client{RawGHURL: srv.URL, HTTPClient: srv.Client()}
	_, err := c.FetchSkillContent("owner", "repo", "main", "skills/nope/SKILL.md")
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if !contains(err.Error(), "not found") {
		t.Errorf("error = %q, want containing 'not found'", err.Error())
	}
}

func TestClientFetchRepoTree(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		want := "/repos/vercel-labs/agent-skills/git/trees/main"
		if r.URL.Path != want {
			t.Errorf("path = %q, want %q", r.URL.Path, want)
		}
		resp := githubTreeResponse{
			SHA: "abc123",
			Tree: []GitHubTreeEntry{
				{Path: "skills/creating-pr/SKILL.md", Type: "blob"},
				{Path: "skills/commit/SKILL.md", Type: "blob"},
				{Path: "README.md", Type: "blob"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := &Client{GitHubURL: srv.URL, HTTPClient: srv.Client()}
	tree, resolvedRef, err := c.FetchRepoTree("vercel-labs", "agent-skills", "main")
	if err != nil {
		t.Fatalf("FetchRepoTree: %v", err)
	}
	if len(tree) != 3 {
		t.Fatalf("len(tree) = %d, want 3", len(tree))
	}
	if resolvedRef != "main" {
		t.Errorf("resolvedRef = %q, want %q", resolvedRef, "main")
	}
}

func TestClientFetchRepoTree_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := &Client{GitHubURL: srv.URL, HTTPClient: srv.Client()}
	_, _, err := c.FetchRepoTree("nonexistent", "repo", "main")
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if !contains(err.Error(), "not found") {
		t.Errorf("error = %q, want containing 'not found'", err.Error())
	}
}

func TestFetchRepoTree_RefFallback(t *testing.T) {
	tests := []struct {
		name         string
		inputRef     string
		successPath  string // which path returns 200; empty = all 404
		wantRef      string
		wantErr      string
		wantRequests []string
		wantTree     int
	}{
		{
			name:         "explicit ref skips fallback",
			inputRef:     "dev",
			successPath:  "/repos/org/repo/git/trees/dev",
			wantRef:      "dev",
			wantRequests: []string{"/repos/org/repo/git/trees/dev"},
			wantTree:     1,
		},
		{
			name:        "empty ref falls back through HEAD, main, master",
			inputRef:    "",
			successPath: "/repos/org/repo/git/trees/master",
			wantRef:     "master",
			wantRequests: []string{
				"/repos/org/repo/git/trees/HEAD",
				"/repos/org/repo/git/trees/main",
				"/repos/org/repo/git/trees/master",
			},
			wantTree: 1,
		},
		{
			name:     "all fallbacks fail",
			inputRef: "",
			wantErr:  "could not resolve default branch",
			wantRequests: []string{
				"/repos/org/repo/git/trees/HEAD",
				"/repos/org/repo/git/trees/main",
				"/repos/org/repo/git/trees/master",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var requestedPaths []string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestedPaths = append(requestedPaths, r.URL.Path)
				if tt.successPath != "" && r.URL.Path == tt.successPath {
					resp := githubTreeResponse{
						SHA:  "abc",
						Tree: []GitHubTreeEntry{{Path: "skills/s1/SKILL.md", Type: "blob"}},
					}
					_ = json.NewEncoder(w).Encode(resp)
					return
				}
				w.WriteHeader(http.StatusNotFound)
			}))
			defer srv.Close()

			c := &Client{GitHubURL: srv.URL, HTTPClient: srv.Client()}
			tree, resolvedRef, err := c.FetchRepoTree("org", "repo", tt.inputRef)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("error = %q, want containing %q", err.Error(), tt.wantErr)
				}
			} else {
				if err != nil {
					t.Fatalf("FetchRepoTree: %v", err)
				}
				if len(tree) != tt.wantTree {
					t.Fatalf("len(tree) = %d, want %d", len(tree), tt.wantTree)
				}
				if resolvedRef != tt.wantRef {
					t.Errorf("resolvedRef = %q, want %q", resolvedRef, tt.wantRef)
				}
			}

			if len(requestedPaths) != len(tt.wantRequests) {
				t.Fatalf("requests = %v, want %v", requestedPaths, tt.wantRequests)
			}
			for i, p := range tt.wantRequests {
				if requestedPaths[i] != p {
					t.Errorf("request[%d] = %q, want %q", i, requestedPaths[i], p)
				}
			}
		})
	}
}

func TestNewClient_Defaults(t *testing.T) {
	// Clear any env overrides that might be set
	t.Setenv("SKILLS_API_URL", "")
	os.Unsetenv("SKILLS_API_URL")
	t.Setenv("SKILLS_AUDIT_URL", "")
	os.Unsetenv("SKILLS_AUDIT_URL")
	t.Setenv("GITHUB_API_URL", "")
	os.Unsetenv("GITHUB_API_URL")
	t.Setenv("GITHUB_RAW_URL", "")
	os.Unsetenv("GITHUB_RAW_URL")

	c := NewClient()
	if c.SearchURL != defaultSearchURL {
		t.Errorf("SearchURL = %q, want %q", c.SearchURL, defaultSearchURL)
	}
	if c.AuditURL != defaultAuditURL {
		t.Errorf("AuditURL = %q, want %q", c.AuditURL, defaultAuditURL)
	}
	if c.GitHubURL != defaultGitHubURL {
		t.Errorf("GitHubURL = %q, want %q", c.GitHubURL, defaultGitHubURL)
	}
	if c.RawGHURL != defaultRawGHURL {
		t.Errorf("RawGHURL = %q, want %q", c.RawGHURL, defaultRawGHURL)
	}
	if c.HTTPClient == nil {
		t.Error("HTTPClient is nil")
	}
}

func TestNewClient_EnvOverrides(t *testing.T) {
	t.Setenv("SKILLS_API_URL", "http://custom-search")
	t.Setenv("SKILLS_AUDIT_URL", "http://custom-audit")
	t.Setenv("GITHUB_API_URL", "http://custom-github")
	t.Setenv("GITHUB_RAW_URL", "http://custom-raw")

	c := NewClient()
	if c.SearchURL != "http://custom-search" {
		t.Errorf("SearchURL = %q, want %q", c.SearchURL, "http://custom-search")
	}
	if c.AuditURL != "http://custom-audit" {
		t.Errorf("AuditURL = %q, want %q", c.AuditURL, "http://custom-audit")
	}
	if c.GitHubURL != "http://custom-github" {
		t.Errorf("GitHubURL = %q, want %q", c.GitHubURL, "http://custom-github")
	}
	if c.RawGHURL != "http://custom-raw" {
		t.Errorf("RawGHURL = %q, want %q", c.RawGHURL, "http://custom-raw")
	}
}

func TestEnvOrDefault(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		envValue string
		setEnv   bool
		def      string
		want     string
	}{
		{
			name:     "env set returns env value",
			key:      "TEST_ENV_OR_DEFAULT_SET",
			envValue: "from-env",
			setEnv:   true,
			def:      "default-val",
			want:     "from-env",
		},
		{
			name:   "env not set returns default",
			key:    "TEST_ENV_OR_DEFAULT_UNSET",
			setEnv: false,
			def:    "default-val",
			want:   "default-val",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				t.Setenv(tt.key, tt.envValue)
			} else {
				t.Setenv(tt.key, "")
				os.Unsetenv(tt.key)
			}
			got := envOrDefault(tt.key, tt.def)
			if got != tt.want {
				t.Errorf("envOrDefault(%q, %q) = %q, want %q", tt.key, tt.def, got, tt.want)
			}
		})
	}
}

func TestGithubToken(t *testing.T) {
	tests := []struct {
		name        string
		githubToken string
		ghToken     string
		want        string
	}{
		{
			name:        "GITHUB_TOKEN only",
			githubToken: "gh-token-1",
			ghToken:     "",
			want:        "gh-token-1",
		},
		{
			name:        "GH_TOKEN only",
			githubToken: "",
			ghToken:     "gh-token-2",
			want:        "gh-token-2",
		},
		{
			name:        "both set GITHUB_TOKEN wins",
			githubToken: "primary",
			ghToken:     "secondary",
			want:        "primary",
		},
		{
			name:        "neither set returns empty",
			githubToken: "",
			ghToken:     "",
			want:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear both vars before each subtest
			t.Setenv("GITHUB_TOKEN", "")
			os.Unsetenv("GITHUB_TOKEN")
			t.Setenv("GH_TOKEN", "")
			os.Unsetenv("GH_TOKEN")

			if tt.githubToken != "" {
				t.Setenv("GITHUB_TOKEN", tt.githubToken)
			}
			if tt.ghToken != "" {
				t.Setenv("GH_TOKEN", tt.ghToken)
			}

			got := githubToken()
			if got != tt.want {
				t.Errorf("githubToken() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSearch_DefaultLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limit := r.URL.Query().Get("limit")
		if limit != "10" {
			t.Errorf("limit = %q, want %q", limit, "10")
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(searchResponse{Skills: []SkillSearchResult{}})
	}))
	defer srv.Close()

	c := &Client{SearchURL: srv.URL, HTTPClient: srv.Client()}
	_, err := c.Search("test", 0)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
}

func TestSearch_JSONDecodeError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{invalid`))
	}))
	defer srv.Close()

	c := &Client{SearchURL: srv.URL, HTTPClient: srv.Client()}
	_, err := c.Search("test", 10)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "invalid response") {
		t.Errorf("error = %q, want containing 'invalid response'", err.Error())
	}
}

func TestFetchAudit_NonOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := &Client{AuditURL: srv.URL, HTTPClient: srv.Client()}
	_, err := c.FetchAudit("owner/repo", []string{"skill-a"})
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestFetchAudit_JSONDecodeError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{invalid`))
	}))
	defer srv.Close()

	c := &Client{AuditURL: srv.URL, HTTPClient: srv.Client()}
	_, err := c.FetchAudit("owner/repo", []string{"skill-a"})
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "invalid response") {
		t.Errorf("error = %q, want containing 'invalid response'", err.Error())
	}
}

func TestFetchRepoTree_AuthHeader(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "test-token-abc")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token-abc" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer test-token-abc")
		}
		resp := githubTreeResponse{
			SHA:  "def456",
			Tree: []GitHubTreeEntry{},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := &Client{GitHubURL: srv.URL, HTTPClient: srv.Client()}
	_, _, err := c.FetchRepoTree("owner", "repo", "main")
	if err != nil {
		t.Fatalf("FetchRepoTree: %v", err)
	}
}

func TestFetchRepoTree_JSONDecodeError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{invalid`))
	}))
	defer srv.Close()

	c := &Client{GitHubURL: srv.URL, HTTPClient: srv.Client()}
	_, _, err := c.FetchRepoTree("owner", "repo", "main")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "invalid response") {
		t.Errorf("error = %q, want containing 'invalid response'", err.Error())
	}
}

func TestFetchSkillContent_Non200Non404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	c := &Client{RawGHURL: srv.URL, HTTPClient: srv.Client()}
	_, err := c.FetchSkillContent("owner", "repo", "main", "skills/test/SKILL.md")
	if err == nil {
		t.Fatal("expected error for 403 response")
	}
	if !strings.Contains(err.Error(), "failed to fetch") {
		t.Errorf("error = %q, want containing 'failed to fetch'", err.Error())
	}
}
