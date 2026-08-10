package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/tabwriter"
	"time"

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

// VarSessionPruneCmd represents the prune command
var VarSessionPruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Delete old sessions",
	Long: `Delete sessions older than a configurable threshold.

By default, uses the TTL from specledger.yaml (session.ttl_days, default: 30).
Use --older-than to specify a custom threshold (e.g., 30d, 7d, 90d).
Use --expired to delete sessions past their configured TTL.

Examples:
  sl session prune --older-than 30d          # Delete sessions older than 30 days
  sl session prune --older-than 30d --dry-run # Preview what would be deleted
  sl session prune --expired                  # Delete sessions past their TTL`,
	RunE: runSessionPrune,
}

// VarSessionStatsCmd represents the stats command
var VarSessionStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show session usage statistics",
	Long: `Aggregate view of session usage.

Shows total sessions, total size, sessions per branch, average messages
per session, and oldest/newest session timestamps.

Examples:
  sl session stats                 # Stats for current project
  sl session stats --all-projects  # Global stats
  sl session stats --json          # Machine-readable output`,
	RunE: runSessionStats,
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

func init() {
	VarSessionCmd.AddCommand(VarSessionCaptureCmd, VarSessionListCmd, VarSessionGetCmd, VarSessionSyncCmd, VarSessionPruneCmd, VarSessionStatsCmd)

	// Capture flags
	VarSessionCaptureCmd.Flags().Bool("test-mode", false, "Run in test mode with simulated hook input")

	// List flags
	VarSessionListCmd.Flags().String("feature", "", "Feature branch to list sessions for (default: current branch)")
	VarSessionListCmd.Flags().String("commit", "", "Filter by commit hash (partial or full)")
	VarSessionListCmd.Flags().String("task", "", "Filter by task ID (e.g., SL-42)")
	VarSessionListCmd.Flags().String("tag", "", "Filter by tag (e.g., docker, bugfix, feature)")
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
	VarSessionPruneCmd.Flags().String("older-than", "", "Delete sessions older than duration (e.g., 30d, 7d, 90d)")
	VarSessionPruneCmd.Flags().Bool("expired", false, "Delete sessions past their configured TTL")
	VarSessionPruneCmd.Flags().Bool("dry-run", false, "Preview what would be deleted without deleting")

	// Stats flags
	VarSessionStatsCmd.Flags().Bool("all-projects", false, "Show stats across all projects")
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

	// Get access token
	accessToken, err := auth.GetValidAccessToken()
	if err != nil {
		return fmt.Errorf("authentication required: run 'sl auth login' first\n\nDetails: %w", err)
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
		fmt.Printf("No sessions found for branch '%s'\n", featureBranch)
		return nil
	}

	// Table output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "COMMIT\tMESSAGES\tSIZE\tSTATUS\tCAPTURED")

	for _, s := range sessions {
		commit := "-"
		if s.CommitHash != nil {
			commit = (*s.CommitHash)[:7]
		} else if s.TaskID != nil {
			commit = *s.TaskID
		}

		size := formatSize(s.SizeBytes)
		captured := s.CreatedAt.Format("2006-01-02 15:04")

		fmt.Fprintf(w, "%s\t%d\t%s\t%s\t%s\n",
			commit, s.MessageCount, size, s.Status, captured)
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

// parseDuration parses a duration string like "30d", "7d", "90d" into a time.Duration
func parseDuration(s string) (time.Duration, error) {
	re := regexp.MustCompile(`^(\d+)d$`)
	matches := re.FindStringSubmatch(s)
	if matches == nil {
		return 0, fmt.Errorf("invalid duration format %q (expected e.g. 30d)", s)
	}
	days := 0
	_, err := fmt.Sscanf(matches[1], "%d", &days)
	if err != nil {
		return 0, err
	}
	return time.Duration(days) * 24 * time.Hour, nil
}

func runSessionPrune(cmd *cobra.Command, args []string) error {
	olderThan, _ := cmd.Flags().GetString("older-than")
	expired, _ := cmd.Flags().GetBool("expired")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	if olderThan == "" && !expired {
		return fmt.Errorf("must specify --older-than or --expired")
	}

	// Determine cutoff time
	var cutoff time.Time
	if expired {
		// Use TTL from config
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
		meta, err := metadata.Load(cwd)
		if err != nil {
			// Default to 30 days if config not found
			cutoff = time.Now().AddDate(0, 0, -30)
		} else {
			ttlDays := meta.GetSessionTTLDays()
			cutoff = time.Now().AddDate(0, 0, -ttlDays)
		}
	} else {
		dur, err := parseDuration(olderThan)
		if err != nil {
			return err
		}
		cutoff = time.Now().Add(-dur)
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

	// Get access token
	accessToken, err := auth.GetValidAccessToken()
	if err != nil {
		return fmt.Errorf("authentication required: run 'sl auth login' first\n\nDetails: %w", err)
	}

	// Query sessions older than cutoff
	client := session.NewMetadataClient()
	endDate := cutoff
	sessions, err := client.Query(accessToken, &session.QueryOptions{
		ProjectID: projectID,
		EndDate:   &endDate,
		OrderBy:   "created_at",
		OrderDesc: false,
	})
	if err != nil {
		return fmt.Errorf("failed to query sessions: %w", err)
	}

	if len(sessions) == 0 {
		fmt.Println("No sessions to prune")
		return nil
	}

	if dryRun {
		fmt.Printf("Would delete %d session(s) created before %s:\n", len(sessions), cutoff.Format("2006-01-02"))
		for _, s := range sessions {
			commit := "-"
			if s.CommitHash != nil && len(*s.CommitHash) >= 7 {
				commit = (*s.CommitHash)[:7]
			}
			fmt.Printf("  %s  %s  %s  %s\n", s.ID[:8], commit, formatSize(s.SizeBytes), s.CreatedAt.Format("2006-01-02 15:04"))
		}
		return nil
	}

	// Delete sessions
	storageClient := session.NewStorageClient()
	deleted := 0
	var errs []error

	for _, s := range sessions {
		// Delete storage object
		if err := storageClient.Delete(accessToken, s.StoragePath); err != nil {
			errs = append(errs, fmt.Errorf("storage delete %s: %w", s.ID[:8], err))
			continue
		}
		// Delete metadata row
		if err := client.Delete(accessToken, s.ID); err != nil {
			errs = append(errs, fmt.Errorf("metadata delete %s: %w", s.ID[:8], err))
			continue
		}
		deleted++
	}

	fmt.Printf("Deleted %d session(s)\n", deleted)
	if len(errs) > 0 {
		fmt.Printf("%d error(s) occurred:\n", len(errs))
		for _, e := range errs {
			fmt.Printf("  %s\n", e)
		}
		return fmt.Errorf("prune completed with errors")
	}

	return nil
}

// SessionStats holds aggregated session statistics
type SessionStats struct {
	TotalSessions    int              `json:"total_sessions"`
	TotalSize        int64            `json:"total_size_bytes"`
	TotalRawSize     int64            `json:"total_raw_size_bytes"`
	AvgMessages      float64          `json:"avg_messages_per_session"`
	OldestSession    *time.Time       `json:"oldest_session,omitempty"`
	NewestSession    *time.Time       `json:"newest_session,omitempty"`
	SessionsByBranch map[string]int   `json:"sessions_by_branch"`
	SizeByBranch     map[string]int64 `json:"size_by_branch"`
}

func runSessionStats(cmd *cobra.Command, args []string) error {
	jsonOutput, _ := cmd.Flags().GetBool("json")
	allProjects, _ := cmd.Flags().GetBool("all-projects")

	var projectID string
	if !allProjects {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
		projectID, err = session.GetProjectIDWithFallback(cwd)
		if err != nil {
			return fmt.Errorf("project not configured: %w", err)
		}
	}

	accessToken, err := auth.GetValidAccessToken()
	if err != nil {
		return fmt.Errorf("authentication required: run 'sl auth login' first\n\nDetails: %w", err)
	}

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
		fmt.Println("No sessions found")
		return nil
	}

	// Compute stats
	stats := SessionStats{
		TotalSessions:    len(sessions),
		SessionsByBranch: make(map[string]int),
		SizeByBranch:     make(map[string]int64),
	}

	totalMessages := 0
	for i, s := range sessions {
		stats.TotalSize += s.SizeBytes
		stats.TotalRawSize += s.RawSizeBytes
		totalMessages += s.MessageCount
		stats.SessionsByBranch[s.FeatureBranch]++
		stats.SizeByBranch[s.FeatureBranch] += s.SizeBytes

		if i == 0 {
			t := s.CreatedAt
			stats.OldestSession = &t
		}
		if i == len(sessions)-1 {
			t := s.CreatedAt
			stats.NewestSession = &t
		}
	}

	if stats.TotalSessions > 0 {
		stats.AvgMessages = float64(totalMessages) / float64(stats.TotalSessions)
	}

	if jsonOutput {
		data, err := json.MarshalIndent(stats, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	// Pretty print
	fmt.Printf("Total sessions:  %d\n", stats.TotalSessions)
	fmt.Printf("Total size:      %s (compressed), %s (raw)\n", formatSize(stats.TotalSize), formatSize(stats.TotalRawSize))
	fmt.Printf("Avg messages:    %.1f per session\n", stats.AvgMessages)
	if stats.OldestSession != nil {
		fmt.Printf("Oldest session:  %s\n", stats.OldestSession.Format("2006-01-02 15:04"))
	}
	if stats.NewestSession != nil {
		fmt.Printf("Newest session:  %s\n", stats.NewestSession.Format("2006-01-02 15:04"))
	}

	fmt.Printf("\nSessions by branch:\n")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "BRANCH\tSESSIONS\tSIZE")
	for branch, count := range stats.SessionsByBranch {
		fmt.Fprintf(w, "%s\t%d\t%s\n", branch, count, formatSize(stats.SizeByBranch[branch]))
	}
	w.Flush()

	return nil
}
