package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/specledger/specledger/pkg/cli/auth"
	"github.com/specledger/specledger/pkg/cli/metadata"
	"github.com/specledger/specledger/pkg/cli/session"
	"github.com/spf13/cobra"
)

// VarSessionCmd represents the session command group
var VarSessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Manage AI session captures",
	Long: `Manage AI session captures for checkpoints and tasks.

Sessions are automatically captured when you commit changes while working
with Claude Code. They provide a record of the AI conversation that led
to each commit.

Commands:
  list     List sessions for a branch
  get      Retrieve session content by ID, commit hash, or task ID
  sync     Upload queued sessions (for offline captures)
  capture  (Internal) Called by Claude Code hooks

Examples:
  sl session list                    # List sessions for current branch
  sl session list --feature main     # List sessions for main branch
  sl session get abc123              # Get session by partial commit hash
  sl session get SL-42               # Get session by task ID
  sl session sync                    # Upload any queued sessions
  sl session sync --status           # Check queue status without uploading`,
}

// VarSessionCaptureCmd represents the capture command (called by hooks)
var VarSessionCaptureCmd = &cobra.Command{
	Use:   "capture",
	Short: "Capture session from hook input",
	Long: `Capture an AI session from Claude Code hook input.

This command is designed to be called by Claude Code hooks, not manually.
It reads hook JSON from stdin, detects git commits, and captures the
conversation delta since the last commit.

Test mode (for manual testing):
  sl session capture --test-mode

Exit codes:
  0 - Success (session captured/queued) or no-op (not a commit)
  1 - Fatal error (logged to stderr)`,
	RunE:         runSessionCapture,
	SilenceUsage: true,
}

// VarSessionListCmd represents the list command
var VarSessionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List sessions for a feature",
	Long: `List sessions for a feature branch.

By default, lists sessions for the current git branch.
Use --feature to specify a different branch.

Examples:
  sl session list                        # Sessions for current branch
  sl session list --feature main         # Sessions for main branch
  sl session list --commit abc123        # Filter by partial commit hash
  sl session list --task SL-42           # Filter by task ID
  sl session list --limit 10             # Show only last 10 sessions
  sl session list --json                 # Output as JSON (for scripts/AI)`,
	RunE: runSessionList,
}

// VarSessionGetCmd represents the get command
var VarSessionGetCmd = &cobra.Command{
	Use:   "get <session-id|commit-hash|task-id>",
	Short: "Retrieve a session's content",
	Long: `Retrieve and display a session's conversation content.

You can look up a session by:
  - Session ID (UUID)
  - Commit hash (full or partial, e.g., "abc1234" or full 40-char hash)
  - Task ID (e.g., SL-42)

Examples:
  sl session get abc1234               # Get by partial commit hash
  sl session get SL-42                 # Get by task ID
  sl session get 550e8400-e29b...      # Get by full UUID
  sl session get abc1234 --json        # Output as JSON (for AI processing)
  sl session get abc1234 --raw         # Output raw gzip stream`,
	Args: cobra.ExactArgs(1),
	RunE: runSessionGet,
}

// VarSessionSyncCmd represents the sync command
var VarSessionSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Upload queued sessions",
	Long: `Upload sessions that were queued due to network failures.

When session capture fails (e.g., network down), sessions are stored
locally in ~/.specledger/session-queue/ and can be uploaded later.

Examples:
  sl session sync            # Upload all queued sessions
  sl session sync --status   # Check queue status without uploading
  sl session sync --json     # Output results as JSON`,
	RunE: runSessionSync,
}

// VarSessionPruneCmd represents the prune command
var VarSessionPruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Delete old sessions",
	Long: `Delete sessions older than a specified age threshold.

By default, deletes sessions older than 30 days. Use --days to set a
different threshold, or --expired to use the configured TTL from
specledger.yaml.

Examples:
  sl session prune                  # Delete sessions older than 30 days
  sl session prune --days 14        # Delete sessions older than 14 days
  sl session prune --dry-run        # Preview what would be deleted
  sl session prune --expired        # Use configured TTL from specledger.yaml`,
	RunE: runSessionPrune,
}

// VarSessionStatsCmd represents the stats command
var VarSessionStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show session usage statistics",
	Long: `Display aggregate metrics about session usage.

Shows total count, total storage size, per-branch distribution,
average message count, and date range for sessions in the current project.

Examples:
  sl session stats                  # Show stats for current project
  sl session stats --json           # Output as JSON`,
	RunE: runSessionStats,
}

