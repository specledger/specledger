package mockup

import (
	"strings"
	"testing"
)

func TestBuildMockupPromptContext(t *testing.T) {
	ds := &DesignSystem{
		ExternalLibs: []string{"@mui/material"},
	}
	style := &StyleInfo{
		CSSFramework:    "Tailwind CSS",
		StylingApproach: "utility-first",
		ThemeColors: map[string]string{
			"--primary": "#3b82f6",
		},
	}

	ctx := BuildMockupPromptContext("042-registration", "specledger/042-registration/spec.md", "User Registration", FrameworkReact, MockupFormatHTML, "specledger/042-registration/mockup.html", ds, style)

	if ctx.SpecName != "042-registration" {
		t.Errorf("SpecName = %q, want %q", ctx.SpecName, "042-registration")
	}
	if ctx.SpecPath != "specledger/042-registration/spec.md" {
		t.Errorf("SpecPath = %q, want spec path", ctx.SpecPath)
	}
	if ctx.SpecTitle != "User Registration" {
		t.Errorf("SpecTitle = %q, want %q", ctx.SpecTitle, "User Registration")
	}
	if ctx.OutputPath != "specledger/042-registration/mockup.html" {
		t.Errorf("OutputPath = %q, unexpected", ctx.OutputPath)
	}
	if !ctx.HasDesignSystem {
		t.Error("expected HasDesignSystem = true")
	}
	if !ctx.HasStyle {
		t.Error("expected HasStyle = true")
	}
	if len(ctx.ExternalLibs) != 1 {
		t.Errorf("ExternalLibs count = %d, want 1", len(ctx.ExternalLibs))
	}
}

func TestRenderMockupPrompt_ReactHTML(t *testing.T) {
	ctx := &MockupPromptContext{
		SpecName:        "042-registration",
		SpecPath:        "specledger/042-registration/spec.md",
		SpecTitle:       "User Registration",
		Framework:       FrameworkReact,
		Format:          MockupFormatHTML,
		OutputPath:      "specledger/042-registration/mockup.html",
		HasDesignSystem: true,
		ExternalLibs:    []string{"@mui/material"},
		HasStyle:        true,
		Style: &StyleInfo{
			CSSFramework:    "Tailwind CSS",
			StylingApproach: "utility-first",
		},
	}

	result, err := RenderMockupPrompt(ctx)
	if err != nil {
		t.Fatal(err)
	}

	checks := []string{
		"User Registration",
		"042-registration",
		"spec.md",
		"mockup.html",
		"React",
		"@mui/material",
		"self-contained",
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
		OutputPath:      "specledger/042-registration/mockup.jsx",
		HasDesignSystem: true,
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
		OutputPath:      "specledger/050-dashboard/mockup.html",
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

func TestRenderMockupPrompt_NoDesignSystem(t *testing.T) {
	ctx := &MockupPromptContext{
		SpecName:        "060-settings",
		SpecPath:        "specledger/060-settings/spec.md",
		SpecTitle:       "Settings Page",
		Framework:       FrameworkReact,
		Format:          MockupFormatHTML,
		OutputPath:      "specledger/060-settings/mockup.html",
		HasDesignSystem: false,
	}

	result, err := RenderMockupPrompt(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(result, "Settings Page") {
		t.Error("prompt should include spec title even with no design system")
	}
	if !strings.Contains(result, "mockup.html") {
		t.Error("prompt should include output path")
	}
}
