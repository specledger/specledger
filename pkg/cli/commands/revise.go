package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/specledger/specledger/pkg/cli/auth"
	"github.com/specledger/specledger/pkg/cli/config"
	cligit "github.com/specledger/specledger/pkg/cli/git"
	"github.com/specledger/specledger/pkg/cli/launcher"
	"github.com/specledger/specledger/pkg/cli/revise"
	"github.com/spf13/cobra"
)

// VarReviseCmd is the sl revise command.
var VarReviseCmd = &cobra.Command{
	Use:   "revise [branch-name]",
	Short: "Fetch and address review comments for a spec",
	Long: `Fetch unresolved review comments from Supabase and guide you through
addressing them with an AI coding agent.

Flow:
  1. Detect or select the target branch
  2. Fetch unresolved comments from Supabase
  3. Select artifacts to work on (multi-select)
  4. Process each comment (provide guidance or skip)
  5. Generate a combined revision prompt
  6. Open the prompt in your editor for refinement
  7. Launch the configured AI coding agent
  8. Offer to commit/push changes and resolve comments

Examples:
  sl revise                          # Interactive: detect branch, fetch comments
  sl revise 136-revise-comments      # Use the specified branch directly
  sl revise --summary                # Print compact comment listing and exit
  sl revise --auto fixture.json      # Non-interactive: fixture-driven prompt generation
  sl revise --dry-run                # Interactive flow but write prompt to file instead of launching agent`,
	Args:         cobra.MaximumNArgs(1),
	RunE:         runRevise,
	SilenceUsage: true,
}

var (
	reviseAutoFixture string
	reviseDryRun      bool
	reviseSummary     bool
)

func init() {
	VarReviseCmd.Flags().StringVar(&reviseAutoFixture, "auto", "", "Non-interactive mode: path to fixture JSON file")
	VarReviseCmd.Flags().BoolVar(&reviseDryRun, "dry-run", false, "Write prompt to file instead of launching agent")
	VarReviseCmd.Flags().BoolVar(&reviseSummary, "summary", false, "Print compact comment listing and exit (for agent integration)")
}

