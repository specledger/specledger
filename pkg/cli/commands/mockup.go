package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	cligit "github.com/specledger/specledger/pkg/cli/git"
	"github.com/specledger/specledger/pkg/cli/launcher"
	"github.com/specledger/specledger/pkg/cli/mockup"
	"github.com/specledger/specledger/pkg/cli/prompt"
	"github.com/specledger/specledger/pkg/cli/ui"
	"github.com/specledger/specledger/pkg/issues"
	"github.com/spf13/cobra"
)

// VarMockupCmd is the sl mockup command.
var VarMockupCmd = &cobra.Command{
	Use:   "mockup [spec-name]",
	Short: "Generate UI mockups from feature specifications",
	Long: `Generate UI mockups from feature specifications using an interactive flow.

Flow:
  1. Resolve spec (from arg, branch, or picker)
  2. Detect frontend framework
  3. Check/generate design system
  4. Select components for mockup
  5. Choose output format (html/jsx)
  6. Generate and review prompt
  7. Launch AI agent to generate mockup
  8. Commit and push changes

Examples:
  sl mockup                              # Auto-detect spec from branch
  sl mockup 042-user-registration        # Explicit spec name
  sl mockup --format jsx                 # Generate JSX mockup
  sl mockup --dry-run                    # Write prompt to file, skip agent
  sl mockup --force                      # Bypass frontend detection check
  sl mockup --json                       # Non-interactive JSON output`,
	Args:         cobra.MaximumNArgs(1),
	RunE:         runMockup,
	SilenceUsage: true,
}

var mockupUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Refresh the design system index by rescanning the codebase",
	Long: `Refresh the design system index by rescanning the codebase.

Rescans project components and updates specledger/design_system.md.
Manual entries are preserved across updates.

Examples:
  sl mockup update          # Interactive update
  sl mockup update --json   # Output as JSON`,
	RunE:         runMockupUpdate,
	SilenceUsage: true,
}

var (
	mockupFormat  string
	mockupForce   bool
	mockupDryRun  bool
	mockupSummary bool
	mockupJSON    bool
	updateJSON    bool
)

func init() {
	VarMockupCmd.Flags().StringVar(&mockupFormat, "format", "html", "Output format: html or jsx")
	VarMockupCmd.Flags().BoolVarP(&mockupForce, "force", "f", false, "Bypass frontend detection check")
	VarMockupCmd.Flags().BoolVar(&mockupDryRun, "dry-run", false, "Write prompt to file instead of launching agent")
	VarMockupCmd.Flags().BoolVar(&mockupSummary, "summary", false, "Compact output for agent/CI integration")
	VarMockupCmd.Flags().BoolVar(&mockupJSON, "json", false, "Non-interactive path, output result as JSON")

	mockupUpdateCmd.Flags().BoolVar(&updateJSON, "json", false, "Output result as JSON")

	VarMockupCmd.AddCommand(mockupUpdateCmd)
}

