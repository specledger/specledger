package skills

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstallSkill(t *testing.T) {
	dir := t.TempDir()
	agentPath := filepath.Join(dir, ".claude", "skills")
	lockPath := filepath.Join(dir, "skills-lock.json")

	content := []byte("---\nname: test-skill\ndescription: Test\n---\nBody")
	files := map[string][]byte{"SKILL.md": content}
	source := &SkillSource{Owner: "org", Repo: "repo", Ref: "main", Type: SourceTypeGitHub}

	err := InstallSkill("test-skill", files, []string{agentPath}, lockPath, source)
	if err != nil {
		t.Fatalf("InstallSkill: %v", err)
	}

	// Verify SKILL.md was written
	skillFile := filepath.Join(agentPath, "test-skill", "SKILL.md")
	data, err := os.ReadFile(skillFile)
	if err != nil {
		t.Fatalf("SKILL.md not created: %v", err)
	}
	if string(data) != string(content) {
		t.Errorf("content mismatch: got %q", string(data))
	}

	// Verify lock file was updated
	lock, err := ReadLocalLock(lockPath)
	if err != nil {
		t.Fatalf("ReadLocalLock: %v", err)
	}
	entry, ok := lock.Skills["test-skill"]
	if !ok {
		t.Fatal("skill not in lock file")
	}
	if entry.Source != "org/repo" {
		t.Errorf("Source = %q, want %q", entry.Source, "org/repo")
	}
	if entry.ComputedHash == "" {
		t.Error("ComputedHash is empty")
	}
}

func TestInstallSkill_MultipleAgentPaths(t *testing.T) {
	dir := t.TempDir()
	path1 := filepath.Join(dir, ".claude", "skills")
	path2 := filepath.Join(dir, ".opencode", "skills")
	lockPath := filepath.Join(dir, "skills-lock.json")

	content := []byte("---\nname: multi\ndescription: Multi-agent\n---\n")
	files := map[string][]byte{"SKILL.md": content}
	source := &SkillSource{Owner: "org", Repo: "repo", Ref: "main", Type: SourceTypeGitHub}

	err := InstallSkill("multi", files, []string{path1, path2}, lockPath, source)
	if err != nil {
		t.Fatalf("InstallSkill: %v", err)
	}

	// Both paths should have the file
	for _, p := range []string{path1, path2} {
		f := filepath.Join(p, "multi", "SKILL.md")
		if _, err := os.Stat(f); os.IsNotExist(err) {
			t.Errorf("SKILL.md not created at %s", f)
		}
	}
}

func TestUninstallSkill(t *testing.T) {
	dir := t.TempDir()
	agentPath := filepath.Join(dir, ".claude", "skills")
	lockPath := filepath.Join(dir, "skills-lock.json")

	// Install first
	content := []byte("---\nname: removeme\ndescription: Test\n---\n")
	files := map[string][]byte{"SKILL.md": content}
	source := &SkillSource{Owner: "org", Repo: "repo", Ref: "main", Type: SourceTypeGitHub}
	if err := InstallSkill("removeme", files, []string{agentPath}, lockPath, source); err != nil {
		t.Fatalf("InstallSkill: %v", err)
	}

	// Uninstall
	err := UninstallSkill("removeme", []string{agentPath}, lockPath)
	if err != nil {
		t.Fatalf("UninstallSkill: %v", err)
	}

	// Verify directory removed
	skillDir := filepath.Join(agentPath, "removeme")
	if _, err := os.Stat(skillDir); !os.IsNotExist(err) {
		t.Error("skill directory still exists after uninstall")
	}

	// Verify lock file updated
	lock, err := ReadLocalLock(lockPath)
	if err != nil {
		t.Fatalf("ReadLocalLock: %v", err)
	}
	if _, ok := lock.Skills["removeme"]; ok {
		t.Error("skill still in lock file after uninstall")
	}
}

func TestIsSkillInstalled(t *testing.T) {
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "skills-lock.json")

	// Not installed
	if IsSkillInstalled("nope", lockPath) {
		t.Error("skill should not be installed")
	}

	// Install
	lock := &LocalSkillLockFile{
		Version: 1,
		Skills: map[string]LocalSkillLockEntry{
			"installed": {Source: "org/repo", SourceType: "github", ComputedHash: "abc"},
		},
	}
	if err := WriteLocalLock(lockPath, lock); err != nil {
		t.Fatal(err)
	}

	if !IsSkillInstalled("installed", lockPath) {
		t.Error("skill should be installed")
	}
}