func runRevise(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// --summary: silent on auth failure, compact output for agent callers
	if reviseSummary {
		return runSummary(cwd, args)
	}

	// --auto: non-interactive fixture-driven flow (US8)
	if reviseAutoFixture != "" {
		return runAuto(cwd, args, reviseAutoFixture)
	}

	// Step 1: Auth check
	accessToken, err := auth.GetValidAccessToken()
	if err != nil {
		return fmt.Errorf("authentication required: %w\n\nRun 'sl auth login' to authenticate.", err)
	}

	client := revise.NewReviseClient(accessToken)

	// Step 2: Branch selection → resolve specKey
	specKey, needsCheckout, err := resolveBranch(cwd, args, client)
	if err != nil {
		return err
	}

	// Step 3: Checkout target branch if different from current (US7)
	stashUsed, err := checkoutIfNeeded(cwd, needsCheckout)
	if err != nil {
		return err
	}

	// Step 4: Fetch comments via PostgREST query chain
	comments, changeID, err := fetchComments(cwd, specKey, client)
	if err != nil {
		return err
	}

	if len(comments) == 0 {
		fmt.Printf("No unresolved comments found for %s.\n", specKey)
		return nil
	}

	// Fetch thread replies for all comments in this change
	replies, err := client.FetchReplies(changeID)
	if err != nil {
		// Non-fatal: proceed without threads
		fmt.Fprintf(os.Stderr, "warning: failed to fetch thread replies: %v\n", err)
		replies = nil
	}
	replyMap := revise.BuildReplyMap(replies)

	replyCount := len(replies)
	if replyCount > 0 {
		fmt.Printf("Fetched %d unresolved comment(s) with %d thread reply(ies) for %s.\n", len(comments), replyCount, specKey)
	} else {
		fmt.Printf("Fetched %d unresolved comment(s) for %s.\n", len(comments), specKey)
	}

	// Step 5: Artifact multi-select (US2)
	selectedComments, err := selectArtifacts(comments)
	if err != nil {
		return err
	}

	if len(selectedComments) == 0 {
		fmt.Println("No artifacts selected. Nothing to process.")
		return nil
	}

	// Step 6: Comment processing loop (US3)
	processed, err := processComments(selectedComments, replyMap)
	if err != nil {
		return err
	}

	if len(processed) == 0 {
		fmt.Println("No comments selected for processing. Nothing to do.")
		return nil
	}

	fmt.Printf("Processing %d comment(s).\n", len(processed))

	// Step 7: Generate revision prompt (US4)
	ctx := revise.BuildRevisionContext(specKey, processed, replies)
	prompt, err := revise.RenderPrompt(ctx)
	if err != nil {
		return fmt.Errorf("failed to render prompt: %w", err)
	}

	tokens := revise.EstimateTokens(prompt)
	revise.PrintTokenWarnings(tokens)

	// Step 8: Open editor for prompt refinement; confirm/re-edit/write-to-file/cancel.
	finalPrompt, err := editAndConfirmPrompt(prompt, reviseDryRun)
	if err != nil {
		return err
	}
	if finalPrompt == "" {
		// User wrote to file (dry-run or manual), or cancelled.
		return nil
	}

	// Step 9: Launch agent (US5) — inlined, called exactly once.
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
	if !al.IsAvailable() {
		fmt.Printf("No AI agent found. Install with: %s\n", al.InstallInstructions())
		return writePromptToFile(finalPrompt)
	}

	// Inject config environment variables (base-url, auth-token, model overrides, etc.)
	al.SetEnv(config.ResolveAgentEnv())

	fmt.Printf("Launching %s...\n", al.Name)
	if err := al.LaunchWithPrompt(finalPrompt); err != nil {
		return fmt.Errorf("agent exited with error: %w", err)
	}

	changesAfterAgent, err := cligit.HasUncommittedChanges(cwd)
	if err != nil {
		return fmt.Errorf("failed to check git status after agent: %w", err)
	}

	// US6: File staging, commit, and push (sl-ssr)
	if changesAfterAgent {
		committed, err := stagingAndCommitFlow(cwd)
		if err != nil {
			return err
		}

		if !committed {
			// FR-019: warn that resolving without pushing may cause inconsistencies,
			// then give the user a chance to bail out (default: No = defer entirely).
			fmt.Println("\n⚠ Changes not committed. Resolving comments without pushing may lead")
			fmt.Println("  to inconsistencies on the remote.")
			fmt.Println()

			var proceedAnyway bool
			err = huh.NewForm(huh.NewGroup(
				huh.NewConfirm().
					Title("Proceed to resolve comments anyway?").
					Value(&proceedAnyway), // default false → No
			)).Run()
			if err != nil {
				return fmt.Errorf("proceed confirmation: %w", err)
			}

			if !proceedAnyway {
				fmt.Println("Unresolved comments remain. Re-run `sl revise` after pushing to resolve them.")
				if stashUsed {
					fmt.Println("\nYou have stashed changes. Run 'git stash pop' to restore them.")
				}
				return nil
			}
		}
	}

	// US6: Comment resolution multi-select (sl-x1o)
	if err := commentResolutionFlow(processed, replyMap, client, stashUsed); err != nil {
		return err
	}

	return nil
}