func runMockup(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Validate format
	format := mockup.MockupFormat(mockupFormat)
	if !format.IsValid() {
		return fmt.Errorf("Error: Invalid format '%s'\n\nSupported formats: html, jsx", mockupFormat)
	}

	// Step 1: Resolve spec
	specName, err := resolveSpec(cwd, args)
	if err != nil {
		return err
	}

	specDir := filepath.Join(cwd, "specledger", specName)
	specFile := filepath.Join(specDir, "spec.md")
	if _, err := os.Stat(specFile); os.IsNotExist(err) {
		return fmt.Errorf("Error: Spec '%s' not found\n\nNo spec.md file at specledger/%s/spec.md\nCreate a spec first with: sl specify %s", specName, specName, specName)
	}

	// Step 2: Framework detection
	fmt.Println("Detecting frontend framework...")
	detection, err := mockup.DetectFramework(cwd)
	if err != nil {
		return fmt.Errorf("framework detection failed: %w", err)
	}

	if !detection.IsFrontend && !mockupForce {
		return fmt.Errorf("Error: Not a frontend project\n\nNo frontend framework detected in this repository.\nUse --force to bypass this check, or run from a frontend project directory.")
	}

	if detection.IsFrontend {
		headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
		fmt.Printf("%s Detected: %s (confidence: %d%%)\n",
			ui.Checkmark(),
			headerStyle.Render(detection.Framework.String()),
			detection.Confidence)

		if !mockupForce && !mockupJSON {
			var confirmed bool
			err = huh.NewForm(huh.NewGroup(
				huh.NewConfirm().
					Title("Confirm framework?").
					Value(&confirmed),
			)).Run()
			if err != nil {
				return fmt.Errorf("framework confirmation: %w", err)
			}
			if !confirmed {
				fmt.Println("Cancelled.")
				return nil
			}
		}
	}

	framework := detection.Framework
	if mockupForce && !detection.IsFrontend {
		framework = mockup.FrameworkUnknown
	}

	// Step 3: Design system check/generate
	dsPath := filepath.Join(cwd, "specledger", "design_system.md")
	var ds *mockup.DesignSystem
	dsCreated := false

	if _, err := os.Stat(dsPath); os.IsNotExist(err) {
		fmt.Println("Design system not found.")
		if !mockupJSON {
			var generate bool
			err = huh.NewForm(huh.NewGroup(
				huh.NewConfirm().
					Title("Generate design system now?").
					Value(&generate),
			)).Run()
			if err != nil {
				return fmt.Errorf("design system prompt: %w", err)
			}
			if !generate {
				fmt.Println("Skipping design system generation.")
				ds = &mockup.DesignSystem{
					Version:   1,
					Framework: framework,
				}
			}
		}
		if ds == nil {
			scanResult, err := mockup.ScanComponents(cwd, framework)
			if err != nil {
				return fmt.Errorf("component scan failed: %w", err)
			}
			ds = mockup.ScanResultToDesignSystem(scanResult, framework)
			if err := mockup.WriteDesignSystem(dsPath, ds); err != nil {
				return fmt.Errorf("Error: Cannot write to specledger/\n\nCheck file permissions and try again.")
			}
			fmt.Printf("%s Scanned %d components\n", ui.Checkmark(), len(ds.Components))
			fmt.Printf("%s Created specledger/design_system.md\n", ui.Checkmark())
			dsCreated = true
		}
	} else {
		loadedDS, err := mockup.LoadDesignSystem(dsPath)
		if err != nil {
			fmt.Printf("%s Design system is malformed, regenerating...\n", ui.WarningIcon())
			scanResult, scanErr := mockup.ScanComponents(cwd, framework)
			if scanErr != nil {
				return fmt.Errorf("component scan failed: %w", scanErr)
			}
			ds = mockup.ScanResultToDesignSystem(scanResult, framework)
			if writeErr := mockup.WriteDesignSystem(dsPath, ds); writeErr != nil {
				return fmt.Errorf("failed to write design system: %w", writeErr)
			}
			dsCreated = true
		} else {
			ds = loadedDS
			fmt.Printf("%s Loaded design system (%d components)\n", ui.Checkmark(), len(ds.Components))
		}
	}

	allComponents := append(ds.Components, ds.ManualEntries...)

	// Step 4: Component selection
	var selectedComponents []mockup.Component
	if len(allComponents) > 0 && !mockupJSON {
		options := make([]huh.Option[string], 0, len(allComponents))
		for _, c := range allComponents {
			label := c.Name
			if c.IsExternal {
				label = fmt.Sprintf("%s (%s)", c.Name, c.Library)
			} else if c.FilePath != "" {
				label = fmt.Sprintf("%s (%s)", c.Name, c.FilePath)
			}
			options = append(options, huh.NewOption(label, c.Name).Selected(true))
		}

		selectedNames := make([]string, 0, len(allComponents))
		err = huh.NewForm(huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select components to include in mockup").
				Options(options...).
				Value(&selectedNames),
		)).Run()
		if err != nil {
			return fmt.Errorf("component selection: %w", err)
		}

		nameSet := make(map[string]struct{}, len(selectedNames))
		for _, n := range selectedNames {
			nameSet[n] = struct{}{}
		}
		for _, c := range allComponents {
			if _, ok := nameSet[c.Name]; ok {
				selectedComponents = append(selectedComponents, c)
			}
		}
	} else {
		selectedComponents = allComponents
	}

	// Step 5: Format selection
	if !mockupJSON && mockupFormat == "html" {
		var formatChoice string
		err = huh.NewForm(huh.NewGroup(
			huh.NewSelect[string]().
				Title("Output format").
				Options(
					huh.NewOption("html", "html"),
					huh.NewOption("jsx", "jsx"),
				).
				Value(&formatChoice),
		)).Run()
		if err != nil {
			return fmt.Errorf("format selection: %w", err)
		}
		format = mockup.MockupFormat(formatChoice)
	}

	// Step 6: Generate prompt
	fmt.Println("\nGenerating prompt...")
	specContent, err := mockup.ParseSpec(specFile)
	if err != nil {
		return err
	}

	if len(specContent.UserStories) == 0 {
		return fmt.Errorf("Error: Spec has no user scenarios\n\nThe spec.md file has no user scenarios to generate mockups from.\nAdd user scenarios with: sl clarify %s", specName)
	}

	// Always use mockup/ folder — the agent decides how to split
	outputDir := filepath.Join("specledger", specName, "mockup")
	fullMockupDir := filepath.Join(cwd, outputDir)

	// Check for existing mockup
	if _, err := os.Stat(fullMockupDir); err == nil && !mockupJSON {
		var overwrite bool
		err = huh.NewForm(huh.NewGroup(
			huh.NewConfirm().
				Title(fmt.Sprintf("Mockup already exists at %s/\nOverwrite?", outputDir)).
				Value(&overwrite),
		)).Run()
		if err != nil {
			return fmt.Errorf("overwrite confirmation: %w", err)
		}
		if !overwrite {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	// Create mockup directory structure
	componentsDir := filepath.Join(fullMockupDir, "components")
	if err := os.MkdirAll(componentsDir, 0755); err != nil {
		return fmt.Errorf("failed to create mockup directory: %w", err)
	}

	// Scan project styling patterns
	styleInfo := mockup.ScanStyles(cwd)

	promptCtx := mockup.BuildMockupPromptContext(specName, specFile, specContent.Title, framework, format, outputDir+"/", selectedComponents, ds, styleInfo)
	promptText, err := mockup.RenderMockupPrompt(promptCtx)
	if err != nil {
		return fmt.Errorf("failed to render prompt: %w", err)
	}

	tokens := prompt.EstimateTokens(promptText)
	fmt.Printf("%s Prompt generated (estimated %d tokens)\n", ui.Checkmark(), tokens)
	prompt.PrintTokenWarnings(tokens)

	// --dry-run: write prompt to file and exit
	if mockupDryRun {
		promptPath := filepath.Join(specDir, "mockup-prompt.md")
		if err := os.WriteFile(promptPath, []byte(promptText), 0600); err != nil {
			return fmt.Errorf("Error: Cannot write to specledger/\n\nCheck file permissions and try again.")
		}
		fmt.Printf("\n%s Prompt written to specledger/%s/mockup-prompt.md\n", ui.Checkmark(), specName)
		fmt.Println("  Run your agent manually with this prompt, or re-run without --dry-run.")

		if mockupJSON {
			result := mockup.MockupResult{
				Status:              "success",
				Framework:           string(framework),
				SpecName:            specName,
				PromptPath:          filepath.Join("specledger", specName, "mockup-prompt.md"),
				Format:              string(format),
				DesignSystemCreated: dsCreated,
				ComponentsScanned:   len(ds.Components),
				ComponentsSelected:  len(selectedComponents),
				AgentLaunched:       false,
				Committed:           false,
			}
			return printJSON(result)
		}
		return nil
	}

	// Step 7: Edit & confirm prompt
	finalPrompt, err := mockupEditAndConfirm(promptText)
	if err != nil {
		return err
	}
	if finalPrompt == "" {
		return nil // user cancelled or wrote to file
	}

	// Step 8: Launch agent
	agentCmd := os.Getenv("SPECLEDGER_AGENT")
	var agentOpt launcher.AgentOption
	if agentCmd != "" {
		agentOpt = launcher.AgentOption{Name: agentCmd, Command: agentCmd}
	} else {
		for _, a := range launcher.DefaultAgents {
			if a.Command == "" {
				continue
			}
			al := launcher.NewAgentLauncher(a, cwd)
			if al.IsAvailable() {
				agentOpt = a
				break
			}
		}
	}

	al := launcher.NewAgentLauncher(agentOpt, cwd)
	agentLaunched := false
	if !al.IsAvailable() {
		promptPath := filepath.Join(specDir, "mockup-prompt.md")
		if err := os.WriteFile(promptPath, []byte(finalPrompt), 0600); err != nil {
			return fmt.Errorf("failed to write prompt: %w", err)
		}
		fmt.Printf("Error: No AI agent available\n\nPrompt written to specledger/%s/mockup-prompt.md\nInstall Claude Code: npm install -g @anthropic-ai/claude-code\n", specName)
	} else {
		fmt.Printf("Launching %s...\n", al.Name)
		if err := al.LaunchWithPrompt(finalPrompt); err != nil {
			return fmt.Errorf("agent exited with error: %w", err)
		}
		agentLaunched = true
	}

	// Step 9: Post-agent commit/push flow
	committed := false
	if agentLaunched {
		changesAfterAgent, err := cligit.HasUncommittedChanges(cwd)
		if err != nil {
			return fmt.Errorf("failed to check git status after agent: %w", err)
		}

		if changesAfterAgent {
			committed, err = mockupStagingAndCommitFlow(cwd, specName)
			if err != nil {
				return err
			}
		}
	}

	// Step 10: Summary
	fmt.Printf("\n%s Mockup saved to %s/\n", ui.Checkmark(), outputDir)

	if mockupJSON {
		result := mockup.MockupResult{
			Status:              "success",
			Framework:           string(framework),
			SpecName:            specName,
			MockupPath:          outputDir + "/",
			Format:              string(format),
			DesignSystemCreated: dsCreated,
			ComponentsScanned:   len(ds.Components),
			ComponentsSelected:  len(selectedComponents),
			AgentLaunched:       agentLaunched,
			Committed:           committed,
		}
		return printJSON(result)
	}

	return nil
}

func runMockupUpdate(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	dsPath := filepath.Join(cwd, "specledger", "design_system.md")
	if _, err := os.Stat(dsPath); os.IsNotExist(err) {
		return fmt.Errorf("Error: Design system not found\n\nNo design system at specledger/design_system.md\nGenerate one first with: sl mockup <spec-name>")
	}

	existing, err := mockup.LoadDesignSystem(dsPath)
	if err != nil {
		return fmt.Errorf("failed to load design system: %w", err)
	}

	framework := existing.Framework

	// Detect framework if missing
	if framework == "" || framework == mockup.FrameworkUnknown {
		detection, err := mockup.DetectFramework(cwd)
		if err != nil {
			return fmt.Errorf("framework detection failed: %w", err)
		}
		if !detection.IsFrontend {
			return fmt.Errorf("Error: Not a frontend project\n\nNo frontend framework detected in this repository.")
		}
		framework = detection.Framework
	}

	if !updateJSON {
		var confirm bool
		err = huh.NewForm(huh.NewGroup(
			huh.NewConfirm().
				Title("Rescan components?").
				Value(&confirm),
		)).Run()
		if err != nil {
			return fmt.Errorf("rescan confirmation: %w", err)
		}
		if !confirm {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	fmt.Println("Updating design system index...")
	scanResult, err := mockup.ScanComponents(cwd, framework)
	if err != nil {
		return fmt.Errorf("component scan failed: %w", err)
	}

	added, removed := mockup.MergeDesignSystem(existing, scanResult)

	existing.Framework = framework
	if err := mockup.WriteDesignSystem(dsPath, existing); err != nil {
		return fmt.Errorf("Error: Cannot write to specledger/\n\nCheck file permissions and try again.")
	}

	fmt.Printf("%s Scanned %d components\n", ui.Checkmark(), len(scanResult.Components))
	if added > 0 {
		fmt.Printf("%s Added %d new components\n", ui.Checkmark(), added)
	}
	if removed > 0 {
		fmt.Printf("%s Removed %d stale components\n", ui.Checkmark(), removed)
	}
	fmt.Printf("%s Updated specledger/design_system.md\n", ui.Checkmark())

	if updateJSON {
		result := mockup.UpdateResult{
			Status:            "success",
			ComponentsTotal:   len(existing.Components),
			ComponentsAdded:   added,
			ComponentsRemoved: removed,
		}
		return printJSON(result)
	}

	return nil
}

// resolveSpec determines the target spec name from args, branch, or interactive picker.
func resolveSpec(cwd string, args []string) (string, error) {
	if len(args) > 0 {
		fmt.Printf("Resolving spec...\n%s Using spec: %s\n", ui.Checkmark(), args[0])
		return args[0], nil
	}

	// Try auto-detect from branch
	detector := issues.NewContextDetector(cwd)
	specName, err := detector.DetectSpecContext()
	if err == nil {
		fmt.Printf("Resolving spec...\n%s Detected spec from branch: %s\n", ui.Checkmark(), specName)
		return specName, nil
	}

	// Fallback: list spec directories for interactive picker
	specledgerDir := filepath.Join(cwd, "specledger")
	entries, err := os.ReadDir(specledgerDir)
	if err != nil {
		return "", fmt.Errorf("Error: Cannot detect spec\n\nNot on a feature branch and no spec-name provided.\nProvide a spec name: sl mockup <spec-name>")
	}

	var specs []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		specFile := filepath.Join(specledgerDir, e.Name(), "spec.md")
		if _, err := os.Stat(specFile); err == nil {
			specs = append(specs, e.Name())
		}
	}

	if len(specs) == 0 {
		return "", fmt.Errorf("Error: Cannot detect spec\n\nNot on a feature branch and no spec-name provided.\nProvide a spec name: sl mockup <spec-name>")
	}

	options := make([]huh.Option[string], 0, len(specs))
	for _, s := range specs {
		options = append(options, huh.NewOption(s, s))
	}

	var selected string
	err = huh.NewForm(huh.NewGroup(
		huh.NewSelect[string]().
			Title("Select a spec").
			Options(options...).
			Value(&selected),
	)).Run()
	if err != nil {
		return "", fmt.Errorf("spec selection: %w", err)
	}

	fmt.Printf("%s Selected spec: %s\n", ui.Checkmark(), selected)
	return selected, nil
}

// mockupEditAndConfirm opens the prompt in the user's editor then shows an action menu.
// Returns the final prompt string (empty signals cancellation or write-to-file).
func mockupEditAndConfirm(initial string) (string, error) {
	const (
		actionLaunch    = "launch"
		actionReEdit    = "reedit"
		actionWriteFile = "writefile"
		actionCancel    = "cancel"
	)

	current := initial
	firstEdit := true
	for {
		editor := prompt.DetectEditor()

		if firstEdit {
			firstEdit = false

			const (
				choiceEdit    = "edit"
				choiceProceed = "proceed"
			)
			var choice string

			editLabel := "Review/edit the prompt"
			if editor != "" {
				editLabel = fmt.Sprintf("Review/edit the prompt in %s", editor)
			}

			err := huh.NewForm(huh.NewGroup(
				huh.NewSelect[string]().
					Title("Mockup prompt is ready.").
					Options(
						huh.NewOption(editLabel, choiceEdit),
						huh.NewOption("Proceed with the generated prompt", choiceProceed),
					).
					Value(&choice),
			)).Run()
			if err != nil {
				return "", fmt.Errorf("prompt choice: %w", err)
			}

			if choice == choiceProceed {
				goto actionMenu
			}
		}

		{
			edited, err := prompt.EditPrompt(current)
			if err != nil {
				if strings.Contains(err.Error(), "no editor found") {
					fmt.Println("  No editor found ($EDITOR/$VISUAL not set)")
				} else {
					fmt.Printf("Editor unavailable (%v).\n", err)
				}
				return "", mockupWritePromptToFile(current)
			}
			current = edited
		}

	actionMenu:
		var action string
		err := huh.NewForm(huh.NewGroup(
			huh.NewSelect[string]().
				Title("Prompt ready - what next?").
				Options(
					huh.NewOption("Launch AI agent with this prompt", actionLaunch),
					huh.NewOption("Re-edit the prompt", actionReEdit),
					huh.NewOption("Write prompt to a file", actionWriteFile),
					huh.NewOption("Cancel", actionCancel),
				).
				Value(&action),
		)).Run()
		if err != nil {
			return "", fmt.Errorf("action menu: %w", err)
		}

		switch action {
		case actionLaunch:
			return current, nil
		case actionReEdit:
			continue
		case actionWriteFile:
			return "", mockupWritePromptToFile(current)
		case actionCancel:
			fmt.Println("Cancelled.")
			return "", nil
		}
	}
}

// mockupWritePromptToFile asks for a filename and writes the prompt there.
func mockupWritePromptToFile(promptText string) error {
	var filename string
	err := huh.NewForm(huh.NewGroup(
		huh.NewInput().
			Title("Save prompt to file").
			Placeholder("mockup-prompt.md").
			Value(&filename),
	)).Run()
	if err != nil {
		return fmt.Errorf("filename input: %w", err)
	}
	if filename == "" {
		filename = "mockup-prompt.md"
	}
	if err := os.WriteFile(filename, []byte(promptText), 0600); err != nil {
		return fmt.Errorf("failed to write prompt: %w", err)
	}
	fmt.Printf("Prompt written to %s\n", filename)
	return nil
}

// mockupStagingAndCommitFlow handles post-agent commit/push for mockup command.
func mockupStagingAndCommitFlow(cwd, specName string) (committed bool, err error) {
	changedFiles, err := cligit.GetChangedFiles(cwd)
	if err != nil {
		return false, fmt.Errorf("failed to list changed files: %w", err)
	}

	if len(changedFiles) == 0 {
		return true, nil
	}

	fmt.Println("\nAgent session complete. Changed files:")
	for _, f := range changedFiles {
		fmt.Printf("  M %s\n", f)
	}
	fmt.Println()

	var doCommit bool
	err = huh.NewForm(huh.NewGroup(
		huh.NewConfirm().
			Title("Commit and push?").
			Value(&doCommit),
	)).Run()
	if err != nil {
		return false, fmt.Errorf("commit confirmation: %w", err)
	}

	if !doCommit {
		return false, nil
	}

	options := make([]huh.Option[string], 0, len(changedFiles))
	for _, f := range changedFiles {
		options = append(options, huh.NewOption(f, f).Selected(true))
	}

	selected := make([]string, 0, len(changedFiles))
	err = huh.NewForm(huh.NewGroup(
		huh.NewMultiSelect[string]().
			Title("Select files to stage").
			Options(options...).
			Value(&selected),
	)).Run()
	if err != nil {
		return false, fmt.Errorf("file selection: %w", err)
	}

	if len(selected) == 0 {
		return false, nil
	}

	var commitMsg string
	defaultMsg := fmt.Sprintf("feat: generate mockup for %s", specName)
	err = huh.NewForm(huh.NewGroup(
		huh.NewInput().
			Title("Commit message").
			Placeholder(defaultMsg).
			Value(&commitMsg),
	)).Run()
	if err != nil {
		return false, fmt.Errorf("commit message input: %w", err)
	}

	if commitMsg == "" {
		commitMsg = defaultMsg
	}

	if err := cligit.AddFiles(cwd, selected); err != nil {
		return false, fmt.Errorf("failed to stage files: %w", err)
	}

	hash, err := cligit.CommitChanges(cwd, commitMsg)
	if err != nil {
		return false, fmt.Errorf("failed to commit: %w", err)
	}
	fmt.Printf("%s Committed %s: %s\n", ui.Checkmark(), hash, commitMsg)

	fmt.Print("Pushing to remote... ")
	if err := cligit.PushToRemote(cwd); err != nil {
		return false, fmt.Errorf("push failed: %w", err)
	}
	fmt.Printf("%s Pushed to remote.\n", ui.Checkmark())

	return true, nil
}

// printJSON marshals the value to JSON and writes to stdout.
func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
