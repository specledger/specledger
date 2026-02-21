package revise

import (
	_ "embed"
	"bytes"
	"fmt"
	"math"
	"text/template"
)

//go:embed prompt.tmpl
var promptTemplate string

// RenderPrompt renders the revision prompt template with the given context.
func RenderPrompt(ctx RevisionContext) (string, error) {
	tmpl, err := template.New("revision").Parse(promptTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse prompt template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return "", fmt.Errorf("failed to render prompt template: %w", err)
	}

	return buf.String(), nil
}

// EstimateTokens estimates the number of tokens in text using the ~3.5 chars/token heuristic.
// See plan.md §6 and research.md R12 for the rationale. Accuracy is within ~20%.
func EstimateTokens(text string) int {
	return int(math.Ceil(float64(len(text)) / 3.5))
}

// BuildRevisionContext converts processed comments into the template rendering context.
func BuildRevisionContext(specKey string, processed []ProcessedComment) RevisionContext {
	comments := make([]PromptComment, 0, len(processed))
	for _, p := range processed {
		target := p.Comment.SelectedText
		if target == "" {
			if p.Comment.Line != nil {
				target = fmt.Sprintf("Line %d", *p.Comment.Line)
			} else {
				target = ""
			}
		}
		// Truncate very long selected_text for the prompt
		if len(target) > 200 {
			target = target[:197] + "..."
		}

		comments = append(comments, PromptComment{
			Index:    p.Index,
			ID:       p.Comment.ID,
			FilePath: p.Comment.FilePath,
			Target:   target,
			Feedback: p.Comment.Content,
			Guidance: p.Guidance,
		})
	}

	return RevisionContext{
		SpecKey:  specKey,
		Comments: comments,
	}
}

// PrintTokenWarnings prints warnings when the prompt is too short or too long.
func PrintTokenWarnings(tokens int) {
	switch {
	case tokens < 100:
		fmt.Println("⚠  Prompt is very short — the agent may lack context.")
	case tokens > 8000:
		fmt.Printf("⚠  Prompt is ~%d tokens — this may reduce agent effectiveness.\n", tokens)
	default:
		fmt.Printf("   Estimated prompt size: ~%d tokens\n", tokens)
	}
}
