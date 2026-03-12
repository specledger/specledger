package hooks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupGitDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git")
	if err := os.MkdirAll(filepath.Join(gitDir, "hooks"), 0755); err != nil {
		t.Fatal(err)
	}
	return gitDir
}

func TestInstallPushHook_Clean(t *testing.T) {
	gitDir := setupGitDir(t)

	if err := InstallPushHook(gitDir, false); err != nil {
		t.Fatalf("InstallPushHook() error: %v", err)
	}

	hookPath := prePushHookPath(gitDir)
	data, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("failed to read hook: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, shebang) {
		t.Error("missing shebang")
	}
	if !strings.Contains(content, beginMarker) {
		t.Error("missing begin marker")
	}
	if !strings.Contains(content, endMarker) {
		t.Error("missing end marker")
	}
	if !strings.Contains(content, "sl hook execute") {
		t.Error("missing hook command")
	}

	// Verify executable
	info, _ := os.Stat(hookPath)
	if info.Mode()&0111 == 0 {
		t.Error("hook file not executable")
	}
}

func TestInstallPushHook_PreservesExisting(t *testing.T) {
	gitDir := setupGitDir(t)
	hookPath := prePushHookPath(gitDir)

	existingContent := "#!/bin/sh\necho 'existing hook'\nrun-tests\n"
	if err := os.WriteFile(hookPath, []byte(existingContent), 0755); err != nil {
		t.Fatal(err)
	}

	if err := InstallPushHook(gitDir, false); err != nil {
		t.Fatalf("InstallPushHook() error: %v", err)
	}

	data, _ := os.ReadFile(hookPath)
	content := string(data)

	if !strings.Contains(content, "existing hook") {
		t.Error("existing hook content was not preserved")
	}
	if !strings.Contains(content, "run-tests") {
		t.Error("existing commands were not preserved")
	}
	if !strings.Contains(content, beginMarker) {
		t.Error("SpecLedger block not added")
	}
}

func TestInstallPushHook_AlreadyInstalled(t *testing.T) {
	gitDir := setupGitDir(t)

	if err := InstallPushHook(gitDir, false); err != nil {
		t.Fatal(err)
	}

	data1, _ := os.ReadFile(prePushHookPath(gitDir))

	// Install again without force
	if err := InstallPushHook(gitDir, false); err != nil {
		t.Fatal(err)
	}

	data2, _ := os.ReadFile(prePushHookPath(gitDir))

	if string(data1) != string(data2) {
		t.Error("non-forced reinstall modified the file")
	}
}

func TestInstallPushHook_ForceReinstall(t *testing.T) {
	gitDir := setupGitDir(t)

	if err := InstallPushHook(gitDir, false); err != nil {
		t.Fatal(err)
	}

	if err := InstallPushHook(gitDir, true); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(prePushHookPath(gitDir))
	content := string(data)
	count := strings.Count(content, beginMarker)
	if count != 1 {
		t.Errorf("force reinstall should have exactly 1 marker block, got %d", count)
	}
}

func TestUninstallPushHook(t *testing.T) {
	gitDir := setupGitDir(t)

	if err := InstallPushHook(gitDir, false); err != nil {
		t.Fatal(err)
	}

	if err := UninstallPushHook(gitDir); err != nil {
		t.Fatalf("UninstallPushHook() error: %v", err)
	}

	// File should be deleted (only had shebang + specledger block)
	if _, err := os.Stat(prePushHookPath(gitDir)); !os.IsNotExist(err) {
		t.Error("hook file should be deleted after uninstall")
	}
}

func TestUninstallPushHook_PreservesOther(t *testing.T) {
	gitDir := setupGitDir(t)
	hookPath := prePushHookPath(gitDir)

	existing := "#!/bin/sh\necho 'other hook'\n"
	if err := os.WriteFile(hookPath, []byte(existing), 0755); err != nil {
		t.Fatal(err)
	}

	if err := InstallPushHook(gitDir, false); err != nil {
		t.Fatal(err)
	}

	if err := UninstallPushHook(gitDir); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatal("hook file should still exist")
	}

	content := string(data)
	if !strings.Contains(content, "other hook") {
		t.Error("other hook content was not preserved")
	}
	if strings.Contains(content, beginMarker) {
		t.Error("SpecLedger block was not removed")
	}
}

func TestUninstallPushHook_NotInstalled(t *testing.T) {
	gitDir := setupGitDir(t)

	// Should not error when not installed
	if err := UninstallPushHook(gitDir); err != nil {
		t.Fatalf("UninstallPushHook() should not error: %v", err)
	}
}

func TestHasPushHook(t *testing.T) {
	gitDir := setupGitDir(t)

	if HasPushHook(gitDir) {
		t.Error("HasPushHook() should be false initially")
	}

	if err := InstallPushHook(gitDir, false); err != nil {
		t.Fatal(err)
	}

	if !HasPushHook(gitDir) {
		t.Error("HasPushHook() should be true after install")
	}

	if err := UninstallPushHook(gitDir); err != nil {
		t.Fatal(err)
	}

	if HasPushHook(gitDir) {
		t.Error("HasPushHook() should be false after uninstall")
	}
}
