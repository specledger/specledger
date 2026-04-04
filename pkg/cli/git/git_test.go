package git

import (
	"testing"
)

func TestParseRepoURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantOwner string
		wantName  string
		wantErr   bool
	}{
		{
			name:      "standard SSH",
			url:       "git@github.com:owner/repo.git",
			wantOwner: "owner",
			wantName:  "repo",
		},
		{
			name:      "SSH without .git suffix",
			url:       "git@github.com:owner/repo",
			wantOwner: "owner",
			wantName:  "repo",
		},
		{
			name:      "standard HTTPS",
			url:       "https://github.com/owner/repo.git",
			wantOwner: "owner",
			wantName:  "repo",
		},
		{
			name:      "HTTPS without .git suffix",
			url:       "https://github.com/owner/repo",
			wantOwner: "owner",
			wantName:  "repo",
		},
		{
			name:      "SSH custom alias without .com",
			url:       "git@github-so0k:so0k/tfc-cli.git",
			wantOwner: "so0k",
			wantName:  "tfc-cli",
		},
		{
			name:      "SSH custom alias with .com suffix",
			url:       "git@github.com-so0k:so0k/tfc-cli.git",
			wantOwner: "so0k",
			wantName:  "tfc-cli",
		},
		{
			name:      "SSH arbitrary hostname",
			url:       "git@my-custom-host:org/project.git",
			wantOwner: "org",
			wantName:  "project",
		},
		{
			name:      "HTTPS GitLab",
			url:       "https://gitlab.com/myorg/myrepo.git",
			wantOwner: "myorg",
			wantName:  "myrepo",
		},
		{
			name:    "empty string",
			url:     "",
			wantErr: true,
		},
		{
			name:    "garbage",
			url:     "not-a-url",
			wantErr: true,
		},
		{
			name:    "missing repo path",
			url:     "git@github.com:owner",
			wantErr: true,
		},
		{
			name:      "ssh:// protocol standard",
			url:       "ssh://git@github.com/owner/repo.git",
			wantOwner: "owner",
			wantName:  "repo",
		},
		{
			name:      "ssh:// protocol custom host",
			url:       "ssh://git@my-host/org/project.git",
			wantOwner: "org",
			wantName:  "project",
		},
		{
			name:      "ssh:// protocol with port",
			url:       "ssh://git@github.com:22/owner/repo.git",
			wantOwner: "owner",
			wantName:  "repo",
		},
		{
			name:      "HTTPS with port",
			url:       "https://git.corp:8443/owner/repo.git",
			wantOwner: "owner",
			wantName:  "repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, name, err := ParseRepoURL(tt.url)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseRepoURL(%q) expected error, got owner=%q name=%q", tt.url, owner, name)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseRepoURL(%q) unexpected error: %v", tt.url, err)
			}
			if owner != tt.wantOwner {
				t.Errorf("ParseRepoURL(%q) owner = %q, want %q", tt.url, owner, tt.wantOwner)
			}
			if name != tt.wantName {
				t.Errorf("ParseRepoURL(%q) name = %q, want %q", tt.url, name, tt.wantName)
			}
		})
	}
}

func TestParseRepoFlag(t *testing.T) {
	tests := []struct {
		name      string
		flag      string
		wantOwner string
		wantName  string
		wantErr   bool
	}{
		{
			name:      "valid owner/repo",
			flag:      "owner/repo",
			wantOwner: "owner",
			wantName:  "repo",
		},
		{
			name:    "missing repo",
			flag:    "owner/",
			wantErr: true,
		},
		{
			name:    "missing owner",
			flag:    "/repo",
			wantErr: true,
		},
		{
			name:    "no slash",
			flag:    "owner",
			wantErr: true,
		},
		{
			name:    "empty string",
			flag:    "",
			wantErr: true,
		},
		{
			name:    "extra slash rejected",
			flag:    "owner/repo/extra",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, name, err := ParseRepoFlag(tt.flag)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseRepoFlag(%q) expected error, got owner=%q name=%q", tt.flag, owner, name)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseRepoFlag(%q) unexpected error: %v", tt.flag, err)
			}
			if owner != tt.wantOwner {
				t.Errorf("ParseRepoFlag(%q) owner = %q, want %q", tt.flag, owner, tt.wantOwner)
			}
			if name != tt.wantName {
				t.Errorf("ParseRepoFlag(%q) name = %q, want %q", tt.flag, name, tt.wantName)
			}
		})
	}
}
