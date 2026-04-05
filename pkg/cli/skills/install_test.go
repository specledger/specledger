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
	source := &SkillSource{Owner: "org", Repo: "repo", Ref: "main", Type: SourceTypeGitHub}

	err := InstallSkill("test-skill", content, []string{agentPath}, lockPath, source)
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
	source := &SkillSource{Owner: "org", Repo: "repo", Ref: "main", Type: SourceTypeGitHub}

	err := InstallSkill("multi", content, []string{path1, path2}, lockPath, source)
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
	source := &SkillSource{Owner: "org", Repo: "repo", Ref: "main", Type: SourceTypeGitHub}
	if err := InstallSkill("removeme", content, []string{agentPath}, lockPath, source); err != nil {
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
