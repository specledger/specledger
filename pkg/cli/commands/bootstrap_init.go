package commands

import (
	"fmt"
	"path/filepath"
	"strings"

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

	fmt.Printf("\n%s Detected frontend framework: %s (confidence: %d%%)\n",
		ui.Checkmark(), ui.Bold(result.Framework.String()), result.Confidence)

	dsPath := filepath.Join(projectPath, ".specledger", "memory", "design-system.md")

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

	// Extract global CSS, design tokens, and app structure
	styleInfo := mockup.ScanStyles(projectPath)
	appStructure := mockup.ScanAppStructure(projectPath, result.Framework)
	ds := &mockup.DesignSystem{
		Version:      1,
		Framework:    result.Framework,
		Style:        styleInfo,
		AppStructure: appStructure,
	}
	if err := mockup.WriteDesignSystem(dsPath, ds); err != nil {
		ui.PrintWarning(fmt.Sprintf("Could not create design system: %v", err))
		return
	}

	// Print summary of what was detected
	if styleInfo.CSSFramework != "" {
		fmt.Printf("  CSS: %s (%s)\n", styleInfo.CSSFramework, styleInfo.StylingApproach)
	}
	if len(styleInfo.ComponentLibs) > 0 {
		fmt.Printf("  Libs: %s\n", strings.Join(styleInfo.ComponentLibs, ", "))
	}
	if len(styleInfo.ThemeColors) > 0 {
		fmt.Printf("  Colors: %d tokens\n", len(styleInfo.ThemeColors))
	}
	if len(styleInfo.FontFamilies) > 0 {
		fmt.Printf("  Fonts: %s\n", strings.Join(styleInfo.FontFamilies, "; "))
	}
	if appStructure != nil {
		fmt.Printf("  Router: %s (%d layouts)\n", appStructure.Router, len(appStructure.Layouts))
	}

	fmt.Printf("%s Created .specledger/memory/design-system.md\n", ui.Checkmark())
}
