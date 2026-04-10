package skills

import (
	"testing"
)

func TestParseSource(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantOwner   string
		wantRepo    string
		wantFilter  string
		wantType    SourceType
		wantURL     string
		wantErr     bool
		errContains string
	}{
		// Shorthand: owner/repo
		{
			name:      "simple owner/repo",
			input:     "vercel-labs/agent-skills",
			wantOwner: "vercel-labs",
			wantRepo:  "agent-skills",
			wantType:  SourceTypeGitHub,
		},
		{
			name:      "owner/repo with whitespace",
			input:     "  vercel-labs/agent-skills  ",
			wantOwner: "vercel-labs",
			wantRepo:  "agent-skills",
			wantType:  SourceTypeGitHub,
		},
		// Shorthand: owner/repo@skill
		{
			name:       "owner/repo@skill",
			input:      "vercel-labs/agent-skills@creating-pr",
			wantOwner:  "vercel-labs",
			wantRepo:   "agent-skills",
			wantFilter: "creating-pr",
			wantType:   SourceTypeGitHub,
		},
		{
			name:       "owner/repo@skill with dots",
			input:      "org/repo@my.skill",
			wantOwner:  "org",
			wantRepo:   "repo",
			wantFilter: "my.skill",
			wantType:   SourceTypeGitHub,
		},
		// HTTPS URLs
		{
			name:      "HTTPS GitHub URL",
			input:     "https://github.com/vercel-labs/agent-skills.git",
			wantOwner: "vercel-labs",
			wantRepo:  "agent-skills",
			wantType:  SourceTypeGit,
			wantURL:   "https://github.com/vercel-labs/agent-skills.git",
		},
		{
			name:      "HTTPS URL without .git",
			input:     "https://github.com/vercel-labs/agent-skills",
			wantOwner: "vercel-labs",
			wantRepo:  "agent-skills",
			wantType:  SourceTypeGit,
			wantURL:   "https://github.com/vercel-labs/agent-skills",
		},
		// SSH URLs
		{
			name:      "SSH URL",
			input:     "git@github.com:vercel-labs/agent-skills.git",
			wantOwner: "vercel-labs",
			wantRepo:  "agent-skills",
			wantType:  SourceTypeGit,
			wantURL:   "git@github.com:vercel-labs/agent-skills.git",
		},
		{
			name:      "SSH protocol URL",
			input:     "ssh://git@github.com/vercel-labs/agent-skills.git",
			wantOwner: "vercel-labs",
			wantRepo:  "agent-skills",
			wantType:  SourceTypeGit,
			wantURL:   "ssh://git@github.com/vercel-labs/agent-skills.git",
		},
		// Error cases
		{
			name:        "empty input",
			input:       "",
			wantErr:     true,
			errContains: "empty source",
		},
		{
			name:        "whitespace only",
			input:       "   ",
			wantErr:     true,
			errContains: "empty source",
		},
		{
			name:        "no slash",
			input:       "just-a-name",
			wantErr:     true,
			errContains: "invalid source",
		},
		{
			name:        "path traversal",
			input:       "../evil/repo",
			wantErr:     true,
			errContains: "path traversal",
		},
		{
			name:        "empty skill filter",
			input:       "owner/repo@",
			wantErr:     true,
			errContains: "empty skill name",
		},
		{
			name:        "skill filter with slash",
			input:       "owner/repo@bad/skill",
			wantErr:     true,
			errContains: "must not contain",
		},
		{
			name:        "triple segments",
			input:       "owner/repo/extra",
			wantErr:     true,
			errContains: "invalid source",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSource(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseSource(%q) = %+v, want error containing %q", tt.input, got, tt.errContains)
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("ParseSource(%q) error = %q, want containing %q", tt.input, err.Error(), tt.errContains)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseSource(%q) unexpected error: %v", tt.input, err)
			}
			if got.Owner != tt.wantOwner {
				t.Errorf("Owner = %q, want %q", got.Owner, tt.wantOwner)
			}
			if got.Repo != tt.wantRepo {
				t.Errorf("Repo = %q, want %q", got.Repo, tt.wantRepo)
			}
			if got.SkillFilter != tt.wantFilter {
				t.Errorf("SkillFilter = %q, want %q", got.SkillFilter, tt.wantFilter)
			}
			if got.Type != tt.wantType {
				t.Errorf("Type = %q, want %q", got.Type, tt.wantType)
			}
			if tt.wantURL != "" && got.URL != tt.wantURL {
				t.Errorf("URL = %q, want %q", got.URL, tt.wantURL)
			}
			if got.Ref != "" {
				t.Errorf("Ref = %q, want empty (auto-resolve)", got.Ref)
			}
		})
	}
}

func TestSkillSourceSourceString(t *testing.T) {
	s := &SkillSource{Owner: "vercel-labs", Repo: "agent-skills"}
	if got := s.SourceString(); got != "vercel-labs/agent-skills" {
		t.Errorf("SourceString() = %q, want %q", got, "vercel-labs/agent-skills")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
