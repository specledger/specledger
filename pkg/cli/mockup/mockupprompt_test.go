package mockup

import (
	"strings"
	"testing"
)

func TestBuildMockupPromptContext(t *testing.T) {
	components := []Component{
		{
			Name:     "Button",
			FilePath: "src/components/Button.tsx",
			Props:    []PropInfo{{Name: "variant", Type: "string"}, {Name: "onClick", Type: "func"}},
		},
		{
			Name:       "TextField",
			IsExternal: true,
			Library:    "@mui/material",
		},
	}
	ds := &DesignSystem{
		ExternalLibs: []string{"@mui/material"},
	}

	ctx := BuildMockupPromptContext("042-registration", "specledger/042-registration/spec.md", "User Registration", FrameworkReact, MockupFormatHTML, "specledger/042-registration/mockup/", components, ds, nil)

	if ctx.SpecName != "042-registration" {
		t.Errorf("SpecName = %q, want %q", ctx.SpecName, "042-registration")
	}
	if ctx.SpecPath != "specledger/042-registration/spec.md" {
		t.Errorf("SpecPath = %q, want spec path", ctx.SpecPath)
	}
	if ctx.SpecTitle != "User Registration" {
		t.Errorf("SpecTitle = %q, want %q", ctx.SpecTitle, "User Registration")
	}
	if ctx.Framework != FrameworkReact {
		t.Errorf("Framework = %s, want react", ctx.Framework)
	}
	if ctx.Format != MockupFormatHTML {
		t.Errorf("Format = %s, want html", ctx.Format)
	}
	if !ctx.HasDesignSystem {
		t.Error("expected HasDesignSystem = true")
	}
	if len(ctx.Components) != 2 {
		t.Errorf("Components count = %d, want 2", len(ctx.Components))
	}
	if ctx.Components[0].Props != "variant: string, onClick: func" {
		t.Errorf("Props = %q, unexpected", ctx.Components[0].Props)
	}
	if ctx.ComponentTree == "" {
		t.Error("expected ComponentTree to be non-empty")
	}
}

func TestRenderMockupPrompt_ReactHTML(t *testing.T) {
	ctx := &MockupPromptContext{
		SpecName:        "042-registration",
		SpecPath:        "specledger/042-registration/spec.md",
		SpecTitle:       "User Registration",
		Framework:       FrameworkReact,
		Format:          MockupFormatHTML,
		OutputDir:       "specledger/042-registration/mockup/",
		HasDesignSystem: true,
		ComponentTree:   "└── src/\n    └── components/\n        └── Button.tsx  →  Button (variant: string)\n",
		Components: []PromptComponent{
			{Name: "Button", FilePath: "src/components/Button.tsx", Props: "variant: string"},
		},
		ExternalLibs: []string{"@mui/material"},
	}

	result, err := RenderMockupPrompt(ctx)
	if err != nil {
		t.Fatal(err)
	}

	checks := []string{
		"User Registration",
		"042-registration",
		"spec.md",
		"html",
		"React",
		"Button",
		"@mui/material",
		"fetch",
		"mockup.html",
		"components/",
	}

	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("prompt missing expected content: %q", check)
		}
	}
}

func TestRenderMockupPrompt_ReactJSX(t *testing.T) {
	ctx := &MockupPromptContext{
		SpecName:        "042-registration",
		SpecPath:        "specledger/042-registration/spec.md",
		SpecTitle:       "User Registration",
		Framework:       FrameworkReact,
		Format:          MockupFormatJSX,
		OutputDir:       "specledger/042-registration/mockup/",
		HasDesignSystem: true,
		Components: []PromptComponent{
			{Name: "Form", FilePath: "src/components/Form.tsx"},
		},
	}

	result, err := RenderMockupPrompt(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(result, "jsx") {
		t.Error("JSX prompt should mention jsx format")
	}
	if !strings.Contains(result, "mockup.jsx") {
		t.Error("JSX prompt should mention mockup.jsx")
	}
}

func TestRenderMockupPrompt_VueHTML(t *testing.T) {
	ctx := &MockupPromptContext{
		SpecName:        "050-dashboard",
		SpecPath:        "specledger/050-dashboard/spec.md",
		SpecTitle:       "Dashboard Feature",
		Framework:       FrameworkVue,
		Format:          MockupFormatHTML,
		OutputDir:       "specledger/050-dashboard/mockup/",
		HasDesignSystem: false,
	}

	result, err := RenderMockupPrompt(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(result, "Vue") {
		t.Error("Vue prompt should mention Vue")
	}
	if !strings.Contains(result, "Dashboard Feature") {
		t.Error("prompt should include spec title")
	}
}

func TestRenderMockupPrompt_EmptyComponents(t *testing.T) {
	ctx := &MockupPromptContext{
		SpecName:        "060-settings",
		SpecPath:        "specledger/060-settings/spec.md",
		SpecTitle:       "Settings Page",
		Framework:       FrameworkReact,
		Format:          MockupFormatHTML,
		OutputDir:       "specledger/060-settings/mockup/",
		HasDesignSystem: false,
		Components:      nil,
	}

	result, err := RenderMockupPrompt(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(result, "Settings Page") {
		t.Error("prompt should include spec title even with no components")
	}
	if !strings.Contains(result, "components/") {
		t.Error("prompt should always include components/ folder structure")
	}
}
