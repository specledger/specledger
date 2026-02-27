package mockup

import (
	_ "embed"
	"fmt"
	"strings"

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
	outputDir string,
	selectedComponents []Component,
	ds *DesignSystem,
	style *StyleInfo,
) *MockupPromptContext {
	ctx := &MockupPromptContext{
		SpecName:  specName,
		SpecPath:  specPath,
		SpecTitle: specTitle,
		Framework: framework,
		Format:    format,
		OutputDir: outputDir,
	}

	if ds != nil {
		ctx.HasDesignSystem = true
		ctx.ExternalLibs = ds.ExternalLibs
	}

	if style != nil && (style.CSSFramework != "" || style.StylingApproach != "" || len(style.ThemeColors) > 0) {
		ctx.Style = style
		ctx.HasStyle = true
	}

	// Convert Components to PromptComponents and build tree
	for _, c := range selectedComponents {
		pc := PromptComponent{
			Name:        c.Name,
			FilePath:    c.FilePath,
			Description: c.Description,
			IsExternal:  c.IsExternal,
			Library:     c.Library,
		}

		if len(c.Props) > 0 {
			propParts := make([]string, 0, len(c.Props))
			for _, p := range c.Props {
				part := p.Name
				if p.Type != "" {
					part += ": " + p.Type
				}
				propParts = append(propParts, part)
			}
			pc.Props = strings.Join(propParts, ", ")
		}

		ctx.Components = append(ctx.Components, pc)
	}

	// Build ASCII tree for prompt display
	if len(selectedComponents) > 0 {
		ctx.ComponentTree = renderComponentTree(selectedComponents)
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