// stagingAndCommitFlow prints changed files, asks the user whether to commit,
// and if confirmed: shows a file multi-select, prompts for a commit message,
// then stages/commits/pushes. Returns committed=true when a commit was made,
// false when the user chose to skip (triggering the FR-019 second confirm in the caller).
func stagingAndCommitFlow(cwd string) (committed bool, err error) {
	changedFiles, err := cligit.GetChangedFiles(cwd)
	if err != nil {
		return false, fmt.Errorf("failed to list changed files: %w", err)
	}

	if len(changedFiles) == 0 {
		return true, nil // nothing uncommitted — treat as committed
	}

	// Print changed-files summary (mirrors quickstart §6 output)
	fmt.Println("\nAgent session complete. Changed files:")
	for _, f := range changedFiles {
		fmt.Printf("  M %s\n", f)
	}
	fmt.Println()

	// First confirm: commit and push?
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
		return false, nil // caller handles FR-019 second confirm
	}

	// File multi-select — all pre-selected
	options := make([]huh.Option[string], 0, len(changedFiles))
	for _, f := range changedFiles {
		options = append(options, huh.NewOption(f, f).Selected(true))
	}

	selected := make([]string, 0, len(changedFiles))
	err = huh.NewForm(huh.NewGroup(
		huh.NewMultiSelect[string]().
			Title("Select files to stage").
			Description("All changed files are pre-selected. Deselect any you want to leave uncommitted.").
			Options(options...).
			Value(&selected),
	)).Run()
	if err != nil {
		return false, fmt.Errorf("file selection: %w", err)
	}

	if len(selected) == 0 {
		return false, nil // nothing staged — treat same as skipping commit
	}

	// Commit message
	var commitMsg string
	err = huh.NewForm(huh.NewGroup(
		huh.NewInput().
			Title("Commit message").
			Placeholder("feat: address review feedback").
			Value(&commitMsg),
	)).Run()
	if err != nil {
		return false, fmt.Errorf("commit message input: %w", err)
	}

	if commitMsg == "" {
		commitMsg = "feat: address review feedback"
	}

	if err := cligit.AddFiles(cwd, selected); err != nil {
		return false, fmt.Errorf("failed to stage files: %w", err)
	}

	hash, err := cligit.CommitChanges(cwd, commitMsg)
	if err != nil {
		return false, fmt.Errorf("failed to commit: %w", err)
	}
	fmt.Printf("✓ Committed %s: %s\n", hash, commitMsg)

	fmt.Print("Pushing to remote... ")
	if err := cligit.PushToRemote(cwd); err != nil {
		return false, fmt.Errorf("push failed: %w", err)
	}
	fmt.Println("✓ Pushed to remote.")

	return true, nil
}

// commentResolutionFlow shows a multi-select of processed comments and marks the
// selected ones as resolved via the API (FR-017, FR-018, FR-021).
// When a parent comment is resolved, its thread replies are also resolved (cascade).
// Prints the stash pop reminder at session end if stashUsed.
func commentResolutionFlow(processed []revise.ProcessedComment, replyMap map[string][]revise.ReviewComment, client *revise.ReviseClient, stashUsed bool) error {
	if len(processed) == 0 {
		return nil
	}

	options := make([]huh.Option[string], 0, len(processed))
	for _, p := range processed {
		label := p.Comment.FilePath
		if p.Comment.SelectedText != "" {
			excerpt := p.Comment.SelectedText
			if len(excerpt) > 40 {
				excerpt = excerpt[:37] + "..."
			}
			label = fmt.Sprintf("%s — %q", p.Comment.FilePath, excerpt)
		}
		if reps := replyMap[p.Comment.ID]; len(reps) > 0 {
			noun := "reply"
			if len(reps) != 1 {
				noun = "replies"
			}
			label = fmt.Sprintf("%s [%d %s]", label, len(reps), noun)
		}
		options = append(options, huh.NewOption(label, p.Comment.ID).Selected(true))
	}

	selected := make([]string, 0, len(processed))
	err := huh.NewForm(huh.NewGroup(
		huh.NewMultiSelect[string]().
			Title("Mark comments as resolved?").
			Description("Select the comments that were successfully addressed by the agent. Thread replies will also be resolved.").
			Options(options...).
			Value(&selected),
	)).Run()
	if err != nil {
		return fmt.Errorf("resolution selection: %w", err)
	}

	if len(selected) == 0 {
		// FR-021: all deferred
		fmt.Println("Unresolved comments remain. Re-run sl revise after pushing to resolve them.")
		if stashUsed {
			fmt.Println("\nYou have stashed changes. Run 'git stash pop' to restore them.")
		}
		return nil
	}

	resolved := 0
	resolvedReplies := 0
	for _, id := range selected {
		replyIDs := make([]string, 0)
		if reps, ok := replyMap[id]; ok {
			for _, r := range reps {
				replyIDs = append(replyIDs, r.ID)
			}
		}

		if len(replyIDs) > 0 {
			if err := client.ResolveCommentWithReplies(id, replyIDs); err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to resolve comment %s: %v\n", id, err)
				continue
			}
			resolvedReplies += len(replyIDs)
		} else {
			if err := client.ResolveComment(id); err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to resolve comment %s: %v\n", id, err)
				continue
			}
		}
		resolved++
	}

	if resolvedReplies > 0 {
		fmt.Printf("Resolved %d of %d comment(s) + %d thread reply(ies).\n", resolved, len(processed), resolvedReplies)
	} else {
		fmt.Printf("Resolved %d of %d comment(s).\n", resolved, len(processed))
	}
	if resolved < len(processed) {
		fmt.Println("Unresolved comments remain. Re-run sl revise after pushing to resolve them.")
	}

	if stashUsed {
		fmt.Println("\nYou have stashed changes. Run 'git stash pop' to restore them.")
	}

	return nil
}

