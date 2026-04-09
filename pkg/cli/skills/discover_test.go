package skills

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestDiscoverSkills_GitHubFastPath(t *testing.T) {
	mux := http.NewServeMux()

	// Trees API
	mux.HandleFunc("/repos/vercel-labs/agent-skills/git/trees/main", func(w http.ResponseWriter, _ *http.Request) {
		resp := githubTreeResponse{
			SHA: "abc",
			Tree: []GitHubTreeEntry{
				{Path: "skills/creating-pr/SKILL.md", Type: "blob"},
				{Path: "skills/commit/SKILL.md", Type: "blob"},
				{Path: "README.md", Type: "blob"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})

	// Raw content
	mux.HandleFunc("/vercel-labs/agent-skills/main/skills/creating-pr/SKILL.md", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("---\nname: creating-pr\ndescription: Create PRs\n---\nContent"))
	})
	mux.HandleFunc("/vercel-labs/agent-skills/main/skills/commit/SKILL.md", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("---\nname: commit\ndescription: Commit changes\n---\nContent"))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := &Client{
		GitHubURL:  srv.URL,
		RawGHURL:   srv.URL,
		HTTPClient: srv.Client(),
	}

	source := &SkillSource{
		Owner: "vercel-labs",
		Repo:  "agent-skills",
		Ref:   "main",
		Type:  SourceTypeGitHub,
	}

	skills, err := DiscoverSkills(client, source)
	if err != nil {
		t.Fatalf("DiscoverSkills: %v", err)
	}
	if len(skills) != 2 {
		t.Fatalf("len(skills) = %d, want 2", len(skills))
	}

	names := map[string]bool{}
	for _, s := range skills {
		names[s.Name] = true
		if s.Source != "vercel-labs/agent-skills" {
			t.Errorf("Source = %q, want %q", s.Source, "vercel-labs/agent-skills")
		}
	}
	if !names["creating-pr"] || !names["commit"] {
		t.Errorf("unexpected skills: %v", names)
	}
}

func TestDiscoverSkills_WithFilter(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/org/repo/git/trees/main", func(w http.ResponseWriter, _ *http.Request) {
		resp := githubTreeResponse{
			Tree: []GitHubTreeEntry{
				{Path: "skills/alpha/SKILL.md", Type: "blob"},
				{Path: "skills/beta/SKILL.md", Type: "blob"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/org/repo/main/skills/alpha/SKILL.md", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("---\nname: alpha\ndescription: Alpha skill\n---\n"))
	})
	mux.HandleFunc("/org/repo/main/skills/beta/SKILL.md", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("---\nname: beta\ndescription: Beta skill\n---\n"))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := &Client{GitHubURL: srv.URL, RawGHURL: srv.URL, HTTPClient: srv.Client()}
	source := &SkillSource{Owner: "org", Repo: "repo", Ref: "main", Type: SourceTypeGitHub, SkillFilter: "alpha"}

	skills, err := DiscoverSkills(client, source)
	if err != nil {
		t.Fatalf("DiscoverSkills: %v", err)
	}
	if len(skills) != 1 {
		t.Fatalf("len(skills) = %d, want 1", len(skills))
	}
	if skills[0].Name != "alpha" {
		t.Errorf("Name = %q, want %q", skills[0].Name, "alpha")
	}
}

func TestParseSkillFrontmatter(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantName    string
		wantDesc    string
		wantErr     bool
		errContains string
	}{
		{
			name:     "valid frontmatter",
			content:  "---\nname: creating-pr\ndescription: Create PRs\n---\nBody content",
			wantName: "creating-pr",
			wantDesc: "Create PRs",
		},
		{
			name:        "missing frontmatter",
			content:     "No frontmatter here",
			wantErr:     true,
			errContains: "missing frontmatter",
		},
		{
			name:        "unterminated frontmatter",
			content:     "---\nname: test\n",
			wantErr:     true,
			errContains: "not terminated",
		},
		{
			name:        "missing name",
			content:     "---\ndescription: Test\n---\n",
			wantErr:     true,
			errContains: "missing required 'name'",
		},
		{
			name:        "missing description",
			content:     "---\nname: test\n---\n",
			wantErr:     true,
			errContains: "missing required 'description'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta, err := ParseSkillFrontmatter([]byte(tt.content))
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("error = %q, want containing %q", err.Error(), tt.errContains)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if meta.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", meta.Name, tt.wantName)
			}
			if meta.Description != tt.wantDesc {
				t.Errorf("Description = %q, want %q", meta.Description, tt.wantDesc)
			}
		})
	}
}