func TestValidateSkillName(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"valid-skill", false},
		{"my_skill", false},
		{"skill123", false},
		{"", true},             // empty
		{"../evil", true},      // path traversal
		{"foo/bar", true},      // slash
		{"foo\\bar", true},     // backslash
		{"/absolute", true},    // absolute path
		{"skill/../etc", true}, // embedded traversal
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSkillName(tt.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSkillName(%q) error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}

func TestInstallSkill_PathTraversal(t *testing.T) {
	dir := t.TempDir()
	agentPath := filepath.Join(dir, ".claude", "skills")
	lockPath := filepath.Join(dir, "skills-lock.json")
	source := &SkillSource{Owner: "org", Repo: "repo", Ref: "main", Type: SourceTypeGitHub}

	err := InstallSkill("../../../etc/passwd", map[string][]byte{"SKILL.md": []byte("evil")}, []string{agentPath}, lockPath, source)
	if err == nil {
		t.Fatal("expected error for path traversal name")
	}
	if !contains(err.Error(), "unsafe skill name") {
		t.Errorf("error = %q, want containing 'unsafe skill name'", err.Error())
	}
}

func TestInstallSkill_EmptyAgentPaths(t *testing.T) {
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "skills-lock.json")
	source := &SkillSource{Owner: "org", Repo: "repo", Ref: "main", Type: SourceTypeGitHub}

	err := InstallSkill("valid-skill", map[string][]byte{"SKILL.md": []byte("content")}, []string{}, lockPath, source)
	if err == nil {
		t.Fatal("expected error for empty agent paths")
	}
	if !contains(err.Error(), "no agent paths") {
		t.Errorf("error = %q, want containing 'no agent paths'", err.Error())
	}
}

func TestInstallSkill_InvalidName(t *testing.T) {
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "skills-lock.json")
	agentPath := filepath.Join(dir, ".claude", "skills")
	source := &SkillSource{Owner: "org", Repo: "repo", Ref: "main", Type: SourceTypeGitHub}

	err := InstallSkill("../evil", map[string][]byte{"SKILL.md": []byte("content")}, []string{agentPath}, lockPath, source)
	if err == nil {
		t.Fatal("expected error for invalid name")
	}
	if !contains(err.Error(), "unsafe skill name") {
		t.Errorf("error = %q, want containing 'unsafe skill name'", err.Error())
	}

	// Verify no I/O occurred — agent dir should not exist
	if _, err := os.Stat(agentPath); !os.IsNotExist(err) {
		t.Error("agent directory should not have been created")
	}
}

func TestInstallSkill_RefInLockEntry(t *testing.T) {
	dir := t.TempDir()
	agentPath := filepath.Join(dir, ".claude", "skills")
	lockPath := filepath.Join(dir, "skills-lock.json")

	content := []byte("---\nname: ref-skill\ndescription: Test\n---\nBody")
	files := map[string][]byte{"SKILL.md": content}
	source := &SkillSource{Owner: "org", Repo: "repo", Ref: "v1.0", Type: SourceTypeGitHub}

	err := InstallSkill("ref-skill", files, []string{agentPath}, lockPath, source)
	if err != nil {
		t.Fatalf("InstallSkill: %v", err)
	}

	lock, err := ReadLocalLock(lockPath)
	if err != nil {
		t.Fatalf("ReadLocalLock: %v", err)
	}

	entry, ok := lock.Skills["ref-skill"]
	if !ok {
		t.Fatal("skill not in lock file")
	}
	if entry.Ref != "v1.0" {
		t.Errorf("Ref = %q, want %q", entry.Ref, "v1.0")
	}
}

func TestUninstallSkill_InvalidName(t *testing.T) {
	dir := t.TempDir()
	agentPath := filepath.Join(dir, ".claude", "skills")
	lockPath := filepath.Join(dir, "skills-lock.json")

	err := UninstallSkill("../evil", []string{agentPath}, lockPath)
	if err == nil {
		t.Fatal("expected error for invalid name")
	}
	if !contains(err.Error(), "unsafe skill name") {
		t.Errorf("error = %q, want containing 'unsafe skill name'", err.Error())
	}
}

func TestUninstallSkill_NotInstalled(t *testing.T) {
	dir := t.TempDir()
	agentPath := filepath.Join(dir, ".claude", "skills")
	lockPath := filepath.Join(dir, "skills-lock.json")

	// Uninstall a skill that was never installed — should succeed
	err := UninstallSkill("nonexistent-skill", []string{agentPath}, lockPath)
	if err != nil {
		t.Fatalf("UninstallSkill should succeed for non-installed skill: %v", err)
	}
}