// resolveBranch determines the target spec key.
// Returns (specKey, targetBranch, error). targetBranch is empty if already on the target branch.
func resolveBranch(cwd string, args []string, client *revise.ReviseClient) (specKey, targetBranch string, err error) {
	// Explicit branch arg → use directly, no picker needed
	if len(args) > 0 {
		return args[0], "", nil
	}

	currentBranch, err := cligit.GetCurrentBranch(cwd)
	if err != nil {
		return "", "", fmt.Errorf("failed to detect current branch: %w", err)
	}

	// On a feature branch: confirm with user
	if cligit.IsFeatureBranch(currentBranch) {
		confirmed := true
		err = huh.NewForm(huh.NewGroup(
			huh.NewConfirm().
				Title(fmt.Sprintf("Use branch %q for revision?", currentBranch)).
				Description("Press Enter to confirm or N to choose a different branch.").
				Value(&confirmed),
		)).Run()
		if err != nil {
			return "", "", fmt.Errorf("branch confirmation: %w", err)
		}

		if confirmed {
			return currentBranch, "", nil
		}
	}

	// Not a feature branch or user wants to pick: show branch list from API
	return pickBranchFromAPI(cwd, currentBranch, client)
}

// pickBranchFromAPI fetches specs with unresolved comments and shows a branch picker.
func pickBranchFromAPI(cwd, currentBranch string, client *revise.ReviseClient) (specKey, targetBranch string, err error) {
	repoOwner, repoName, err := cligit.GetRepoOwnerName(cwd)
	if err != nil {
		return "", "", fmt.Errorf("failed to detect repo: %w", err)
	}

	project, err := client.GetProject(repoOwner, repoName)
	if err != nil {
		return "", "", networkHint(fmt.Errorf("failed to fetch project: %w", err))
	}

	specs, err := client.ListSpecsWithComments(project.ID)
	if err != nil {
		return "", "", networkHint(fmt.Errorf("failed to fetch specs: %w", err))
	}

	if len(specs) == 0 {
		return "", "", fmt.Errorf("no specs with unresolved comments found for %s/%s", repoOwner, repoName)
	}

	options := make([]huh.Option[string], 0, len(specs))
	for _, s := range specs {
		label := fmt.Sprintf("%s (%d comment(s))", s.SpecKey, s.CommentCount)
		options = append(options, huh.NewOption(label, s.SpecKey))
	}

	var selected string
	err = huh.NewForm(huh.NewGroup(
		huh.NewSelect[string]().
			Title("Select a branch to revise").
			Options(options...).
			Value(&selected),
	)).Run()
	if err != nil {
		return "", "", fmt.Errorf("branch selection: %w", err)
	}

	// Determine if checkout is needed
	if selected != currentBranch {
		return selected, selected, nil
	}

	return selected, "", nil
}

// checkoutIfNeeded handles stash detection, confirmation, and branch checkout when the
// target branch differs from the current branch (US7).
func checkoutIfNeeded(cwd, targetBranch string) (stashUsed bool, err error) {
	if targetBranch == "" {
		return false, nil
	}

	dirty, err := cligit.HasUncommittedChanges(cwd)
	if err != nil {
		return false, fmt.Errorf("failed to check git status: %w", err)
	}

	if dirty {
		const (
			actionStash    = "stash"
			actionAbort    = "abort"
			actionContinue = "continue"
		)

		var action string
		err = huh.NewForm(huh.NewGroup(
			huh.NewSelect[string]().
				Title("You have uncommitted changes").
				Description(fmt.Sprintf("Checking out %q requires a clean working tree.", targetBranch)).
				Options(
					huh.NewOption("Stash changes and continue", actionStash),
					huh.NewOption("Abort checkout (stay on current branch)", actionAbort),
					huh.NewOption("Continue without checkout (use current branch)", actionContinue),
				).
				Value(&action),
		)).Run()
		if err != nil {
			return false, fmt.Errorf("stash prompt: %w", err)
		}

		switch action {
		case actionAbort:
			return false, fmt.Errorf("checkout aborted by user")
		case actionContinue:
			return false, nil
		case actionStash:
			if err := cligit.StashChanges(cwd); err != nil {
				return false, fmt.Errorf("stash failed — resolve manually before switching branches: %w", err)
			}
			stashUsed = true
		}
	}

	// Checkout: try local first, fall back to remote tracking
	exists, err := cligit.BranchExists(cwd, targetBranch)
	if err != nil {
		return stashUsed, fmt.Errorf("failed to check branch existence: %w", err)
	}

	if exists {
		if err := cligit.CheckoutBranch(cwd, targetBranch); err != nil {
			return stashUsed, fmt.Errorf("checkout failed: %w", err)
		}
	} else {
		if err := cligit.CheckoutRemoteTracking(cwd, targetBranch); err != nil {
			return stashUsed, fmt.Errorf(
				"branch %q not found on remote (it may have been deleted)\nRun `sl revise` again to select a different branch.",
				targetBranch)
		}
	}

	fmt.Printf("Checked out branch %q.\n", targetBranch)
	return stashUsed, nil
}

