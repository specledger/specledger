package mockup

import (
	"strings"
	"testing"
)

func TestBuildMockupPromptContext(t *testing.T) {
	ctx := BuildMockupPromptContext("042-registration", "specledger/042-registration/spec.md", "User Registration", FrameworkReact, MockupFormatHTML, "specledger/042-registration/mockup.html", "")

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
	if ctx.Framework != FrameworkReact {
		t.Errorf("Framework = %q, want react", ctx.Framework)
	}
}

func TestRenderMockupPrompt_ReactHTML(t *testing.T) {
	ctx := &MockupPromptContext{
		SpecName:   "042-registration",
		SpecPath:   "specledger/042-registration/spec.md",
		SpecTitle:  "User Registration",
		Framework:  FrameworkReact,
		Format:     MockupFormatHTML,
		OutputPath: "specledger/042-registration/mockup.html",
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
		"design-system.md",
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
		SpecName:   "042-registration",
		SpecPath:   "specledger/042-registration/spec.md",
		SpecTitle:  "User Registration",
		Framework:  FrameworkReact,
		Format:     MockupFormatJSX,
		OutputPath: "specledger/042-registration/mockup.jsx",
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
		SpecName:   "050-dashboard",
		SpecPath:   "specledger/050-dashboard/spec.md",
		SpecTitle:  "Dashboard Feature",
		Framework:  FrameworkVue,
		Format:     MockupFormatHTML,
		OutputPath: "specledger/050-dashboard/mockup.html",
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

func TestRenderMockupPrompt_WithUserPrompt(t *testing.T) {
	ctx := &MockupPromptContext{
		SpecName:   "060-settings",
		SpecPath:   "specledger/060-settings/spec.md",
		SpecTitle:  "Settings Page",
		Framework:  FrameworkReact,
		Format:     MockupFormatHTML,
		OutputPath: "specledger/060-settings/mockup.html",
		UserPrompt: "Use dark theme",
	}

	result, err := RenderMockupPrompt(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(result, "Use dark theme") {
		t.Error("prompt should include user prompt")
	}
}
