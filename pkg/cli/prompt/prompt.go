package prompt

import (
	"bytes"
	"fmt"
	"math"
	"text/template"
)

// RenderTemplate renders a Go text/template with the given data.
// The templateName is used for error messages, templateContent is the raw template string.
func RenderTemplate(templateName, templateContent string, data any) (string, error) {
	tmpl, err := template.New(templateName).Parse(templateContent)
	if err != nil {
		return "", fmt.Errorf("failed to parse template %q: %w", templateName, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to render template %q: %w", templateName, err)
	}

	return buf.String(), nil
}

// EstimateTokens estimates the number of tokens in text using the ~3.5 chars/token heuristic.
// Accuracy is within ~20% for typical English text.
func EstimateTokens(text string) int {
	if len(text) == 0 {
		return 0
	}
	return int(math.Ceil(float64(len(text)) / 3.5))
}

// PrintTokenWarnings prints warnings when the prompt is too short or too long.
func PrintTokenWarnings(tokens int) {
	switch {
	case tokens < 100:
		fmt.Println("  Prompt is very short — the agent may lack context.")
	case tokens > 8000:
		fmt.Printf("  Prompt is ~%d tokens — this may reduce agent effectiveness.\n", tokens)
	default:
		fmt.Printf("  Estimated prompt size: ~%d tokens\n", tokens)
	}
}