// fetchComments runs the 4-step PostgREST query chain and returns unresolved comments
// along with the changeID (needed for fetching thread replies).
func fetchComments(cwd, specKey string, client *revise.ReviseClient) ([]revise.ReviewComment, string, error) {
	repoOwner, repoName, err := cligit.GetRepoOwnerName(cwd)
	if err != nil {
		return nil, "", fmt.Errorf("failed to detect repo: %w", err)
	}

	project, err := client.GetProject(repoOwner, repoName)
	if err != nil {
		return nil, "", networkHint(fmt.Errorf("failed to fetch project: %w", err))
	}

	spec, err := client.GetSpec(project.ID, specKey)
	if err != nil {
		return nil, "", networkHint(fmt.Errorf("failed to fetch spec %q: %w", specKey, err))
	}

	change, err := client.GetChange(spec.ID)
	if err != nil {
		return nil, "", networkHint(fmt.Errorf("failed to fetch change for spec %q: %w", specKey, err))
	}

	comments, err := client.FetchComments(change.ID)
	if err != nil {
		return nil, "", networkHint(fmt.Errorf("failed to fetch comments: %w", err))
	}

	return comments, change.ID, nil
}

// selectArtifacts groups comments by file_path, shows a huh multi-select with counts,
// and returns only the comments for the artifacts the user selected.
func selectArtifacts(comments []revise.ReviewComment) ([]revise.ReviewComment, error) {
	// Group by file_path, preserving first-seen order
	order := make([]string, 0)
	byArtifact := make(map[string][]revise.ReviewComment)
	for _, c := range comments {
		if _, seen := byArtifact[c.FilePath]; !seen {
			order = append(order, c.FilePath)
		}
		byArtifact[c.FilePath] = append(byArtifact[c.FilePath], c)
	}

	// Build multi-select options, all pre-selected
	options := make([]huh.Option[string], 0, len(order))
	for _, fp := range order {
		count := len(byArtifact[fp])
		noun := "comment"
		if count != 1 {
			noun = "comments"
		}
		label := fmt.Sprintf("%s (%d %s)", fp, count, noun)
		options = append(options, huh.NewOption(label, fp).Selected(true))
	}

	selected := make([]string, 0, len(order))
	err := huh.NewForm(huh.NewGroup(
		huh.NewMultiSelect[string]().
			Title("Select artifacts to revise").
			Description("All artifacts with unresolved comments are pre-selected. Deselect any you want to skip.").
			Options(options...).
			Value(&selected),
	)).Run()
	if err != nil {
		return nil, fmt.Errorf("artifact selection: %w", err)
	}

	if len(selected) == 0 {
		return nil, nil
	}

	// Filter to selected artifacts
	selectedSet := make(map[string]struct{}, len(selected))
	for _, fp := range selected {
		selectedSet[fp] = struct{}{}
	}

	result := make([]revise.ReviewComment, 0, len(comments))
	for _, c := range comments {
		if _, ok := selectedSet[c.FilePath]; ok {
			result = append(result, c)
		}
	}

	return result, nil
}