func TestDiscoverViaGitHub_NoSkillMDFiles(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/org/repo/git/trees/main", func(w http.ResponseWriter, _ *http.Request) {
		resp := githubTreeResponse{
			SHA: "abc",
			Tree: []GitHubTreeEntry{
				{Path: "README.md", Type: "blob"},
				{Path: "src/main.go", Type: "blob"},
				{Path: "docs", Type: "tree"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := &Client{GitHubURL: srv.URL, RawGHURL: srv.URL, HTTPClient: srv.Client()}
	source := &SkillSource{Owner: "org", Repo: "repo", Ref: "main", Type: SourceTypeGitHub}

	_, err := discoverViaGitHub(client, source)
	if err == nil {
		t.Fatal("expected error")
	}
	if !contains(err.Error(), "no skills found") {
		t.Errorf("error = %q, want containing %q", err.Error(), "no skills found")
	}
}

func TestDiscoverViaGitHub_AllFetchesFail(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/org/repo/git/trees/main", func(w http.ResponseWriter, _ *http.Request) {
		resp := githubTreeResponse{
			SHA: "abc",
			Tree: []GitHubTreeEntry{
				{Path: "skills/alpha/SKILL.md", Type: "blob"},
				{Path: "skills/beta/SKILL.md", Type: "blob"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	// All content fetches return 500
	mux.HandleFunc("/org/repo/main/skills/alpha/SKILL.md", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	mux.HandleFunc("/org/repo/main/skills/beta/SKILL.md", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := &Client{GitHubURL: srv.URL, RawGHURL: srv.URL, HTTPClient: srv.Client()}
	source := &SkillSource{Owner: "org", Repo: "repo", Ref: "main", Type: SourceTypeGitHub}

	_, err := discoverViaGitHub(client, source)
	if err == nil {
		t.Fatal("expected error")
	}
	if !contains(err.Error(), "no valid skills found") {
		t.Errorf("error = %q, want containing %q", err.Error(), "no valid skills found")
	}
}

func TestDiscoverViaGitHub_FilterNotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/org/repo/git/trees/main", func(w http.ResponseWriter, _ *http.Request) {
		resp := githubTreeResponse{
			SHA: "abc",
			Tree: []GitHubTreeEntry{
				{Path: "skills/alpha/SKILL.md", Type: "blob"},
				{Path: "skills/beta/SKILL.md", Type: "blob"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/org/repo/main/skills/alpha/SKILL.md", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("---\nname: alpha\ndescription: Alpha skill\n---\n"))
	})
	mux.HandleFunc("/org/repo/main/skills/beta/SKILL.md", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("---\nname: beta\ndescription: Beta skill\n---\n"))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := &Client{GitHubURL: srv.URL, RawGHURL: srv.URL, HTTPClient: srv.Client()}
	source := &SkillSource{Owner: "org", Repo: "repo", Ref: "main", Type: SourceTypeGitHub, SkillFilter: "nonexistent"}

	_, err := discoverViaGitHub(client, source)
	if err == nil {
		t.Fatal("expected error")
	}
	if !contains(err.Error(), "not found") {
		t.Errorf("error = %q, want containing %q", err.Error(), "not found")
	}
	if !contains(err.Error(), "nonexistent") {
		t.Errorf("error = %q, want containing skill name %q", err.Error(), "nonexistent")
	}
}

func TestDiscoverViaGitHub_PartialFetchFailure(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/org/repo/git/trees/main", func(w http.ResponseWriter, _ *http.Request) {
		resp := githubTreeResponse{
			SHA: "abc",
			Tree: []GitHubTreeEntry{
				{Path: "skills/alpha/SKILL.md", Type: "blob"},
				{Path: "skills/beta/SKILL.md", Type: "blob"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	// alpha fetch fails
	mux.HandleFunc("/org/repo/main/skills/alpha/SKILL.md", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	// beta fetch succeeds
	mux.HandleFunc("/org/repo/main/skills/beta/SKILL.md", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("---\nname: beta\ndescription: Beta skill\n---\n"))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := &Client{GitHubURL: srv.URL, RawGHURL: srv.URL, HTTPClient: srv.Client()}
	source := &SkillSource{Owner: "org", Repo: "repo", Ref: "main", Type: SourceTypeGitHub}

	skills, err := discoverViaGitHub(client, source)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(skills) != 1 {
		t.Fatalf("len(skills) = %d, want 1", len(skills))
	}
	if skills[0].Name != "beta" {
		t.Errorf("Name = %q, want %q", skills[0].Name, "beta")
	}
}

func TestDiscoverSkills_GitHubFallbackToClone(t *testing.T) {
	// This test verifies that when GitHub Trees API fails, DiscoverSkills falls back
	// to git clone. The fallback path constructs a hardcoded https://github.com/...
	// clone URL that cannot be intercepted, so this test requires real network access
	// and git authentication (or a public repo that 404s quickly).
	//
	// Behavioral coverage: this path is covered by E2E integration tests
	// (tests/integration/skills_test.go) which use httptest to mock all endpoints.
	//
	// To run manually: DISCOVER_CLONE_TEST=1 go test ./pkg/cli/skills/ -run TestDiscoverSkills_GitHubFallbackToClone -timeout 60s
	if os.Getenv("DISCOVER_CLONE_TEST") == "" {
		t.Skip("skipped: exercises real git clone (network). Set DISCOVER_CLONE_TEST=1 to run.")
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	client := &Client{GitHubURL: srv.URL, RawGHURL: srv.URL, HTTPClient: srv.Client()}
	source := &SkillSource{
		Owner: "x",
		Repo:  "y",
		Ref:   "main",
		Type:  SourceTypeGitHub,
	}

	_, err := DiscoverSkills(client, source)
	if err == nil {
		t.Fatal("expected error")
	}
	if !contains(err.Error(), "could not clone") {
		t.Errorf("error = %q, want containing 'could not clone' (fallback path)", err.Error())
	}
}

func TestDiscoverSkills_GitSourceType(t *testing.T) {
	// SourceTypeGit should go directly to clone path, skipping GitHub API entirely.
	// Uses file:// protocol pointing to a nonexistent path for fast failure (no network).
	source := &SkillSource{
		Owner: "org",
		Repo:  "repo",
		Ref:   "main",
		Type:  SourceTypeGit,
		URL:   "file:///nonexistent/path/to/repo.git",
	}

	// Client is irrelevant — SourceTypeGit bypasses GitHub API
	client := &Client{}

	_, err := DiscoverSkills(client, source)
	if err == nil {
		t.Fatal("expected error")
	}
	// Should fail with clone error, not a nil pointer from client usage
	if !contains(err.Error(), "could not clone") {
		t.Errorf("error = %q, want containing 'could not clone'", err.Error())
	}
}

func TestParseSkillFrontmatter_UnsafeName(t *testing.T) {
	content := "---\nname: ../etc/passwd\ndescription: Malicious skill\n---\n"
	_, err := ParseSkillFrontmatter([]byte(content))
	if err == nil {
		t.Fatal("expected error")
	}
	if !contains(err.Error(), "unsafe name") {
		t.Errorf("error = %q, want containing %q", err.Error(), "unsafe name")
	}
}

func TestParseSkillFrontmatter_InvalidYAML(t *testing.T) {
	content := "---\n: invalid: yaml\n---\n"
	_, err := ParseSkillFrontmatter([]byte(content))
	if err == nil {
		t.Fatal("expected error")
	}
	if !contains(err.Error(), "invalid") {
		t.Errorf("error = %q, want containing %q", err.Error(), "invalid")
	}
}
