package revise

import (
	"strings"
	"testing"
)

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"empty", "", 0},
		{"3 chars", "abc", 1},                        // ceil(3/3.5) = ceil(0.857) = 1
		{"7 chars", "1234567", 2},                    // ceil(7/3.5) = ceil(2.0) = 2
		{"100 chars", strings.Repeat("a", 100), 29},  // ceil(100/3.5) = ceil(28.571) = 29
		{"350 chars", strings.Repeat("a", 350), 100}, // ceil(350/3.5) = 100 exactly
		{"351 chars", strings.Repeat("a", 351), 101}, // ceil(351/3.5) = ceil(100.28) = 101
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EstimateTokens(tt.input)
			if got != tt.want {
				t.Errorf("EstimateTokens(%d chars): got %d, want %d", len(tt.input), got, tt.want)
			}
		})
	}
}

func TestRenderPrompt(t *testing.T) {
	ctx := RevisionContext{
		SpecKey: "136-revise-comments",
		Comments: []PromptComment{
			{
				Index:    1,
				ID:       "uuid-1",
				FilePath: "specledger/136-revise-comments/spec.md",
				Target:   "selected text here",
				Feedback: "This needs clarification.",
				Guidance: "Be concise.",
			},
			{
				Index:    2,
				ID:       "uuid-2",
				FilePath: "specledger/136-revise-comments/plan.md",
				Target:   "",
				Feedback: "General feedback.",
				Guidance: "",
			},
		},
	}

	prompt, err := RenderPrompt(ctx)
	if err != nil {
		t.Fatalf("RenderPrompt returned error: %v", err)
	}

	checks := []struct {
		section string
		want    string
	}{
		{"spec key", "136-revise-comments"},
		{"file path 1", "specledger/136-revise-comments/spec.md"},
		{"file path 2", "specledger/136-revise-comments/plan.md"},
		{"feedback 1", "This needs clarification."},
		{"feedback 2", "General feedback."},
		{"guidance", "Be concise."},
		{"target quoted", `"selected text here"`},
		{"general feedback fallback", "General feedback"},
		{"comment 1 header", "### Comment 1"},
		{"comment 2 header", "### Comment 2"},
		{"artifacts section", "## Artifacts to Revise"},
		{"comments section", "## Comments to Address"},
		{"revision strategy", "## Revision Strategy"},
		{"thematic clusters", "thematic clusters"},
	}

	for _, c := range checks {
		if !strings.Contains(prompt, c.want) {
			t.Errorf("RenderPrompt output missing %s: expected to contain %q\nGot:\n%s", c.section, c.want, prompt)
		}
	}
}

func TestRenderPrompt_EmptyTarget(t *testing.T) {
	ctx := RevisionContext{
		SpecKey: "test",
		Comments: []PromptComment{
			{Index: 1, FilePath: "spec.md", Target: "", Feedback: "feedback"},
		},
	}

	prompt, err := RenderPrompt(ctx)
	if err != nil {
		t.Fatalf("RenderPrompt returned error: %v", err)
	}

	if !strings.Contains(prompt, "General feedback") {
		t.Errorf("expected 'General feedback' fallback when Target is empty, got:\n%s", prompt)
	}
}

func TestBuildRevisionContext(t *testing.T) {
	line := 10

	processed := []ProcessedComment{
		{
			Comment: ReviewComment{
				ID:           "id-1",
				FilePath:     "spec.md",
				Content:      "Fix this.",
				SelectedText: "some text",
			},
			Guidance: "keep it short",
			Index:    1,
		},
		{
			Comment: ReviewComment{
				ID:       "id-2",
				FilePath: "plan.md",
				Content:  "General comment.",
				Line:     &line,
			},
			Guidance: "",
			Index:    2,
		},
		{
			Comment: ReviewComment{
				ID:       "id-3",
				FilePath: "tasks.md",
				Content:  "No text, no line.",
			},
			Guidance: "",
			Index:    3,
		},
	}

	ctx := BuildRevisionContext("test-spec", processed, nil)

	if ctx.SpecKey != "test-spec" {
		t.Errorf("SpecKey: got %q, want %q", ctx.SpecKey, "test-spec")
	}
	if len(ctx.Comments) != 3 {
		t.Fatalf("Comments length: got %d, want 3", len(ctx.Comments))
	}

	// Comment 1: has selected_text → Target = selected_text
	if ctx.Comments[0].Target != "some text" {
		t.Errorf("Comments[0].Target: got %q, want %q", ctx.Comments[0].Target, "some text")
	}
	if ctx.Comments[0].Guidance != "keep it short" {
		t.Errorf("Comments[0].Guidance: got %q, want %q", ctx.Comments[0].Guidance, "keep it short")
	}

	// Comment 2: no selected_text, has Line → Target = "Line 10"
	if ctx.Comments[1].Target != "Line 10" {
		t.Errorf("Comments[1].Target: got %q, want %q", ctx.Comments[1].Target, "Line 10")
	}

	// Comment 3: no selected_text, no line → Target = ""
	if ctx.Comments[2].Target != "" {
		t.Errorf("Comments[2].Target: got %q, want empty string", ctx.Comments[2].Target)
	}
}

