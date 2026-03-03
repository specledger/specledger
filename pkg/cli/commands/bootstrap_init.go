package commands

import (
	"fmt"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/specledger/specledger/pkg/cli/mockup"
	"github.com/specledger/specledger/pkg/cli/ui"
)

// initFrontendDesignSystem detects a frontend framework and optionally
// initializes the design system during `sl init`.
// This is a non-fatal operation — init succeeds even if detection fails.
func initFrontendDesignSystem(projectPath string, isInteractive bool) {
	result, err := mockup.DetectFramework(projectPath)
	if err != nil || !result.IsFrontend {
		return
	}

	fmt.Printf("\nDetected frontend framework: %s\n", ui.Bold(result.Framework.String()))

	dsPath := filepath.Join(projectPath, "specledger", "design-system.md")

	if isInteractive {
		var generate bool
		err = huh.NewForm(huh.NewGroup(
			huh.NewConfirm().
				Title("Initialize design system?").
				Description("Extracts global CSS, design tokens, and styling patterns").
				Value(&generate),
		)).Run()
		if err != nil || !generate {
			return
		}
	}

	// Extract global CSS and design tokens
	styleInfo := mockup.ScanStyles(projectPath)
	ds := &mockup.DesignSystem{
		Version:   1,
		Framework: result.Framework,
		Style:     styleInfo,
	}
	if err := mockup.WriteDesignSystem(dsPath, ds); err != nil {
		ui.PrintWarning(fmt.Sprintf("Could not create design system: %v", err))
		return
	}

	fmt.Printf("%s Extracted design tokens\n", ui.Checkmark())
	fmt.Printf("%s Created specledger/design-system.md\n", ui.Checkmark())
}