// processComments shows each comment with lipgloss styling and lets the user
// choose to Process (with optional guidance), Skip, or Quit the loop.
// Thread replies from replyMap are displayed inline below each parent comment.
func processComments(comments []revise.ReviewComment, replyMap map[string][]revise.ReviewComment) ([]revise.ProcessedComment, error) {
	// Lipgloss styles for comment display
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	textStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	dividerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	threadStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("6"))

	const (
		actionProcess = "process"
		actionSkip    = "skip"
		actionQuit    = "quit"
	)

	processed := make([]revise.ProcessedComment, 0, len(comments))
	index := 1

	for i, c := range comments {
		// Render comment header
		lineInfo := ""
		if c.StartLine != nil && c.Line != nil {
			lineInfo = fmt.Sprintf(" (lines %d–%d)", *c.StartLine, *c.Line)
		} else if c.Line != nil {
			lineInfo = fmt.Sprintf(" (line %d)", *c.Line)
		}

		fmt.Println()
		fmt.Println(dividerStyle.Render(strings.Repeat("─", 60)))
		fmt.Println(headerStyle.Render(fmt.Sprintf("Comment %d of %d", i+1, len(comments))))
		fmt.Printf("%s %s%s\n", labelStyle.Render("File:"), textStyle.Render(c.FilePath), lineInfo)

		// Edge case 1: file no longer exists locally
		fileExists := true
		if _, statErr := os.Stat(c.FilePath); os.IsNotExist(statErr) {
			fmt.Printf("⚠ File not found locally: %s\n", c.FilePath)
			fileExists = false
		}

		if c.SelectedText != "" {
			// Edge case 2: selected text no longer present in the current file version.
			// Try literal match first; fall back to markdown-stripped comparison to avoid
			// false positives when the reviewer selected from a rendered view (backticks,
			// bold markers etc. are invisible in rendered markdown but present in the raw file).
			if fileExists {
				content, readErr := os.ReadFile(c.FilePath)
				if readErr == nil {
					fileStr := string(content)
					literalFound := strings.Contains(fileStr, c.SelectedText)
					strippedFound := strings.Contains(mdStripper.Replace(fileStr), mdStripper.Replace(c.SelectedText))
					if !literalFound && !strippedFound {
						fmt.Println("⚠ Original selected text not found in current file version")
					}
				}
			}

			excerpt := c.SelectedText
			if len(excerpt) > 120 {
				excerpt = excerpt[:117] + "..."
			}
			fmt.Printf("%s %s\n", labelStyle.Render("Excerpt:"), textStyle.Render(fmt.Sprintf("%q", excerpt)))
		}

		author := c.AuthorName
		if author == "" {
			author = c.AuthorEmail
		}
		fmt.Printf("%s %s\n", labelStyle.Render("Author:"), textStyle.Render(author))
		fmt.Printf("%s %s\n", labelStyle.Render("Feedback:"), textStyle.Render(c.Content))

		// Display thread replies inline
		if reps, ok := replyMap[c.ID]; ok && len(reps) > 0 {
			fmt.Println()
			fmt.Printf("  %s\n", threadStyle.Render("Thread:"))
			for _, r := range reps {
				replyAuthor := r.AuthorName
				if replyAuthor == "" {
					replyAuthor = r.AuthorEmail
				}
				replyContent := r.Content
				if len(replyContent) > 200 {
					replyContent = replyContent[:197] + "..."
				}
				fmt.Printf("  %s %s\n", threadStyle.Render("└─ "+replyAuthor+":"), textStyle.Render(replyContent))
			}
		}

		// Per-comment action prompt
		var action string
		err := huh.NewForm(huh.NewGroup(
			huh.NewSelect[string]().
				Title("What would you like to do with this comment?").
				Options(
					huh.NewOption("Process — add to revision batch (with optional guidance)", actionProcess),
					huh.NewOption("Skip — leave for later", actionSkip),
					huh.NewOption("Quit — stop processing (keep batch so far)", actionQuit),
				).
				Value(&action),
		)).Run()
		if err != nil {
			return nil, fmt.Errorf("comment action prompt: %w", err)
		}

		switch action {
		case actionQuit:
			return processed, nil
		case actionSkip:
			continue
		case actionProcess:
			var guidance string
			err = huh.NewForm(huh.NewGroup(
				huh.NewText().
					Title("Optional guidance for the AI agent").
					Description("Describe how you want this comment addressed (leave empty to let the agent decide).").
					Value(&guidance),
			)).Run()
			if err != nil {
				return nil, fmt.Errorf("guidance input: %w", err)
			}

			processed = append(processed, revise.ProcessedComment{
				Comment:  c,
				Guidance: strings.TrimSpace(guidance),
				Index:    index,
			})
			index++
		}
	}

	return processed, nil
}

