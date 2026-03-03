package mockup

import (
	_ "embed"
	"fmt"

	"github.com/specledger/specledger/pkg/cli/prompt"
)

//go:embed prompt.tmpl
var mockupPromptTemplate string

// BuildMockupPromptContext assembles a MockupPromptContext from the gathered data.
func BuildMockupPromptContext(
	specName string,
	specPath string,
	specTitle string,
	framework FrameworkType,
	format MockupFormat,
	outputPath string,
	ds *DesignSystem,
	style *StyleInfo,
	userPrompt string,
) *MockupPromptContext {
	ctx := &MockupPromptContext{
		SpecName:   specName,
		SpecPath:   specPath,
		SpecTitle:  specTitle,
		Framework:  framework,
		Format:     format,
		OutputPath: outputPath,
		UserPrompt: userPrompt,
	}

	if ds != nil {
		ctx.HasDesignSystem = true
		ctx.ExternalLibs = ds.ExternalLibs
	}

	if style != nil && (style.CSSFramework != "" || style.StylingApproach != "" || len(style.ThemeColors) > 0) {
		ctx.Style = style
		ctx.HasStyle = true
	}

	return ctx
}

// RenderMockupPrompt renders the mockup prompt template with the given context.
func RenderMockupPrompt(ctx *MockupPromptContext) (string, error) {
	if mockupPromptTemplate == "" {
		return "", fmt.Errorf("mockup prompt template is empty")
	}

	return prompt.RenderTemplate("mockup-prompt", mockupPromptTemplate, ctx)
}
