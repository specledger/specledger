package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/specledger/specledger/pkg/cli/hooks"
)

func setupTestGitDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git")
	if err := os.MkdirAll(filepath.Join(gitDir, "hooks"), 0755); err != nil {
		t.Fatal(err)
	}
	return gitDir
}

func TestHookInstall_CleanRepo(t *testing.T) {
	gitDir := setupTestGitDir(t)

	if err := hooks.InstallPushHook(gitDir, false); err != nil {
		t.Fatalf("InstallPushHook() error: %v", err)
	}

	if !hooks.HasPushHook(gitDir) {
		t.Error("hook should be installed")
	}

	hookPath := filepath.Join(gitDir, "hooks", "pre-push")
	data, _ := os.ReadFile(hookPath)
	content := string(data)

	if !strings.Contains(content, "#!/bin/sh") {
		t.Error("missing shebang")
	}
	if !strings.Contains(content, "sl hook execute") {
		t.Error("missing sl hook execute command")
	}

	info, _ := os.Stat(hookPath)
	if info.Mode()&0111 == 0 {
		t.Error("hook not executable")
	}
}

func TestHookInstall_PreservesExistingHooks(t *testing.T) {
	gitDir := setupTestGitDir(t)
	hookPath := filepath.Join(gitDir, "hooks", "pre-push")

	existing := "#!/bin/sh\necho 'lint check'\nnpm run lint\n"
	if err := os.WriteFile(hookPath, []byte(existing), 0755); err != nil {
		t.Fatal(err)
	}

	if err := hooks.InstallPushHook(gitDir, false); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(hookPath)
	content := string(data)

	if !strings.Contains(content, "lint check") {
		t.Error("existing hook content not preserved")
	}
	if !strings.Contains(content, "npm run lint") {
		t.Error("existing commands not preserved")
	}
	if !strings.Contains(content, "sl hook execute") {
		t.Error("SpecLedger hook not added")
	}
}

func TestHookUninstall_LeavesOtherHooks(t *testing.T) {
	gitDir := setupTestGitDir(t)
	hookPath := filepath.Join(gitDir, "hooks", "pre-push")

	existing := "#!/bin/sh\necho 'my hook'\n"
	if err := os.WriteFile(hookPath, []byte(existing), 0755); err != nil {
		t.Fatal(err)
	}

	// Install then uninstall
	if err := hooks.InstallPushHook(gitDir, false); err != nil {
		t.Fatal(err)
	}
	if err := hooks.UninstallPushHook(gitDir); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(hookPath)
	content := string(data)

	if !strings.Contains(content, "my hook") {
		t.Error("other hook content not preserved after uninstall")
	}
	if strings.Contains(content, "sl hook execute") {
		t.Error("SpecLedger hook should be removed")
	}
}

func TestHookStatus_Detection(t *testing.T) {
	gitDir := setupTestGitDir(t)

	if hooks.HasPushHook(gitDir) {
		t.Error("should not be installed initially")
	}

	_ = hooks.InstallPushHook(gitDir, false)
	if !hooks.HasPushHook(gitDir) {
		t.Error("should be installed after install")
	}

	_ = hooks.UninstallPushHook(gitDir)
	if hooks.HasPushHook(gitDir) {
		t.Error("should not be installed after uninstall")
	}
}

func TestHookInstall_ForceReinstall(t *testing.T) {
	gitDir := setupTestGitDir(t)

	_ = hooks.InstallPushHook(gitDir, false)
	_ = hooks.InstallPushHook(gitDir, true)

	data, _ := os.ReadFile(filepath.Join(gitDir, "hooks", "pre-push"))
	content := string(data)
	count := strings.Count(content, "# BEGIN SPECLEDGER PUSH HOOK")
	if count != 1 {
		t.Errorf("expected 1 marker block after force reinstall, got %d", count)
	}
}
