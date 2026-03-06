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
	Use:   "mockup [prompt...]",
	Short: "Generate UI mockups from feature specifications",
	Long: `Generate UI mockups from feature specifications.

Auto-detects spec from current branch. Pass instructions for the AI agent.

Examples:
  sl mockup                                    # Interactive flow
  sl mockup help me gen mockup ui for spec     # With custom instructions
  sl mockup focus on the login form            # With custom instructions
  sl mockup -y                                 # Auto-confirm all prompts
  sl mockup --format jsx                       # Generate JSX mockup`,
	Args:         cobra.ArbitraryArgs,
	RunE:         runMockup,
	SilenceUsage: true,
}

var mockupUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Refresh the design system by re-extracting global CSS and design tokens",
	Long: `Refresh the design system by re-extracting global CSS and design tokens.

Re-extracts CSS variables, theme colors, and styling patterns.

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
	mockupYes     bool
	mockupPrompt  string
	updateJSON    bool
)

func init() {
	VarMockupCmd.Flags().StringVar(&mockupFormat, "format", "html", "Output format: html or jsx")
	VarMockupCmd.Flags().BoolVarP(&mockupForce, "force", "f", false, "Bypass frontend detection check")
	VarMockupCmd.Flags().BoolVar(&mockupDryRun, "dry-run", false, "Write prompt to file instead of launching agent")
	VarMockupCmd.Flags().BoolVar(&mockupSummary, "summary", false, "Compact output for agent/CI integration")
	VarMockupCmd.Flags().BoolVar(&mockupJSON, "json", false, "Non-interactive path, output result as JSON")
	VarMockupCmd.Flags().BoolVarP(&mockupYes, "yes", "y", false, "Auto-confirm all prompts and launch agent directly")
	VarMockupCmd.Flags().StringVarP(&mockupPrompt, "prompt", "p", "", "Additional instructions for the AI agent")

	mockupUpdateCmd.Flags().BoolVar(&updateJSON, "json", false, "Output result as JSON")

	VarMockupCmd.AddCommand(mockupUpdateCmd)
}

func runMockup(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// If args provided, use as user prompt and skip confirmations
	hasUserInput := len(args) > 0
	if hasUserInput {
		mockupPrompt = strings.Join(args, " ")
	}
	// Skip confirmations when: --yes flag, --json flag, or user provided input
	skipConfirm := mockupYes || mockupJSON || hasUserInput

	// Validate format
	format := mockup.MockupFormat(mockupFormat)
	if !format.IsValid() {
		return fmt.Errorf("Error: Invalid format '%s'\n\nSupported formats: html, jsx", mockupFormat)
	}

	// Step 1: Resolve spec (always auto-detect from branch)
	specName, err := resolveSpec(cwd, nil)
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
	}

	framework := detection.Framework
	if mockupForce && !detection.IsFrontend {
		framework = mockup.FrameworkUnknown
	}

	// Step 3: Design system check/generate (extracts global CSS/design tokens only)
	dsPath := filepath.Join(cwd, ".specledger", "memory", "design-system.md")
	dsCreated := false
	skipGenerate := false

	if _, err := os.Stat(dsPath); os.IsNotExist(err) {
		fmt.Println("Design system not found.")
		if !skipConfirm {
			generate := true
			err = huh.NewForm(huh.NewGroup(
				huh.NewConfirm().
					Title("Generate design system now?").
					Description("Extracts global CSS, design tokens, and styling patterns").
					Value(&generate),
			)).Run()
			if err != nil {
				return fmt.Errorf("design system prompt: %w", err)
			}
			if !generate {
				fmt.Println("Skipping design system generation.")
				skipGenerate = true
			}
		}
		if !skipGenerate {
			// Extract global CSS/design tokens and app structure
			styleInfo := mockup.ScanStyles(cwd)
			ds := &mockup.DesignSystem{
				Version:      1,
				Framework:    framework,
				Style:        styleInfo,
				AppStructure: mockup.ScanAppStructure(cwd, framework),
			}
			if err := mockup.WriteDesignSystem(dsPath, ds); err != nil {
				return fmt.Errorf("Error: Cannot write to .specledger/memory/\n\nCheck file permissions and try again.")
			}
			fmt.Printf("%s Extracted design tokens\n", ui.Checkmark())
			fmt.Printf("%s Created .specledger/memory/design-system.md\n", ui.Checkmark())
			dsCreated = true
		}
	} else {
		_, loadErr := mockup.LoadDesignSystem(dsPath)
		if loadErr != nil {
			fmt.Printf("%s Design system is malformed, regenerating...\n", ui.WarningIcon())
			styleInfo := mockup.ScanStyles(cwd)
			ds := &mockup.DesignSystem{
				Version:      1,
				Framework:    framework,
				Style:        styleInfo,
				AppStructure: mockup.ScanAppStructure(cwd, framework),
			}
			if writeErr := mockup.WriteDesignSystem(dsPath, ds); writeErr != nil {
				return fmt.Errorf("failed to write design system: %w", writeErr)
			}
			dsCreated = true
		} else {
			fmt.Printf("%s Loaded design system\n", ui.Checkmark())
		}
	}

	// Step 4: Generate prompt
	fmt.Println("\nGenerating prompt...")
	specContent, err := mockup.ParseSpec(specFile)
	if err != nil {
		return err
	}

	if len(specContent.UserStories) == 0 {
		return fmt.Errorf("Error: Spec has no user scenarios\n\nThe spec.md file has no user scenarios to generate mockups from.\nAdd user scenarios with: sl clarify %s", specName)
	}

	// Output path for the mockup file
	outputPath := filepath.Join("specledger", specName, fmt.Sprintf("mockup.%s", format))
	fullOutputPath := filepath.Join(cwd, outputPath)

	// Check for existing mockup
	if _, err := os.Stat(fullOutputPath); err == nil && !skipConfirm {
		var overwrite bool
		err = huh.NewForm(huh.NewGroup(
			huh.NewConfirm().
				Title(fmt.Sprintf("Mockup already exists at %s\nOverwrite?", outputPath)).
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

	promptCtx := mockup.BuildMockupPromptContext(specName, specFile, specContent.Title, framework, format, outputPath, mockupPrompt)
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
				AgentLaunched:       false,
				Committed:           false,
			}
			return printJSON(result)
		}
		return nil
	}

	// Edit & confirm prompt (skip if --yes flag, --json, or user provided input)
	var finalPrompt string
	if skipConfirm {
		finalPrompt = promptText
	} else {
		finalPrompt, err = mockupEditAndConfirm(promptText)
		if err != nil {
			return err
		}
		if finalPrompt == "" {
			return nil // user cancelled or wrote to file
		}
	}

	// Step 5: Launch agent
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
		opts := launcher.LaunchOptions{
			SkipPermissions: true,
			Model:           "claude-sonnet-4-6",
			MaxOutputTokens: 64000,
		}
		if err := al.LaunchWithPromptAndOptions(finalPrompt, opts); err != nil {
			return fmt.Errorf("agent exited with error: %w", err)
		}
		agentLaunched = true
	}

	// Step 6: Post-agent commit/push flow (always handled by CLI, not AI agent)
	committed := false
	if agentLaunched {
		changesAfterAgent, err := cligit.HasUncommittedChanges(cwd)
		if err != nil {
			return fmt.Errorf("failed to check git status after agent: %w", err)
		}

		if changesAfterAgent {
			// With -y flag: auto-confirm commit, otherwise ask user
			committed, err = mockupStagingAndCommitFlow(cwd, specName, mockupYes)
			if err != nil {
				return err
			}
		}
	}

	// Summary
	fmt.Printf("\n%s Mockup saved to %s\n", ui.Checkmark(), outputPath)

	if mockupJSON {
		result := mockup.MockupResult{
			Status:              "success",
			Framework:           string(framework),
			SpecName:            specName,
			MockupPath:          outputPath,
			Format:              string(format),
			DesignSystemCreated: dsCreated,
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

	dsPath := filepath.Join(cwd, ".specledger", "memory", "design-system.md")
	if _, err := os.Stat(dsPath); os.IsNotExist(err) {
		return fmt.Errorf("Error: Design system not found\n\nNo design system at .specledger/memory/design-system.md\nGenerate one first with: sl mockup <spec-name>")
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
				Title("Re-extract design tokens?").
				Description("Updates CSS variables, theme colors, and styling patterns").
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

	fmt.Println("Re-extracting design tokens...")
	styleInfo := mockup.ScanStyles(cwd)

	existing.Framework = framework
	existing.Style = styleInfo
	existing.AppStructure = mockup.ScanAppStructure(cwd, framework)
	if err := mockup.WriteDesignSystem(dsPath, existing); err != nil {
		return fmt.Errorf("Error: Cannot write to .specledger/memory/\n\nCheck file permissions and try again.")
	}

	fmt.Printf("%s Extracted design tokens\n", ui.Checkmark())
	fmt.Printf("%s Updated .specledger/memory/design-system.md\n", ui.Checkmark())

	if updateJSON {
		result := mockup.UpdateResult{
			Status: "success",
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
func mockupStagingAndCommitFlow(cwd, specName string, autoConfirm bool) (committed bool, err error) {
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
	if autoConfirm {
		doCommit = true
	} else {
		err = huh.NewForm(huh.NewGroup(
			huh.NewConfirm().
				Title("Commit and push?").
				Value(&doCommit),
		)).Run()
		if err != nil {
			return false, fmt.Errorf("commit confirmation: %w", err)
		}
	}

	if !doCommit {
		return false, nil
	}

	var selected []string
	if autoConfirm {
		selected = changedFiles
	} else {
		options := make([]huh.Option[string], 0, len(changedFiles))
		for _, f := range changedFiles {
			options = append(options, huh.NewOption(f, f).Selected(true))
		}

		selected = make([]string, 0, len(changedFiles))
		err = huh.NewForm(huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select files to stage").
				Options(options...).
				Value(&selected),
		)).Run()
		if err != nil {
			return false, fmt.Errorf("file selection: %w", err)
		}
	}

	if len(selected) == 0 {
		return false, nil
	}

	defaultMsg := fmt.Sprintf("feat: generate mockup for %s", specName)
	var commitMsg string
	if autoConfirm {
		commitMsg = defaultMsg
	} else {
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
