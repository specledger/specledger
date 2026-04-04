package context

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/specledger/specledger/pkg/cli/playbooks"
)

func TestUpdate_NewFile(t *testing.T) {
	dir := t.TempDir()
	updater := &AgentUpdater{
		AgentType: "claude",
		FilePath:  filepath.Join(dir, "CLAUDE.md"),
	}
	ctx := &TechnicalContext{
		Language: "Go 1.24",
		Testing:  "go test",
	}

	if err := updater.Update(ctx); err != nil {
		t.Fatalf("Update() error: %v", err)
	}

	content, err := os.ReadFile(updater.FilePath)
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}

	got := string(content)

	// Should contain sentinel markers
	if !strings.Contains(got, playbooks.HTMLMarkers.Begin) {
		t.Error("missing sentinel begin marker")
	}
	if !strings.Contains(got, playbooks.HTMLMarkers.End) {
		t.Error("missing sentinel end marker")
	}
	// Should contain tech entries
	if !strings.Contains(got, "- Go 1.24") {
		t.Error("missing Go 1.24 entry")
	}
	if !strings.Contains(got, "- go test") {
		t.Error("missing go test entry")
	}
}

func TestUpdate_PreservesExistingContent(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "CLAUDE.md")

	userContent := "# My Project\n\n## Build Commands\n\nmake test\nmake lint\n"
	if err := os.WriteFile(filePath, []byte(userContent), 0644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	updater := &AgentUpdater{
		AgentType: "claude",
		FilePath:  filePath,
	}
	ctx := &TechnicalContext{
		Language: "Go 1.24",
	}

	if err := updater.Update(ctx); err != nil {
		t.Fatalf("Update() error: %v", err)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}

	got := string(content)

	// User content must be preserved
	if !strings.Contains(got, "# My Project") {
		t.Error("user content '# My Project' was lost")
	}
	if !strings.Contains(got, "make test") {
		t.Error("user content 'make test' was lost")
	}
	// Managed section must be present
	if !strings.Contains(got, playbooks.HTMLMarkers.Begin) {
		t.Error("missing sentinel begin marker")
	}
	if !strings.Contains(got, "- Go 1.24") {
		t.Error("missing Go 1.24 entry")
	}
}

func TestUpdate_OldMarkersBecomesUserContent(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "CLAUDE.md")

	// Simulate a file with old-format markers
	oldContent := "# Active Technologies\n\n" +
		"This file is auto-generated from plan.md. Manual additions are preserved below.\n\n" +
		"## Active Technologies\n\n" +
		"- Old Tech\n\n" +
		"<!-- MANUAL ADDITIONS START -->\n\n" +
		"## Commits & PRs\n\nSee release-flow.md\n\n" +
		"<!-- MANUAL ADDITIONS END -->\n"

	if err := os.WriteFile(filePath, []byte(oldContent), 0644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	updater := &AgentUpdater{
		AgentType: "claude",
		FilePath:  filePath,
	}
	ctx := &TechnicalContext{
		Language: "Go 1.24",
	}

	if err := updater.Update(ctx); err != nil {
		t.Fatalf("Update() error: %v", err)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}

	got := string(content)

	// Old markers are treated as user content (preserved)
	if !strings.Contains(got, "## Commits & PRs") {
		t.Error("user content from old manual additions was lost")
	}
	// New sentinel markers must be present
	if !strings.Contains(got, playbooks.HTMLMarkers.Begin) {
		t.Error("missing new sentinel begin marker")
	}
	if !strings.Contains(got, "- Go 1.24") {
		t.Error("missing Go 1.24 entry")
	}
}

func TestUpdate_Idempotency(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "CLAUDE.md")

	userContent := "# My Project\n\nCustom instructions.\n"
	if err := os.WriteFile(filePath, []byte(userContent), 0644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	updater := &AgentUpdater{
		AgentType: "claude",
		FilePath:  filePath,
	}
	ctx := &TechnicalContext{
		Language:    "Go 1.24",
		PrimaryDeps: "Cobra (CLI)",
	}

	// First run
	if err := updater.Update(ctx); err != nil {
		t.Fatalf("Update() first run error: %v", err)
	}
	first, _ := os.ReadFile(filePath)

	// Second run
	if err := updater.Update(ctx); err != nil {
		t.Fatalf("Update() second run error: %v", err)
	}
	second, _ := os.ReadFile(filePath)

	// Third run
	if err := updater.Update(ctx); err != nil {
		t.Fatalf("Update() third run error: %v", err)
	}
	third, _ := os.ReadFile(filePath)

	if string(first) != string(second) {
		t.Errorf("Not idempotent: first != second\nfirst:\n%q\nsecond:\n%q", first, second)
	}
	if string(second) != string(third) {
		t.Errorf("Not idempotent: second != third\nsecond:\n%q\nthird:\n%q", second, third)
	}
}

func TestUpdate_UpdatesManagedSection(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "CLAUDE.md")

	userContent := "# My Project\n"
	if err := os.WriteFile(filePath, []byte(userContent), 0644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	updater := &AgentUpdater{
		AgentType: "claude",
		FilePath:  filePath,
	}

	// First update with Go
	ctx1 := &TechnicalContext{Language: "Go 1.24"}
	if err := updater.Update(ctx1); err != nil {
		t.Fatalf("Update() error: %v", err)
	}
	first, _ := os.ReadFile(filePath)
	if !strings.Contains(string(first), "- Go 1.24") {
		t.Error("first update missing Go 1.24")
	}

	// Second update with Python — managed section should update
	ctx2 := &TechnicalContext{Language: "Python 3.12"}
	if err := updater.Update(ctx2); err != nil {
		t.Fatalf("Update() error: %v", err)
	}
	second, _ := os.ReadFile(filePath)

	if !strings.Contains(string(second), "- Python 3.12") {
		t.Error("second update missing Python 3.12")
	}
	if strings.Contains(string(second), "- Go 1.24") {
		t.Error("second update still has Go 1.24 (should be replaced)")
	}
	// User content still there
	if !strings.Contains(string(second), "# My Project") {
		t.Error("user content lost after second update")
	}
}
