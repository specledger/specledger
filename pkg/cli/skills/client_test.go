package skills

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	tree, err := c.FetchRepoTree("vercel-labs", "agent-skills", "main")
	if err != nil {
		t.Fatalf("FetchRepoTree: %v", err)
	}
	if len(tree) != 3 {
		t.Fatalf("len(tree) = %d, want 3", len(tree))
	}
}

func TestClientFetchRepoTree_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := &Client{GitHubURL: srv.URL, HTTPClient: srv.Client()}
	_, err := c.FetchRepoTree("nonexistent", "repo", "main")
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if !contains(err.Error(), "not found") {
		t.Errorf("error = %q, want containing 'not found'", err.Error())
	}
}