func init() {
	VarSessionCmd.AddCommand(VarSessionCaptureCmd, VarSessionListCmd, VarSessionGetCmd, VarSessionSyncCmd, VarSessionPruneCmd, VarSessionStatsCmd)

	// Capture flags
	VarSessionCaptureCmd.Flags().Bool("test-mode", false, "Run in test mode with simulated hook input")

	// List flags
	VarSessionListCmd.Flags().String("feature", "", "Feature branch to list sessions for (default: current branch)")
	VarSessionListCmd.Flags().String("commit", "", "Filter by commit hash (partial or full)")
	VarSessionListCmd.Flags().String("task", "", "Filter by task ID (e.g., SL-42)")
	VarSessionListCmd.Flags().String("tag", "", "Filter by tag")
	VarSessionListCmd.Flags().Bool("all-projects", false, "List sessions across all projects")
	VarSessionListCmd.Flags().Int("limit", 0, "Maximum number of sessions to return (0 = unlimited)")
	VarSessionListCmd.Flags().Bool("json", false, "Output as JSON (for scripts/AI)")

	// Get flags
	VarSessionGetCmd.Flags().Bool("json", false, "Output as JSON (for AI processing)")
	VarSessionGetCmd.Flags().Bool("raw", false, "Output raw gzip stream (for piping)")

	// Sync flags
	VarSessionSyncCmd.Flags().Bool("json", false, "Output results as JSON")
	VarSessionSyncCmd.Flags().Bool("status", false, "Check queue status without uploading")

	// Prune flags
	VarSessionPruneCmd.Flags().Int("days", 30, "Delete sessions older than this many days")
	VarSessionPruneCmd.Flags().Bool("dry-run", false, "Preview what would be deleted without removing")
	VarSessionPruneCmd.Flags().Bool("expired", false, "Use configured TTL from specledger.yaml")

	// Stats flags
	VarSessionStatsCmd.Flags().Bool("json", false, "Output as JSON")
}

func runSessionCapture(cmd *cobra.Command, args []string) error {
	testMode, _ := cmd.Flags().GetBool("test-mode")

	var result *session.CaptureResult

	if testMode {
		// Run in test mode with simulated input
		fmt.Fprintln(os.Stderr, "Running in test mode...")
		result = session.CaptureTestMode()
	} else {
		// Normal mode: read from stdin
		result = session.CaptureFromStdin()
	}

	if result.Error != nil {
		// Log error but exit 0 to not block commits
		fmt.Fprintf(os.Stderr, "Session capture warning: %v\n", result.Error)
		return nil
	}

	if result.Captured {
		fmt.Fprintf(os.Stderr, "Session captured: %s (%d messages, %d bytes)\n",
			result.SessionID, result.MessageCount, result.SizeBytes)
	} else if result.Queued {
		fmt.Fprintf(os.Stderr, "Session queued for upload: %s (%d messages, %d bytes)\n",
			result.SessionID, result.MessageCount, result.SizeBytes)
	}

	return nil
}

