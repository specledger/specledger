package skills

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
