package mockup

import (
	_ "embed"
	"fmt"

	"github.com/specledger/specledger/pkg/cli/prompt"
)

//go:embed prompt.tmpl
var mockupPromptTemplate string

// BuildMockupPromptContext assembles a MockupPromptContext from the gathered data.
// The prompt template instructs the AI agent to read design-system.md directly,
// so style/design-system data is not embedded in the prompt context.
func BuildMockupPromptContext(
	specName string,
	specPath string,
	specTitle string,
	framework FrameworkType,
	format MockupFormat,
	outputPath string,
	userPrompt string,
) *MockupPromptContext {
	return &MockupPromptContext{
		SpecName:   specName,
		SpecPath:   specPath,
		SpecTitle:  specTitle,
		Framework:  framework,
		Format:     format,
		OutputPath: outputPath,
		UserPrompt: userPrompt,
	}
}

// RenderMockupPrompt renders the mockup prompt template with the given context.
func RenderMockupPrompt(ctx *MockupPromptContext) (string, error) {
	if mockupPromptTemplate == "" {
		return "", fmt.Errorf("mockup prompt template is empty")
	}

	return prompt.RenderTemplate("mockup-prompt", mockupPromptTemplate, ctx)
}