func runSessionList(cmd *cobra.Command, args []string) error {
	jsonOutput, _ := cmd.Flags().GetBool("json")
	featureBranch, _ := cmd.Flags().GetString("feature")
	commitHash, _ := cmd.Flags().GetString("commit")
	taskID, _ := cmd.Flags().GetString("task")
	tag, _ := cmd.Flags().GetString("tag")
	allProjects, _ := cmd.Flags().GetBool("all-projects")
	limit, _ := cmd.Flags().GetInt("limit")

	// Get access token
	accessToken, err := auth.GetValidAccessToken()
	if err != nil {
		return fmt.Errorf("authentication required: run 'sl auth login' first\n\nDetails: %w", err)
	}

	var projectID string
	if !allProjects {
		// Get current branch if not specified
		if featureBranch == "" {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get working directory: %w", err)
			}
			featureBranch, err = session.GetCurrentBranch(cwd)
			if err != nil {
				return fmt.Errorf("failed to get current branch (not in a git repo?): %w", err)
			}
		}

		// Get project ID
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
		projectID, err = session.GetProjectIDWithFallback(cwd)
		if err != nil {
			return fmt.Errorf("project not configured: %w\n\nHint: Ensure specledger.yaml has 'project.id' set, or the project is registered in Supabase with matching git remote.", err)
		}
	}

	// Query sessions
	client := session.NewMetadataClient()
	opts := &session.QueryOptions{
		ProjectID:     projectID,
		FeatureBranch: featureBranch,
		CommitHash:    commitHash,
		TaskID:        taskID,
		Tag:           tag,
		Limit:         limit,
		OrderBy:       "created_at",
		OrderDesc:     true,
	}

	sessions, err := client.Query(accessToken, opts)
	if err != nil {
		return fmt.Errorf("failed to query sessions: %w", err)
	}

	if jsonOutput {
		data, err := json.MarshalIndent(sessions, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	if len(sessions) == 0 {
		if allProjects {
			fmt.Println("No sessions found across projects")
		} else {
			fmt.Printf("No sessions found for branch '%s'\n", featureBranch)
		}
		return nil
	}

	// Table output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	if allProjects {
		fmt.Fprintln(w, "PROJECT\tBRANCH\tCOMMIT\tMESSAGES\tSIZE\tCAPTURED")
	} else {
		fmt.Fprintln(w, "COMMIT\tMESSAGES\tSIZE\tSTATUS\tCAPTURED")
	}

	for _, s := range sessions {
		commit := "-"
		if s.CommitHash != nil && len(*s.CommitHash) >= 7 {
			commit = (*s.CommitHash)[:7]
		} else if s.TaskID != nil {
			commit = *s.TaskID
		}

		size := formatSize(s.SizeBytes)
		captured := s.CreatedAt.Format("2006-01-02 15:04")

		if allProjects {
			projectShort := s.ProjectID
			if len(projectShort) > 8 {
				projectShort = projectShort[:8]
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%s\n",
				projectShort, s.FeatureBranch, commit, s.MessageCount, size, captured)
		} else {
			fmt.Fprintf(w, "%s\t%d\t%s\t%s\t%s\n",
				commit, s.MessageCount, size, s.Status, captured)
		}
	}

	w.Flush()
	return nil
}

