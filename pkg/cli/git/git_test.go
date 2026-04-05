package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// gitCmd runs a git command in dir, failing the test on error.
func gitCmd(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\nOutput: %s", args, err, string(output))
	}
}

// initTestRepo creates a git repo with an initial commit in dir.
func initTestRepo(t *testing.T, dir string) {
	t.Helper()
	gitCmd(t, dir, "init")
	gitCmd(t, dir, "config", "user.email", "test@test.com")
	gitCmd(t, dir, "config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(dir, ".gitkeep"), nil, 0644); err != nil {
		t.Fatal(err)
	}
	gitCmd(t, dir, "add", "-A")
	gitCmd(t, dir, "commit", "-m", "init")
}

func TestGetCurrentBranch_Worktree(t *testing.T) {
	mainDir := t.TempDir()
	initTestRepo(t, mainDir)

	worktreePath := filepath.Join(t.TempDir(), "wt")
	gitCmd(t, mainDir, "worktree", "add", worktreePath, "-b", "test-worktree-branch")

	branch, err := GetCurrentBranch(worktreePath)
	if err != nil {
		t.Fatalf("GetCurrentBranch in worktree: %v", err)
	}
	if branch != "test-worktree-branch" {
		t.Errorf("got branch %q, want %q", branch, "test-worktree-branch")
	}
}

func TestHasUncommittedChanges_Worktree(t *testing.T) {
	mainDir := t.TempDir()
	initTestRepo(t, mainDir)

	worktreePath := filepath.Join(t.TempDir(), "wt")
	gitCmd(t, mainDir, "worktree", "add", worktreePath, "-b", "wt-clean")

	dirty, err := HasUncommittedChanges(worktreePath)
	if err != nil {
		t.Fatalf("HasUncommittedChanges in worktree: %v", err)
	}
	if dirty {
		t.Error("expected clean worktree, got dirty")
	}

	// Create an untracked file to make it dirty
	if err := os.WriteFile(filepath.Join(worktreePath, "new.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	dirty, err = HasUncommittedChanges(worktreePath)
	if err != nil {
		t.Fatalf("HasUncommittedChanges in dirty worktree: %v", err)
	}
	if !dirty {
		t.Error("expected dirty worktree after adding file")
	}
}

func TestGetCurrentBranch_NormalRepo(t *testing.T) {
	dir := t.TempDir()
	initTestRepo(t, dir)

	branch, err := GetCurrentBranch(dir)
	if err != nil {
		t.Fatalf("GetCurrentBranch: %v", err)
	}
	// git init creates "main" or "master" depending on git config
	if branch != "main" && branch != "master" {
		t.Errorf("got branch %q, want main or master", branch)
	}
}

func TestBranchExists_Worktree(t *testing.T) {
	mainDir := t.TempDir()
	initTestRepo(t, mainDir)
	gitCmd(t, mainDir, "branch", "feature-branch")

	worktreePath := filepath.Join(t.TempDir(), "wt")
	gitCmd(t, mainDir, "worktree", "add", worktreePath, "-b", "wt-branch")

	// Branch created in main repo should be visible from worktree
	exists, err := BranchExists(worktreePath, "feature-branch")
	if err != nil {
		t.Fatalf("BranchExists in worktree: %v", err)
	}
	if !exists {
		t.Error("expected feature-branch to exist when queried from worktree")
	}

	// Non-existent branch
	exists, err = BranchExists(worktreePath, "no-such-branch")
	if err != nil {
		t.Fatalf("BranchExists for missing branch: %v", err)
	}
	if exists {
		t.Error("expected no-such-branch to not exist")
	}
}

func TestCheckoutBranch_Worktree(t *testing.T) {
	mainDir := t.TempDir()
	initTestRepo(t, mainDir)

	worktreePath := filepath.Join(t.TempDir(), "wt")
	gitCmd(t, mainDir, "worktree", "add", worktreePath, "-b", "wt-checkout")

	// Create a branch in the worktree, then checkout back
	gitCmd(t, worktreePath, "branch", "target-branch")

	err := CheckoutBranch(worktreePath, "target-branch")
	if err != nil {
		t.Fatalf("CheckoutBranch in worktree: %v", err)
	}

	branch, err := GetCurrentBranch(worktreePath)
	if err != nil {
		t.Fatalf("GetCurrentBranch after checkout: %v", err)
	}
	if branch != "target-branch" {
		t.Errorf("got branch %q after checkout, want %q", branch, "target-branch")
	}
}

func TestAddFilesAndCommitChanges_Worktree(t *testing.T) {
	mainDir := t.TempDir()
	initTestRepo(t, mainDir)

	worktreePath := filepath.Join(t.TempDir(), "wt")
	gitCmd(t, mainDir, "worktree", "add", worktreePath, "-b", "wt-commit")

	// Create a file in the worktree
	if err := os.WriteFile(filepath.Join(worktreePath, "test.txt"), []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Stage it
	err := AddFiles(worktreePath, []string{"test.txt"})
	if err != nil {
		t.Fatalf("AddFiles in worktree: %v", err)
	}

	// Commit it
	hash, err := CommitChanges(worktreePath, "test commit in worktree")
	if err != nil {
		t.Fatalf("CommitChanges in worktree: %v", err)
	}
	if len(hash) != 8 {
		t.Errorf("expected 8-char hash, got %q", hash)
	}

	// Verify clean after commit
	dirty, err := HasUncommittedChanges(worktreePath)
	if err != nil {
		t.Fatalf("HasUncommittedChanges after commit: %v", err)
	}
	if dirty {
		t.Error("expected clean worktree after commit")
	}
}

func TestGetRepoOwnerName_Worktree(t *testing.T) {
	mainDir := t.TempDir()
	initTestRepo(t, mainDir)
	gitCmd(t, mainDir, "remote", "add", "origin", "https://github.com/test-owner/test-repo.git")

	worktreePath := filepath.Join(t.TempDir(), "wt")
	gitCmd(t, mainDir, "worktree", "add", worktreePath, "-b", "wt-remote")

	owner, name, err := GetRepoOwnerName(worktreePath)
	if err != nil {
		t.Fatalf("GetRepoOwnerName in worktree: %v", err)
	}
	if owner != "test-owner" {
		t.Errorf("owner = %q, want %q", owner, "test-owner")
	}
	if name != "test-repo" {
		t.Errorf("name = %q, want %q", name, "test-repo")
	}
}

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
		{
			name:      "SSH dotted repo name",
			url:       "git@github.com:owner/my.repo.git",
			wantOwner: "owner",
			wantName:  "my.repo",
		},
		{
			name:      "HTTPS dotted repo name without .git",
			url:       "https://github.com/owner/my.repo",
			wantOwner: "owner",
			wantName:  "my.repo",
		},
		{
			name:      "ssh:// dotted repo name",
			url:       "ssh://git@host:22/owner/my.dotted.repo.git",
			wantOwner: "owner",
			wantName:  "my.dotted.repo",
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
