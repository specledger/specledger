package revise

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseFixture_Valid(t *testing.T) {
	content := `{
		"branch": "009-feature",
		"comments": [
			{"file_path": "spec.md", "selected_text": "hello", "guidance": "fix it"},
			{"file_path": "plan.md", "selected_text": "", "guidance": ""}
		]
	}`

	f := writeTempFixture(t, content)

	fixture, err := ParseFixture(f)
	if err != nil {
		t.Fatalf("ParseFixture returned error: %v", err)
	}

	if fixture.Branch != "009-feature" {
		t.Errorf("Branch: got %q, want %q", fixture.Branch, "009-feature")
	}
	if len(fixture.Comments) != 2 {
		t.Fatalf("Comments length: got %d, want 2", len(fixture.Comments))
	}
	if fixture.Comments[0].FilePath != "spec.md" {
		t.Errorf("Comments[0].FilePath: got %q, want %q", fixture.Comments[0].FilePath, "spec.md")
	}
	if fixture.Comments[0].Guidance != "fix it" {
		t.Errorf("Comments[0].Guidance: got %q, want %q", fixture.Comments[0].Guidance, "fix it")
	}
}

func TestParseFixture_InvalidJSON(t *testing.T) {
	f := writeTempFixture(t, `{not valid json}`)

	_, err := ParseFixture(f)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestParseFixture_NotFound(t *testing.T) {
	_, err := ParseFixture("/nonexistent/path/fixture.json")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestMatchFixtureComments(t *testing.T) {
	comments := []ReviewComment{
		{ID: "c1", FilePath: "spec.md", SelectedText: "hello world", Content: "fix this"},
		{ID: "c2", FilePath: "plan.md", SelectedText: "foo bar", Content: "update this"},
		{ID: "c3", FilePath: "spec.md", SelectedText: "another bit", Content: "other"},
	}

	t.Run("all matched", func(t *testing.T) {
		fixture := &AutoFixture{
			Branch: "test",
			Comments: []FixtureComment{
				{FilePath: "spec.md", SelectedText: "hello world", Guidance: "g1"},
				{FilePath: "plan.md", SelectedText: "foo bar", Guidance: "g2"},
			},
		}

		processed, warnings := MatchFixtureComments(fixture, comments)
		if len(warnings) != 0 {
			t.Errorf("expected 0 warnings, got %v", warnings)
		}
		if len(processed) != 2 {
			t.Fatalf("expected 2 matched, got %d", len(processed))
		}
		if processed[0].Comment.ID != "c1" {
			t.Errorf("processed[0].Comment.ID: got %q, want %q", processed[0].Comment.ID, "c1")
		}
		if processed[0].Guidance != "g1" {
			t.Errorf("processed[0].Guidance: got %q, want %q", processed[0].Guidance, "g1")
		}
		if processed[1].Comment.ID != "c2" {
			t.Errorf("processed[1].Comment.ID: got %q, want %q", processed[1].Comment.ID, "c2")
		}
	})

	t.Run("unmatched entries produce warnings", func(t *testing.T) {
		fixture := &AutoFixture{
			Comments: []FixtureComment{
				{FilePath: "spec.md", SelectedText: "hello world"},
				{FilePath: "missing.md", SelectedText: "not there"},
			},
		}

		processed, warnings := MatchFixtureComments(fixture, comments)
		if len(processed) != 1 {
			t.Errorf("expected 1 matched, got %d", len(processed))
		}
		if len(warnings) != 1 {
			t.Fatalf("expected 1 warning, got %d: %v", len(warnings), warnings)
		}
		if !strings.Contains(warnings[0], "missing.md") {
			t.Errorf("warning should mention missing.md, got: %s", warnings[0])
		}
	})

	t.Run("index assigned sequentially from 1", func(t *testing.T) {
		fixture := &AutoFixture{
			Comments: []FixtureComment{
				{FilePath: "spec.md", SelectedText: "hello world"},
				{FilePath: "plan.md", SelectedText: "foo bar"},
			},
		}

		processed, _ := MatchFixtureComments(fixture, comments)
		for i, p := range processed {
			if p.Index != i+1 {
				t.Errorf("processed[%d].Index: got %d, want %d", i, p.Index, i+1)
			}
		}
	})

	t.Run("empty fixture returns empty", func(t *testing.T) {
		fixture := &AutoFixture{}

		processed, warnings := MatchFixtureComments(fixture, comments)
		if len(processed) != 0 {
			t.Errorf("expected 0 matched, got %d", len(processed))
		}
		if len(warnings) != 0 {
			t.Errorf("expected 0 warnings, got %d", len(warnings))
		}
	})

	t.Run("file_path must also match", func(t *testing.T) {
		// Same selected_text, different file_path — should not match
		fixture := &AutoFixture{
			Comments: []FixtureComment{
				{FilePath: "wrong.md", SelectedText: "hello world"},
			},
		}

		processed, warnings := MatchFixtureComments(fixture, comments)
		if len(processed) != 0 {
			t.Errorf("expected 0 matched (wrong file_path), got %d", len(processed))
		}
		if len(warnings) != 1 {
			t.Errorf("expected 1 warning, got %d", len(warnings))
		}
	})
}

func TestPromptSnapshot(t *testing.T) {
	fixture := &AutoFixture{
		Branch: "136-revise-comments",
		Comments: []FixtureComment{
			{FilePath: "specledger/136-revise-comments/spec.md", SelectedText: "selected text", Guidance: "be concise"},
		},
	}

	comments := []ReviewComment{
		{
			ID:           "snap-1",
			FilePath:     "specledger/136-revise-comments/spec.md",
			SelectedText: "selected text",
			Content:      "This is unclear.",
		},
	}

	matched, warnings := MatchFixtureComments(fixture, comments)
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}

	ctx := BuildRevisionContext("136-revise-comments", matched, nil)
	prompt, err := RenderPrompt(ctx)
	if err != nil {
		t.Fatalf("RenderPrompt error: %v", err)
	}

	goldenPath := filepath.Join("testdata", "snapshot_prompt.golden")

	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.MkdirAll(filepath.Dir(goldenPath), 0755); err != nil {
			t.Fatalf("failed to create testdata dir: %v", err)
		}
		if err := os.WriteFile(goldenPath, []byte(prompt), 0600); err != nil {
			t.Fatalf("failed to write golden file: %v", err)
		}
		t.Logf("golden file updated: %s", goldenPath)
		return
	}

	golden, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("golden file not found: %s\nRun with UPDATE_GOLDEN=1 to create it.", goldenPath)
	}

	if prompt != string(golden) {
		t.Errorf("prompt does not match golden file.\nGot:\n%s\n\nExpected:\n%s", prompt, string(golden))
	}
}

// writeTempFixture writes content to a temp file and returns its path.
func writeTempFixture(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "fixture-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}
	f.Close()
	return f.Name()
}