func TestIsSkillInstalled_CorruptLock(t *testing.T) {
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "skills-lock.json")

	// Write invalid JSON to the lock file
	if err := os.WriteFile(lockPath, []byte("{corrupt json!!!}"), 0644); err != nil {
		t.Fatal(err)
	}

	// ReadLocalLock will error → IsSkillInstalled should return false
	if IsSkillInstalled("any-skill", lockPath) {
		t.Error("expected false for corrupt lock file")
	}
}

func TestInstallSkill_DirCreationFailure(t *testing.T) {
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "skills-lock.json")
	source := &SkillSource{Owner: "org", Repo: "repo", Ref: "main", Type: SourceTypeGitHub}

	// Create a regular file where the agent path should be a directory.
	// This makes MkdirAll fail when trying to create a subdirectory.
	agentFile := filepath.Join(dir, "not-a-dir")
	if err := os.WriteFile(agentFile, []byte("I am a file"), 0644); err != nil {
		t.Fatal(err)
	}

	err := InstallSkill("some-skill", map[string][]byte{"SKILL.md": []byte("content")}, []string{agentFile}, lockPath, source)
	if err == nil {
		t.Fatal("expected error when agent path is a file")
	}
	if !contains(err.Error(), "failed to create") {
		t.Errorf("error = %q, want containing 'failed to create'", err.Error())
	}
}

func TestInstallSkill_PathTraversalInFiles(t *testing.T) {
	dir := t.TempDir()
	agentPath := filepath.Join(dir, ".claude", "skills")
	lockPath := filepath.Join(dir, "skills-lock.json")
	source := &SkillSource{Owner: "org", Repo: "repo", Ref: "main", Type: SourceTypeGitHub}

	files := map[string][]byte{
		"SKILL.md":   []byte("---\nname: test\ndescription: Test\n---\n"),
		"../evil.md": []byte("malicious content"),
	}
	err := InstallSkill("test-skill", files, []string{agentPath}, lockPath, source)
	if err == nil {
		t.Fatal("expected error for path traversal in files map key")
	}
	if !contains(err.Error(), "unsafe file path") {
		t.Errorf("error = %q, want containing 'unsafe file path'", err.Error())
	}
}

func TestInstallSkill_MissingSKILLMD(t *testing.T) {
	dir := t.TempDir()
	agentPath := filepath.Join(dir, ".claude", "skills")
	lockPath := filepath.Join(dir, "skills-lock.json")
	source := &SkillSource{Owner: "org", Repo: "repo", Ref: "main", Type: SourceTypeGitHub}

	files := map[string][]byte{
		"GENERATION.md": []byte("# Generation"),
	}
	err := InstallSkill("test-skill", files, []string{agentPath}, lockPath, source)
	if err == nil {
		t.Fatal("expected error for missing SKILL.md")
	}
	if !contains(err.Error(), "missing SKILL.md") {
		t.Errorf("error = %q, want containing 'missing SKILL.md'", err.Error())
	}
}

func TestInstallSkill_MultipleFiles(t *testing.T) {
	dir := t.TempDir()
	agentPath := filepath.Join(dir, ".claude", "skills")
	lockPath := filepath.Join(dir, "skills-lock.json")
	source := &SkillSource{Owner: "org", Repo: "repo", Ref: "main", Type: SourceTypeGitHub}

	files := map[string][]byte{
		"SKILL.md":               []byte("---\nname: multi-file\ndescription: Test\n---\nBody"),
		"GENERATION.md":          []byte("# Generation"),
		"references/core.md":     []byte("# Core"),
		"references/advanced.md": []byte("# Advanced"),
	}
	err := InstallSkill("multi-file", files, []string{agentPath}, lockPath, source)
	if err != nil {
		t.Fatalf("InstallSkill: %v", err)
	}

	// Verify all files were written
	for relPath, expected := range files {
		fullPath := filepath.Join(agentPath, "multi-file", filepath.FromSlash(relPath))
		data, readErr := os.ReadFile(fullPath)
		if readErr != nil {
			t.Errorf("file %s not created: %v", relPath, readErr)
			continue
		}
		if string(data) != string(expected) {
			t.Errorf("file %s: content mismatch: got %q, want %q", relPath, string(data), string(expected))
		}
	}

	// Verify lock file has an entry
	lock, readErr := ReadLocalLock(lockPath)
	if readErr != nil {
		t.Fatalf("ReadLocalLock: %v", readErr)
	}
	if _, ok := lock.Skills["multi-file"]; !ok {
		t.Error("skill not in lock file")
	}
}
