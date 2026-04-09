package skills

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadLocalLock_MissingFile(t *testing.T) {
	lock, err := ReadLocalLock("/nonexistent/skills-lock.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lock.Version != 1 {
		t.Errorf("Version = %d, want 1", lock.Version)
	}
	if len(lock.Skills) != 0 {
		t.Errorf("Skills = %v, want empty map", lock.Skills)
	}
}

func TestReadLocalLock_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "skills-lock.json")
	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	lock, err := ReadLocalLock(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lock.Version != 1 {
		t.Errorf("Version = %d, want 1", lock.Version)
	}
}

func TestReadLocalLock_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "skills-lock.json")
	if err := os.WriteFile(path, []byte("{invalid json}"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := ReadLocalLock(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !contains(err.Error(), "skills-lock.json is invalid") {
		t.Errorf("error = %q, want containing 'skills-lock.json is invalid'", err.Error())
	}
}

func TestReadLocalLock_ValidFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "skills-lock.json")
	data := `{
  "version": 1,
  "skills": {
    "creating-pr": {
      "source": "vercel-labs/agent-skills",
      "sourceType": "github",
      "computedHash": "abc123"
    }
  }
}`
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	lock, err := ReadLocalLock(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lock.Version != 1 {
		t.Errorf("Version = %d, want 1", lock.Version)
	}
	entry, ok := lock.Skills["creating-pr"]
	if !ok {
		t.Fatal("missing 'creating-pr' entry")
	}
	if entry.Source != "vercel-labs/agent-skills" {
		t.Errorf("Source = %q, want %q", entry.Source, "vercel-labs/agent-skills")
	}
	if entry.ComputedHash != "abc123" {
		t.Errorf("ComputedHash = %q, want %q", entry.ComputedHash, "abc123")
	}
}

func TestWriteLocalLock_SortedAndFormatted(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "skills-lock.json")

	lock := &LocalSkillLockFile{
		Version: 1,
		Skills: map[string]LocalSkillLockEntry{
			"zebra-skill": {Source: "org/repo", SourceType: "github", ComputedHash: "z123"},
			"alpha-skill": {Source: "org/repo", SourceType: "github", ComputedHash: "a123"},
			"mid-skill":   {Source: "org/repo", SourceType: "github", ComputedHash: "m123"},
		},
	}

	if err := WriteLocalLock(path, lock); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)

	// Must end with newline
	if content[len(content)-1] != '\n' {
		t.Error("file does not end with newline")
	}

	// Alpha must appear before mid, mid before zebra
	alphaIdx := indexOf(content, "alpha-skill")
	midIdx := indexOf(content, "mid-skill")
	zebraIdx := indexOf(content, "zebra-skill")
	if alphaIdx >= midIdx || midIdx >= zebraIdx {
		t.Errorf("skills not sorted alphabetically: alpha=%d, mid=%d, zebra=%d", alphaIdx, midIdx, zebraIdx)
	}
}

func TestWriteReadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "skills-lock.json")

	original := &LocalSkillLockFile{
		Version: 1,
		Skills: map[string]LocalSkillLockEntry{
			"creating-pr": {
				Source:       "vercel-labs/agent-skills",
				Ref:          "main",
				SourceType:   "github",
				ComputedHash: "abc123def456",
			},
		},
	}

	if err := WriteLocalLock(path, original); err != nil {
		t.Fatalf("WriteLocalLock: %v", err)
	}

	readBack, err := ReadLocalLock(path)
	if err != nil {
		t.Fatalf("ReadLocalLock: %v", err)
	}

	if readBack.Version != original.Version {
		t.Errorf("Version = %d, want %d", readBack.Version, original.Version)
	}

	entry, ok := readBack.Skills["creating-pr"]
	if !ok {
		t.Fatal("missing 'creating-pr' after round-trip")
	}
	orig := original.Skills["creating-pr"]
	if entry.Source != orig.Source || entry.Ref != orig.Ref || entry.SourceType != orig.SourceType || entry.ComputedHash != orig.ComputedHash {
		t.Errorf("round-trip mismatch: got %+v, want %+v", entry, orig)
	}
}