// runAuto handles --auto mode: fixture-driven, non-interactive, prints prompt to stdout.
func runAuto(cwd string, args []string, fixturePath string) error {
	accessToken, err := auth.GetValidAccessToken()
	if err != nil {
		return fmt.Errorf("authentication required: %w\n\nRun 'sl auth login' to authenticate.", err)
	}

	fixture, err := revise.ParseFixture(fixturePath)
	if err != nil {
		return err
	}

	client := revise.NewReviseClient(accessToken)

	// Resolve branch from fixture or args
	specKey := fixture.Branch
	if len(args) > 0 {
		specKey = args[0]
	}
	if specKey == "" {
		specKey, err = cligit.GetCurrentBranch(cwd)
		if err != nil {
			return fmt.Errorf("failed to detect branch: %w", err)
		}
	}

	comments, changeID, err := fetchComments(cwd, specKey, client)
	if err != nil {
		return err
	}

	// Fetch thread replies (non-fatal if it fails)
	replies, _ := client.FetchReplies(changeID)

	matched, warnings := revise.MatchFixtureComments(fixture, comments)
	for _, w := range warnings {
		fmt.Fprintf(os.Stderr, "warning: %s\n", w)
	}

	if len(matched) == 0 {
		fmt.Fprintf(os.Stderr, "no fixture comments matched any fetched comments\n")
		os.Exit(1)
	}

	ctx := revise.BuildRevisionContext(specKey, matched, replies)
	prompt, err := revise.RenderPrompt(ctx)
	if err != nil {
		return fmt.Errorf("failed to render prompt: %w", err)
	}

	fmt.Print(prompt)
	return nil
}

// editAndConfirmPrompt opens the prompt in the user's editor then shows a
// Launch / Re-edit / Write-to-file / Cancel menu. Returns the final prompt
// string (empty string signals cancellation or dry-run write).
// When dryRun is true, skips the action menu and immediately prompts for a filename.
func editAndConfirmPrompt(initial string, dryRun bool) (string, error) {
	const (
		actionLaunch    = "launch"
		actionReEdit    = "reedit"
		actionWriteFile = "writefile"
		actionCancel    = "cancel"
	)

	current := initial
	firstEdit := true
	for {
		editor := revise.DetectEditor()

		if firstEdit {
			firstEdit = false

			// Let the user choose: review/edit the prompt or proceed as-is.
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
					Title("Revision prompt is ready.").
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
				// Skip editor, go straight to action menu.
				if dryRun {
					_, err := writePromptInteractive(current)
					return "", err
				}
				goto actionMenu
			}
		}

		{
			edited, err := revise.EditPrompt(current)
			if err != nil {
				// Editor not found: fall through to write-to-file path
				if strings.Contains(err.Error(), "no editor found") {
					fmt.Println("⚠ No editor found ($EDITOR/$VISUAL not set, vi not available)")
				} else {
					fmt.Printf("Editor unavailable (%v).\n", err)
				}
				_, err := writePromptInteractive(current)
				return "", err
			}
			current = edited
		}

		if dryRun {
			_, err := writePromptInteractive(current)
			return "", err
		}

	actionMenu:
		var action string
		err := huh.NewForm(huh.NewGroup(
			huh.NewSelect[string]().
				Title("Prompt ready — what next?").
				Options(
					huh.NewOption("Launch AI agent with this prompt", actionLaunch),
					huh.NewOption("Re-edit the prompt", actionReEdit),
					huh.NewOption("Write prompt to a file (dry-run)", actionWriteFile),
					huh.NewOption("Cancel", actionCancel),
				).
				Value(&action),
		)).Run()
		if err != nil {
			return "", fmt.Errorf("confirm prompt: %w", err)
		}

		switch action {
		case actionLaunch:
			return current, nil
		case actionReEdit:
			continue
		case actionWriteFile:
			if err := writePromptToFile(current); err != nil {
				return "", err
			}
			return "", nil // written to file, no agent launch
		case actionCancel:
			fmt.Println("Cancelled.")
			return "", nil
		}
	}
}