func runSessionGet(cmd *cobra.Command, args []string) error {
	identifier := args[0]
	jsonOutput, _ := cmd.Flags().GetBool("json")
	rawOutput, _ := cmd.Flags().GetBool("raw")

	// Get project ID
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	projectID, err := session.GetProjectIDWithFallback(cwd)
	if err != nil {
		return fmt.Errorf("project not configured: %w\n\nHint: Ensure specledger.yaml has 'project.id' set, or the project is registered in Supabase with matching git remote.", err)
	}

	// Get access token
	accessToken, err := auth.GetValidAccessToken()
	if err != nil {
		return fmt.Errorf("authentication required: run 'sl auth login' first\n\nDetails: %w", err)
	}

	// Try to find the session by different identifiers
	metaClient := session.NewMetadataClient()
	var sessionMeta *session.SessionMetadata

	// Check if identifier looks like a UUID (36 chars with dashes)
	isUUID := len(identifier) == 36 && identifier[8] == '-' && identifier[13] == '-'

	// Try as UUID if it looks like one
	if isUUID {
		sessionMeta, _ = metaClient.GetByID(accessToken, identifier)
	}

	// Try as commit hash
	if sessionMeta == nil {
		sessionMeta, _ = metaClient.GetByCommitHash(accessToken, projectID, identifier)
	}

	// Try as task ID
	if sessionMeta == nil {
		sessionMeta, _ = metaClient.GetByTaskID(accessToken, projectID, identifier)
	}

	if sessionMeta == nil {
		return fmt.Errorf("session not found: %s", identifier)
	}

	// Download content
	storageClient := session.NewStorageClient()
	compressed, err := storageClient.Download(accessToken, sessionMeta.StoragePath)
	if err != nil {
		return fmt.Errorf("failed to download session: %w", err)
	}

	if rawOutput {
		_, err = os.Stdout.Write(compressed)
		return err
	}

	// Decompress
	contentJSON, err := session.Decompress(compressed)
	if err != nil {
		return fmt.Errorf("failed to decompress session: %w", err)
	}

	if jsonOutput {
		fmt.Println(string(contentJSON))
		return nil
	}

	// Parse and format output
	var content session.SessionContent
	if err := json.Unmarshal(contentJSON, &content); err != nil {
		return fmt.Errorf("failed to parse session content: %w", err)
	}

	// Pretty print
	fmt.Printf("Session: %s\n", content.SessionID)
	fmt.Printf("Branch:  %s\n", content.FeatureBranch)
	if content.CommitHash != "" {
		fmt.Printf("Commit:  %s\n", content.CommitHash)
	}
	if content.TaskID != "" {
		fmt.Printf("Task:    %s\n", content.TaskID)
	}
	fmt.Printf("Author:  %s\n", content.Author)
	fmt.Printf("Date:    %s\n", content.CapturedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Messages: %d\n", len(content.Messages))
	fmt.Println(strings.Repeat("-", 60))

	for _, msg := range content.Messages {
		role := strings.ToUpper(msg.Role)
		fmt.Printf("\n[%s] %s\n", role, msg.Timestamp.Format("15:04:05"))
		fmt.Println(msg.Content)
	}

	return nil
}

func runSessionSync(cmd *cobra.Command, args []string) error {
	jsonOutput, _ := cmd.Flags().GetBool("json")
	statusOnly, _ := cmd.Flags().GetBool("status")

	queue := session.NewQueue()

	// Status-only mode: just show queue info without uploading
	if statusOnly {
		entries, err := queue.ListEntries()
		if err != nil {
			return fmt.Errorf("failed to list queue: %w", err)
		}

		if jsonOutput {
			result := map[string]interface{}{
				"queued_count": len(entries),
				"entries":      entries,
			}
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		if len(entries) == 0 {
			fmt.Println("No sessions in queue")
			return nil
		}

		fmt.Printf("%d session(s) queued for upload:\n", len(entries))
		for _, e := range entries {
			commit := "-"
			if e.CommitHash != nil && len(*e.CommitHash) >= 7 {
				commit = (*e.CommitHash)[:7]
			} else if e.TaskID != nil {
				commit = *e.TaskID
			}
			sessionIDShort := e.SessionID
			if len(sessionIDShort) > 8 {
				sessionIDShort = sessionIDShort[:8]
			}
			fmt.Printf("  %s  %s  (retries: %d)\n", sessionIDShort, commit, e.RetryCount)
		}
		return nil
	}

	// Get access token
	accessToken, err := auth.GetValidAccessToken()
	if err != nil {
		return fmt.Errorf("authentication required: run 'sl auth login' first\n\nDetails: %w", err)
	}

	// Load TTL from config for expired session discard
	if cwd, err := os.Getwd(); err == nil {
		if projectMeta, err := metadata.LoadFromProject(cwd); err == nil {
			queue.TTLDays = projectMeta.GetSessionTTLDays()
		}
	}

	uploaded, failed, skipped, errors := queue.ProcessQueue(accessToken)

	if jsonOutput {
		result := map[string]interface{}{
			"uploaded": uploaded,
			"failed":   failed,
			"skipped":  skipped,
			"errors":   errorsToStrings(errors),
		}
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	total := uploaded + failed + skipped
	if total == 0 {
		fmt.Println("No queued sessions to sync")
		return nil
	}

	fmt.Printf("Uploaded %d session(s)\n", uploaded)
	if failed > 0 {
		fmt.Printf("%d session(s) failed (will retry on next sync)\n", failed)
	}
	if skipped > 0 {
		fmt.Printf("%d session(s) skipped (max retries reached)\n", skipped)
	}

	return nil
}

func runSessionPrune(cmd *cobra.Command, args []string) error {
	days, _ := cmd.Flags().GetInt("days")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	expired, _ := cmd.Flags().GetBool("expired")

	// Get access token (required for pruning)
	accessToken, err := auth.GetValidAccessToken()
	if err != nil {
		return fmt.Errorf("authentication required: run 'sl auth login' first\n\nPruning modifies cloud storage and requires authentication.")
	}

	// Get project ID
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	projectID, err := session.GetProjectIDWithFallback(cwd)
	if err != nil {
		return fmt.Errorf("project not configured: %w", err)
	}

	// If --expired, use TTL from config (can be overridden by explicit --days)
	if expired && !cmd.Flags().Changed("days") {
		projectMeta, err := metadata.LoadFromProject(cwd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not load specledger.yaml, using default TTL of 30 days\n")
		} else {
			days = projectMeta.GetSessionTTLDays()
		}
	}

	// Handle TTL of 0 = no automatic expiry
	if days == 0 {
		fmt.Println("TTL is set to 0 (no automatic expiry). No sessions to prune.")
		return nil
	}

	result, err := session.PruneSessions(accessToken, &session.PruneOptions{
		DaysOld:   days,
		DryRun:    dryRun,
		ProjectID: projectID,
	})
	if err != nil {
		return fmt.Errorf("prune failed: %w", err)
	}

	if len(result.Candidates) == 0 {
		fmt.Printf("No sessions older than %d days found.\n", days)
		return nil
	}

	if dryRun {
		fmt.Printf("Dry run: %d session(s) would be deleted (older than %d days):\n\n", len(result.Candidates), days)
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tBRANCH\tSIZE\tCREATED")
		for _, s := range result.Candidates {
			idShort := s.ID
			if len(idShort) > 8 {
				idShort = idShort[:8]
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				idShort, s.FeatureBranch, formatSize(s.SizeBytes), s.CreatedAt.Format("2006-01-02"))
		}
		w.Flush()
		return nil
	}

	fmt.Printf("Pruned %d session(s), %d failed.\n", result.Deleted, result.Failed)
	for _, e := range result.Errors {
		fmt.Fprintf(os.Stderr, "  Error: %v\n", e)
	}

	return nil
}

func runSessionStats(cmd *cobra.Command, args []string) error {
	jsonOutput, _ := cmd.Flags().GetBool("json")

	// Get access token
	accessToken, err := auth.GetValidAccessToken()
	if err != nil {
		return fmt.Errorf("authentication required: run 'sl auth login' first\n\nDetails: %w", err)
	}

	// Get project ID
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	projectID, err := session.GetProjectIDWithFallback(cwd)
	if err != nil {
		return fmt.Errorf("project not configured: %w", err)
	}

	// Query all sessions for this project
	client := session.NewMetadataClient()
	sessions, err := client.Query(accessToken, &session.QueryOptions{
		ProjectID: projectID,
		OrderBy:   "created_at",
		OrderDesc: false,
	})
	if err != nil {
		return fmt.Errorf("failed to query sessions: %w", err)
	}

	if len(sessions) == 0 {
		fmt.Println("No sessions available for this project.")
		return nil
	}

	// Compute aggregate metrics
	var totalSize int64
	var totalMessages int
	branchStats := make(map[string]*branchStat)
	oldest := sessions[0].CreatedAt
	newest := sessions[0].CreatedAt

	for _, s := range sessions {
		totalSize += s.SizeBytes
		totalMessages += s.MessageCount

		if s.CreatedAt.Before(oldest) {
			oldest = s.CreatedAt
		}
		if s.CreatedAt.After(newest) {
			newest = s.CreatedAt
		}

		bs, ok := branchStats[s.FeatureBranch]
		if !ok {
			bs = &branchStat{Branch: s.FeatureBranch}
			branchStats[s.FeatureBranch] = bs
		}
		bs.Count++
		bs.Size += s.SizeBytes
	}

	avgMessages := float64(totalMessages) / float64(len(sessions))

	if jsonOutput {
		branches := make([]branchStat, 0, len(branchStats))
		for _, bs := range branchStats {
			branches = append(branches, *bs)
		}
		result := map[string]interface{}{
			"total_sessions":      len(sessions),
			"total_size_bytes":    totalSize,
			"avg_message_count":   avgMessages,
			"oldest":              oldest,
			"newest":              newest,
			"branch_distribution": branches,
		}
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	// Text output
	fmt.Printf("Session Statistics\n")
	fmt.Printf("==================\n")
	fmt.Printf("Total sessions:   %d\n", len(sessions))
	fmt.Printf("Total storage:    %s\n", formatSize(totalSize))
	fmt.Printf("Avg messages:     %.1f\n", avgMessages)
	fmt.Printf("Date range:       %s to %s\n", oldest.Format("2006-01-02"), newest.Format("2006-01-02"))
	fmt.Printf("\nPer-branch distribution:\n")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "BRANCH\tSESSIONS\tSIZE")
	for _, bs := range branchStats {
		fmt.Fprintf(w, "%s\t%d\t%s\n", bs.Branch, bs.Count, formatSize(bs.Size))
	}
	w.Flush()

	return nil
}

type branchStat struct {
	Branch string `json:"branch"`
	Count  int    `json:"count"`
	Size   int64  `json:"size_bytes"`
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func errorsToStrings(errs []error) []string {
	result := make([]string, len(errs))
	for i, err := range errs {
		result[i] = err.Error()
	}
	return result
}