func TestAddRemoveSkill(t *testing.T) {
	lock := &LocalSkillLockFile{
		Version: 1,
		Skills:  make(map[string]LocalSkillLockEntry),
	}

	AddSkill(lock, "test-skill", LocalSkillLockEntry{
		Source:       "org/repo",
		SourceType:   "github",
		ComputedHash: "hash123",
	})

	if _, ok := lock.Skills["test-skill"]; !ok {
		t.Fatal("AddSkill: skill not added")
	}

	RemoveSkill(lock, "test-skill")
	if _, ok := lock.Skills["test-skill"]; ok {
		t.Fatal("RemoveSkill: skill not removed")
	}

	// Removing non-existent skill should not panic
	RemoveSkill(lock, "nonexistent")
}

func TestReadLocalLock_NullSkills(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "skills-lock.json")
	if err := os.WriteFile(path, []byte(`{"version":1,"skills":null}`), 0644); err != nil {
		t.Fatal(err)
	}

	lock, err := ReadLocalLock(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lock.Skills == nil {
		t.Error("Skills map should be non-nil after reading null skills")
	}
	if len(lock.Skills) != 0 {
		t.Errorf("Skills = %v, want empty map", lock.Skills)
	}
}

func TestAddSkill_NilSkillsMap(t *testing.T) {
	lock := &LocalSkillLockFile{
		Version: 1,
		Skills:  nil,
	}

	AddSkill(lock, "new-skill", LocalSkillLockEntry{
		Source:       "org/repo",
		SourceType:   "github",
		ComputedHash: "abc123",
	})

	if lock.Skills == nil {
		t.Fatal("Skills map should be initialized")
	}
	entry, ok := lock.Skills["new-skill"]
	if !ok {
		t.Fatal("skill not added")
	}
	if entry.Source != "org/repo" {
		t.Errorf("Source = %q, want %q", entry.Source, "org/repo")
	}
}

func TestWriteLocalLock_EmptySkills(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "skills-lock.json")

	lock := &LocalSkillLockFile{
		Version: 1,
		Skills:  make(map[string]LocalSkillLockEntry),
	}

	if err := WriteLocalLock(path, lock); err != nil {
		t.Fatalf("WriteLocalLock: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !contains(content, `"skills": {}`) {
		t.Errorf("expected empty skills object in output, got:\n%s", content)
	}
}

func TestWriteLocalLock_VersionZero(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "skills-lock.json")

	lock := &LocalSkillLockFile{
		Version: 0,
		Skills:  make(map[string]LocalSkillLockEntry),
	}

	if err := WriteLocalLock(path, lock); err != nil {
		t.Fatalf("WriteLocalLock: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !contains(content, `"version": 1`) {
		t.Errorf("expected version 1 in output, got:\n%s", content)
	}
}

func TestReadLocalLock_PermissionDenied(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "skills-lock.json")

	// Create file then make it unreadable
	if err := os.WriteFile(path, []byte(`{"version":1,"skills":{}}`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0000); err != nil {
		t.Skip("cannot change file permissions on this OS")
	}
	// Restore permissions on cleanup so t.TempDir() can clean up
	t.Cleanup(func() { _ = os.Chmod(path, 0644) })

	_, err := ReadLocalLock(path)
	if err == nil {
		t.Fatal("expected error for permission denied")
	}
	if !contains(err.Error(), "failed to read") {
		t.Errorf("error = %q, want containing 'failed to read'", err.Error())
	}
}

func TestWriteLocalLock_PermissionDenied(t *testing.T) {
	dir := t.TempDir()

	// Make directory read-only so file creation fails
	if err := os.Chmod(dir, 0555); err != nil {
		t.Skip("cannot change directory permissions on this OS")
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0755) })

	path := filepath.Join(dir, "skills-lock.json")
	lock := &LocalSkillLockFile{
		Version: 1,
		Skills:  make(map[string]LocalSkillLockEntry),
	}

	err := WriteLocalLock(path, lock)
	if err == nil {
		t.Fatal("expected error for permission denied")
	}
	if !contains(err.Error(), "failed to write") {
		t.Errorf("error = %q, want containing 'failed to write'", err.Error())
	}
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
