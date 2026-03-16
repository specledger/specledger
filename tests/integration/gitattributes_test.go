package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/specledger/specledger/pkg/cli/playbooks"
)

// buildBinary is a test helper that builds sl once per test function using a shared temp dir.
func buildBinary(t *testing.T) string {
	t.Helper()
	return buildSLBinary(t, t.TempDir())
}

// initProject runs sl init in the given directory and returns the combined output.
func initProject(t *testing.T, slBinary, dir string, extraArgs ...string) string {
	t.Helper()
	args := append([]string{"init", "--short-code", "ga", "--ci"}, extraArgs...)
	cmd := exec.Command(slBinary, args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sl init failed: %v\nOutput: %s", err, string(output))
	}
	return string(output)
}

// readGitattributes reads the .gitattributes file from the project directory.
func readGitattributes(t *testing.T, dir string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, ".gitattributes"))
	if err != nil {
		t.Fatalf("Failed to read .gitattributes: %v", err)
	}
	return string(data)
}

// TestGitattributesInitCreates tests that sl init creates .gitattributes with sentinel block (US1).
func TestGitattributesInitCreates(t *testing.T) {
	slBinary := buildBinary(t)
	projectDir := filepath.Join(t.TempDir(), "new-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	initProject(t, slBinary, projectDir)

	content := readGitattributes(t, projectDir)

	if !strings.Contains(content, playbooks.SentinelBegin) {
		t.Error(".gitattributes missing sentinel begin marker")
	}
	if !strings.Contains(content, playbooks.SentinelEnd) {
		t.Error(".gitattributes missing sentinel end marker")
	}
	if !strings.Contains(content, "issues.jsonl linguist-generated=true") {
		t.Error(".gitattributes missing issues.jsonl pattern")
	}
	if !strings.Contains(content, "tasks.md linguist-generated=true") {
		t.Error(".gitattributes missing tasks.md pattern")
	}
}

// TestGitattributesInitMerges tests that sl init merges into existing .gitattributes (US2).
func TestGitattributesInitMerges(t *testing.T) {
	slBinary := buildBinary(t)
	projectDir := filepath.Join(t.TempDir(), "existing-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create existing .gitattributes with user content
	userContent := "*.pbxproj binary\n*.png filter=lfs diff=lfs merge=lfs -text\n"
	if err := os.WriteFile(filepath.Join(projectDir, ".gitattributes"), []byte(userContent), 0644); err != nil {
		t.Fatal(err)
	}

	initProject(t, slBinary, projectDir)

	content := readGitattributes(t, projectDir)

	// User content preserved
	if !strings.Contains(content, "*.pbxproj binary") {
		t.Error("User content not preserved: *.pbxproj binary")
	}
	if !strings.Contains(content, "*.png filter=lfs") {
		t.Error("User content not preserved: *.png filter=lfs")
	}

	// Sentinel block added
	if !strings.Contains(content, playbooks.SentinelBegin) {
		t.Error("Sentinel block not added")
	}
	if !strings.Contains(content, "issues.jsonl linguist-generated=true") {
		t.Error("Missing issues.jsonl pattern")
	}
}

// TestGitattributesInitUpdates tests that sl init updates an existing sentinel block (US2 re-init).
func TestGitattributesInitUpdates(t *testing.T) {
	slBinary := buildBinary(t)
	projectDir := filepath.Join(t.TempDir(), "update-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create .gitattributes with an old sentinel block
	oldContent := "*.pbxproj binary\n\n" +
		playbooks.SentinelBegin + "\n" +
		playbooks.SentinelComment + "\n" +
		"old/pattern linguist-generated=true\n" +
		playbooks.SentinelEnd + "\n"
	if err := os.WriteFile(filepath.Join(projectDir, ".gitattributes"), []byte(oldContent), 0644); err != nil {
		t.Fatal(err)
	}

	initProject(t, slBinary, projectDir)

	content := readGitattributes(t, projectDir)

	// Old content replaced
	if strings.Contains(content, "old/pattern") {
		t.Error("Old sentinel content should have been replaced")
	}

	// New content present
	if !strings.Contains(content, "issues.jsonl linguist-generated=true") {
		t.Error("New sentinel content missing")
	}

	// User content preserved
	if !strings.Contains(content, "*.pbxproj binary") {
		t.Error("User content not preserved")
	}

	// No duplicate sentinel markers
	if strings.Count(content, playbooks.SentinelBegin) != 1 {
		t.Errorf("Expected exactly 1 sentinel begin marker, got %d", strings.Count(content, playbooks.SentinelBegin))
	}
}

// TestGitattributesInitIdempotent tests that running sl init twice produces identical output (US3).
func TestGitattributesInitIdempotent(t *testing.T) {
	slBinary := buildBinary(t)
	projectDir := filepath.Join(t.TempDir(), "idempotent-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create existing user content
	if err := os.WriteFile(filepath.Join(projectDir, ".gitattributes"), []byte("*.pbxproj binary\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// First init
	initProject(t, slBinary, projectDir)
	first := readGitattributes(t, projectDir)

	// Second init (with --force to re-run copy)
	initProject(t, slBinary, projectDir, "--force")
	second := readGitattributes(t, projectDir)

	if first != second {
		t.Errorf("Not idempotent:\nFirst:\n%s\nSecond:\n%s", first, second)
	}
}

// TestGitattributesInitForceMerges tests that sl init --force merges (not overwrites) (FR-010).
func TestGitattributesInitForceMerges(t *testing.T) {
	slBinary := buildBinary(t)
	projectDir := filepath.Join(t.TempDir(), "force-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create existing .gitattributes with user content
	userContent := "*.pbxproj binary\n"
	if err := os.WriteFile(filepath.Join(projectDir, ".gitattributes"), []byte(userContent), 0644); err != nil {
		t.Fatal(err)
	}

	// First init to get sentinel block
	initProject(t, slBinary, projectDir)

	// Force re-init
	initProject(t, slBinary, projectDir, "--force")

	content := readGitattributes(t, projectDir)

	// User content MUST still be there (--force should merge, not overwrite)
	if !strings.Contains(content, "*.pbxproj binary") {
		t.Error("--force overwrote user content instead of merging")
	}
	if !strings.Contains(content, playbooks.SentinelBegin) {
		t.Error("Sentinel block missing after --force")
	}
}

// TestGitattributesInitMalformedSentinel tests that sl init handles malformed sentinel (FR-011).
func TestGitattributesInitMalformedSentinel(t *testing.T) {
	slBinary := buildBinary(t)
	projectDir := filepath.Join(t.TempDir(), "malformed-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create .gitattributes with malformed sentinel (begin without end)
	malformed := "*.pbxproj binary\n\n" +
		playbooks.SentinelBegin + "\n" +
		"orphaned content\n" +
		"more orphaned\n"
	if err := os.WriteFile(filepath.Join(projectDir, ".gitattributes"), []byte(malformed), 0644); err != nil {
		t.Fatal(err)
	}

	initProject(t, slBinary, projectDir)

	content := readGitattributes(t, projectDir)

	// Malformed content should be replaced
	if strings.Contains(content, "orphaned content") {
		t.Error("Malformed sentinel content should have been replaced")
	}

	// Proper sentinel block should be present
	if !strings.Contains(content, playbooks.SentinelBegin) {
		t.Error("Sentinel begin missing")
	}
	if !strings.Contains(content, playbooks.SentinelEnd) {
		t.Error("Sentinel end missing")
	}

	// User content preserved
	if !strings.Contains(content, "*.pbxproj binary") {
		t.Error("User content not preserved")
	}
}

// TestGitattributesDoctorTemplateMerges tests that sl doctor --template merges .gitattributes.
func TestGitattributesDoctorTemplateMerges(t *testing.T) {
	slBinary := buildBinary(t)
	projectDir := filepath.Join(t.TempDir(), "doctor-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	// First init to set up the project
	initProject(t, slBinary, projectDir)

	// Add user content to .gitattributes after init
	content := readGitattributes(t, projectDir)
	newContent := "*.custom-ext binary\n" + content
	if err := os.WriteFile(filepath.Join(projectDir, ".gitattributes"), []byte(newContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Run sl doctor --template
	cmd := exec.Command(slBinary, "doctor", "--template")
	cmd.Dir = projectDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sl doctor --template failed: %v\nOutput: %s", err, string(output))
	}

	afterDoctor := readGitattributes(t, projectDir)

	// User content preserved
	if !strings.Contains(afterDoctor, "*.custom-ext binary") {
		t.Error("User content not preserved after doctor --template")
	}

	// Sentinel block present
	if !strings.Contains(afterDoctor, playbooks.SentinelBegin) {
		t.Error("Sentinel block missing after doctor --template")
	}
}

// TestGitattributesDoctorTemplateIdempotent tests that sl doctor --template is idempotent (US3).
func TestGitattributesDoctorTemplateIdempotent(t *testing.T) {
	slBinary := buildBinary(t)
	projectDir := filepath.Join(t.TempDir(), "doctor-idempotent")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Init project
	initProject(t, slBinary, projectDir)

	// First doctor --template
	cmd1 := exec.Command(slBinary, "doctor", "--template")
	cmd1.Dir = projectDir
	if output, err := cmd1.CombinedOutput(); err != nil {
		t.Fatalf("First doctor --template failed: %v\nOutput: %s", err, string(output))
	}
	first := readGitattributes(t, projectDir)

	// Second doctor --template
	cmd2 := exec.Command(slBinary, "doctor", "--template")
	cmd2.Dir = projectDir
	if output, err := cmd2.CombinedOutput(); err != nil {
		t.Fatalf("Second doctor --template failed: %v\nOutput: %s", err, string(output))
	}
	second := readGitattributes(t, projectDir)

	if first != second {
		t.Errorf("doctor --template not idempotent:\nFirst:\n%s\nSecond:\n%s", first, second)
	}
}
