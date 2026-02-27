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

	dsPath := filepath.Join(projectPath, "specledger", "design_system.md")

	if isInteractive {
		var generate bool
		err = huh.NewForm(huh.NewGroup(
			huh.NewConfirm().
				Title("Initialize design system?").
				Value(&generate),
		)).Run()
		if err != nil || !generate {
			return
		}
	}

	scanResult, err := mockup.ScanComponents(projectPath, result.Framework)
	if err != nil {
		ui.PrintWarning(fmt.Sprintf("Could not scan components: %v", err))
		return
	}

	ds := mockup.ScanResultToDesignSystem(scanResult, result.Framework)
	if err := mockup.WriteDesignSystem(dsPath, ds); err != nil {
		ui.PrintWarning(fmt.Sprintf("Could not create design system: %v", err))
		return
	}

	fmt.Printf("%s Scanned %d components\n", ui.Checkmark(), len(ds.Components))
	fmt.Printf("%s Created specledger/design_system.md\n", ui.Checkmark())
}