func TestBuildRevisionContext_Truncation(t *testing.T) {
	longText := strings.Repeat("a", 250)
	processed := []ProcessedComment{
		{
			Comment: ReviewComment{
				ID:           "id-1",
				FilePath:     "spec.md",
				Content:      "feedback",
				SelectedText: longText,
			},
			Index: 1,
		},
	}

	ctx := BuildRevisionContext("spec", processed, nil)

	if len(ctx.Comments[0].Target) != 200 {
		t.Errorf("Target truncation: got len %d, want 200", len(ctx.Comments[0].Target))
	}
	if !strings.HasSuffix(ctx.Comments[0].Target, "...") {
		t.Errorf("Target should end with '...', got suffix %q", ctx.Comments[0].Target[197:])
	}
}

func TestBuildRevisionContext_WithReplies(t *testing.T) {
	processed := []ProcessedComment{
		{
			Comment: ReviewComment{
				ID:           "parent-1",
				FilePath:     "spec.md",
				Content:      "Fix this section.",
				SelectedText: "some text",
			},
			Guidance: "",
			Index:    1,
		},
		{
			Comment: ReviewComment{
				ID:           "parent-2",
				FilePath:     "plan.md",
				Content:      "Update approach.",
				SelectedText: "other text",
			},
			Guidance: "",
			Index:    2,
		},
	}

	replies := []ReviewComment{
		{
			ID:              "reply-1",
			ParentCommentID: "parent-1",
			AuthorName:      "alice",
			Content:         "Also affects user story 2",
			CreatedAt:       "2026-01-01T00:00:00Z",
		},
		{
			ID:              "reply-2",
			ParentCommentID: "parent-1",
			AuthorName:      "bob",
			Content:         "Agreed, needs coordinated fix",
			CreatedAt:       "2026-01-01T01:00:00Z",
		},
	}

	ctx := BuildRevisionContext("test-spec", processed, replies)

	// Parent-1 should have 2 replies
	if len(ctx.Comments[0].Replies) != 2 {
		t.Fatalf("Comments[0].Replies length: got %d, want 2", len(ctx.Comments[0].Replies))
	}
	if ctx.Comments[0].Replies[0].AuthorName != "alice" {
		t.Errorf("Replies[0].AuthorName: got %q, want %q", ctx.Comments[0].Replies[0].AuthorName, "alice")
	}
	if ctx.Comments[0].Replies[0].ID != "reply-1" {
		t.Errorf("Replies[0].ID: got %q, want %q", ctx.Comments[0].Replies[0].ID, "reply-1")
	}
	if ctx.Comments[0].Replies[1].Content != "Agreed, needs coordinated fix" {
		t.Errorf("Replies[1].Content: got %q", ctx.Comments[0].Replies[1].Content)
	}

	// Parent-2 should have no replies
	if len(ctx.Comments[1].Replies) != 0 {
		t.Errorf("Comments[1].Replies length: got %d, want 0", len(ctx.Comments[1].Replies))
	}
}

func TestRenderPrompt_WithThreads(t *testing.T) {
	ctx := RevisionContext{
		SpecKey: "test-threads",
		Comments: []PromptComment{
			{
				Index:    1,
				ID:       "uuid-1",
				FilePath: "spec.md",
				Target:   "some target",
				Feedback: "Fix this section.",
				Replies: []ThreadReply{
					{ID: "r1", AuthorName: "alice", Content: "Also affects story 2"},
					{ID: "r2", AuthorName: "bob", Content: "Agreed"},
				},
			},
		},
	}

	prompt, err := RenderPrompt(ctx)
	if err != nil {
		t.Fatalf("RenderPrompt returned error: %v", err)
	}

	checks := []string{
		"**Thread:**",
		"**alice**",
		"Also affects story 2",
		"**bob**",
		"Agreed",
		"Revision Strategy",
		"thematic clusters",
	}

	for _, want := range checks {
		if !strings.Contains(prompt, want) {
			t.Errorf("RenderPrompt output missing %q\nGot:\n%s", want, prompt)
		}
	}
}

func TestBuildReplyMap(t *testing.T) {
	replies := []ReviewComment{
		{ID: "r1", ParentCommentID: "p1", Content: "reply 1"},
		{ID: "r2", ParentCommentID: "p1", Content: "reply 2"},
		{ID: "r3", ParentCommentID: "p2", Content: "reply 3"},
	}

	m := BuildReplyMap(replies)

	if len(m["p1"]) != 2 {
		t.Errorf("p1 replies: got %d, want 2", len(m["p1"]))
	}
	if len(m["p2"]) != 1 {
		t.Errorf("p2 replies: got %d, want 1", len(m["p2"]))
	}
	if len(m["p3"]) != 0 {
		t.Errorf("p3 replies: got %d, want 0", len(m["p3"]))
	}
}

func TestBuildReplyMap_Nil(t *testing.T) {
	m := BuildReplyMap(nil)
	if m == nil {
		t.Error("BuildReplyMap(nil) should return non-nil empty map")
	}
	if len(m) != 0 {
		t.Errorf("BuildReplyMap(nil) should return empty map, got %d entries", len(m))
	}
}