// writePromptInteractive asks the user for a filename and writes the prompt there.
func writePromptInteractive(prompt string) (string, error) {
	var filename string
	err := huh.NewForm(huh.NewGroup(
		huh.NewInput().
			Title("Save prompt to file").
			Description("Enter a filename (e.g., revision-prompt.md)").
			Value(&filename),
	)).Run()
	if err != nil {
		return "", fmt.Errorf("filename input: %w", err)
	}
	if err := os.WriteFile(filename, []byte(prompt), 0600); err != nil {
		return "", fmt.Errorf("failed to write prompt: %w", err)
	}
	fmt.Printf("Prompt written to %s\n", filename)
	return "", nil
}

// mdStripper removes common Markdown formatting characters so that selected_text
// (captured from a rendered view) can be matched against raw .md file content.
var mdStripper = strings.NewReplacer("`", "", "**", "", "*", "", "_", "", "~~", "")

// networkHint appends a network connectivity hint to err when the error message
// does not already contain actionable API-level detail (e.g., "API error (404)").
// Network errors (connection refused, timeout, DNS failures) benefit from the hint;
// API-level errors already contain enough context.
func networkHint(err error) error {
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), "API error (") {
		return err
	}
	return fmt.Errorf("%w\nCheck your network connection and try again.", err)
}

// writePromptToFile prompts for a filename and writes the prompt there.
func writePromptToFile(prompt string) error {
	var filename string
	err := huh.NewForm(huh.NewGroup(
		huh.NewInput().
			Title("Write prompt to file").
			Placeholder("revision-prompt.md").
			Value(&filename),
	)).Run()
	if err != nil {
		return fmt.Errorf("filename input: %w", err)
	}
	if filename == "" {
		filename = "revision-prompt.md"
	}
	if err := os.WriteFile(filename, []byte(prompt), 0600); err != nil {
		return fmt.Errorf("failed to write prompt: %w", err)
	}
	fmt.Printf("Prompt written to %s\n", filename)
	return nil
}

// runSummary implements --summary: compact comment listing for agent callers.
// Auth failures exit silently with code 1.
func runSummary(cwd string, args []string) error {
	accessToken, err := auth.GetValidAccessToken()
	if err != nil {
		// Silent exit on auth failure (FR-025)
		os.Exit(1)
		return nil
	}

	client := revise.NewReviseClient(accessToken)

	var specKey string
	if len(args) > 0 {
		specKey = args[0]
	} else {
		currentBranch, err := cligit.GetCurrentBranch(cwd)
		if err != nil {
			os.Exit(1)
			return nil
		}
		specKey = currentBranch
	}

	comments, changeID, err := fetchComments(cwd, specKey, client)
	if err != nil {
		// Silent exit on any fetch error
		os.Exit(1)
		return nil
	}

	if len(comments) == 0 {
		fmt.Printf("0 unresolved comments for %s\n", specKey)
		return nil
	}

	// Fetch thread replies (non-fatal)
	replies, _ := client.FetchReplies(changeID)
	replyMap := revise.BuildReplyMap(replies)

	// Compact format: file_path:line  'selected_text'  (author)  [N replies]
	artifacts := make(map[string]struct{})
	for _, c := range comments {
		artifacts[c.FilePath] = struct{}{}

		lineStr := ""
		if c.StartLine != nil && c.Line != nil {
			lineStr = fmt.Sprintf("%d-%d", *c.StartLine, *c.Line)
		} else if c.Line != nil {
			lineStr = fmt.Sprintf("%d", *c.Line)
		} else {
			lineStr = "-"
		}

		selectedText := c.SelectedText
		if len(selectedText) > 50 {
			selectedText = selectedText[:47] + "..."
		}
		selectedText = strings.ReplaceAll(selectedText, "\n", " ")

		author := c.AuthorName
		if author == "" {
			author = c.AuthorEmail
		}

		replyInfo := ""
		if reps := replyMap[c.ID]; len(reps) > 0 {
			noun := "reply"
			if len(reps) != 1 {
				noun = "replies"
			}
			replyInfo = fmt.Sprintf("  [%d %s]", len(reps), noun)
		}

		fmt.Printf("%s:%s  %q  (%s)%s\n", c.FilePath, lineStr, selectedText, author, replyInfo)
	}

	fmt.Printf("\n%d unresolved comment(s) across %d artifact(s)\n", len(comments), len(artifacts))
	return nil
}
