package commands

import (
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/specledger/specledger/pkg/cli/auth"
	"github.com/specledger/specledger/pkg/cli/memory"
	"github.com/specledger/specledger/pkg/cli/ui"
	"github.com/spf13/cobra"
)

var (
	memoryStatusFlag    string
	memoryForceFlag     bool
	memoryProjectIDFlag string
)

// VarMemoryCmd is the memory command group
var VarMemoryCmd = &cobra.Command{
	Use:   "memory",
	Short: "Manage project knowledge entries",
	Long: `Manage knowledge entries extracted from AI session transcripts.

Knowledge entries are scored on three axes (Recurrence, Impact, Specificity)
and promoted to the project knowledge base when composite score >= 7.0.

Commands:
  sl memory show       Display knowledge entries with scores and status
  sl memory promote    Promote an entry to the knowledge base
  sl memory demote     Demote an entry back to candidate
  sl memory delete     Delete an entry entirely
  sl memory sync       Push local promoted entries to cloud
  sl memory pull       Download knowledge from cloud to local cache

Examples:
  sl memory show
  sl memory show --status promoted
  sl memory promote KE-a3f5d8
  sl memory demote KE-a3f5d8
  sl memory delete KE-a3f5d8
  sl memory sync --project-id <uuid>
  sl memory pull --project-id <uuid>`,
}

var memoryShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display knowledge entries with scores and status",
	Long: `Display all knowledge entries sorted by composite score.

Optionally filter by status (candidate, promoted, archived).`,
	Example: `  sl memory show
  sl memory show --status promoted
  sl memory show --status candidate`,
	RunE: runMemoryShow,
}

var memoryPromoteCmd = &cobra.Command{
	Use:     "promote <entry-id>",
	Short:   "Promote an entry to the knowledge base",
	Long:    `Promote a knowledge entry to the project's persistent knowledge base, regardless of its score.`,
	Example: `  sl memory promote KE-a3f5d8`,
	Args:    cobra.ExactArgs(1),
	RunE:    runMemoryPromote,
}

var memoryDemoteCmd = &cobra.Command{
	Use:     "demote <entry-id>",
	Short:   "Demote an entry back to candidate",
	Long:    `Remove a knowledge entry from the knowledge base, setting its status back to candidate.`,
	Example: `  sl memory demote KE-a3f5d8`,
	Args:    cobra.ExactArgs(1),
	RunE:    runMemoryDemote,
}

var memoryDeleteCmd = &cobra.Command{
	Use:     "delete <entry-id>",
	Short:   "Delete a knowledge entry entirely",
	Long:    `Permanently remove a knowledge entry from the store.`,
	Example: `  sl memory delete KE-a3f5d8`,
	Args:    cobra.ExactArgs(1),
	RunE:    runMemoryDelete,
}

var memorySyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Push local promoted entries to cloud",
	Long: `Push all locally promoted knowledge entries to the cloud index.
Requires authentication (sl auth login) and a project ID.`,
	Example: `  sl memory sync --project-id <uuid>`,
	RunE:    runMemorySync,
}

var memoryPullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Download knowledge from cloud to local cache",
	Long: `Download knowledge entries from the cloud index to the local cache.
Requires authentication (sl auth login) and a project ID.`,
	Example: `  sl memory pull --project-id <uuid>`,
	RunE:    runMemoryPull,
}

func init() {
	VarMemoryCmd.AddCommand(memoryShowCmd)
	VarMemoryCmd.AddCommand(memoryPromoteCmd)
	VarMemoryCmd.AddCommand(memoryDemoteCmd)
	VarMemoryCmd.AddCommand(memoryDeleteCmd)
	VarMemoryCmd.AddCommand(memorySyncCmd)
	VarMemoryCmd.AddCommand(memoryPullCmd)

	memoryShowCmd.Flags().StringVar(&memoryStatusFlag, "status", "", "Filter by status (candidate, promoted, archived)")
	memoryDeleteCmd.Flags().BoolVar(&memoryForceFlag, "force", false, "Skip confirmation prompt")
	memorySyncCmd.Flags().StringVar(&memoryProjectIDFlag, "project-id", "", "Supabase project ID (required)")
	memoryPullCmd.Flags().StringVar(&memoryProjectIDFlag, "project-id", "", "Supabase project ID (required)")
}

func getMemoryStore() (*memory.Store, error) {
	return memory.NewStore(memory.DefaultCachePath())
}

func runMemoryShow(cmd *cobra.Command, args []string) error {
	store, err := getMemoryStore()
	if err != nil {
		return fmt.Errorf("failed to open store: %w", err)
	}

	var statusFilter *memory.EntryStatus
	if memoryStatusFlag != "" {
		s := memory.EntryStatus(memoryStatusFlag)
		if !memory.IsValidStatus(s) {
			return fmt.Errorf("invalid status: %s (must be candidate, promoted, or archived)", memoryStatusFlag)
		}
		statusFilter = &s
	}

	entries, err := store.List(statusFilter)
	if err != nil {
		return fmt.Errorf("failed to list entries: %w", err)
	}

	if len(entries) == 0 {
		fmt.Println("No knowledge entries found.")
		if memoryStatusFlag != "" {
			fmt.Printf("Try without --status filter, or extract knowledge with /specledger.memory\n")
		}
		return nil
	}

	// Sort by composite score descending
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Scores.Composite > entries[j].Scores.Composite
	})

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "ID\tTITLE\tSCORE\tSTATUS\tTAGS\n")
	for _, e := range entries {
		title := e.Title
		if len(title) > 50 {
			title = title[:47] + "..."
		}
		tags := strings.Join(e.Tags, ",")
		if len(tags) > 30 {
			tags = tags[:27] + "..."
		}
		fmt.Fprintf(w, "%s\t%s\t%.1f\t%s\t%s\n",
			e.ID, title, e.Scores.Composite, e.Status, tags)
	}
	w.Flush()

	fmt.Printf("\n%d entries shown.\n", len(entries))
	return nil
}

func runMemoryPromote(cmd *cobra.Command, args []string) error {
	store, err := getMemoryStore()
	if err != nil {
		return fmt.Errorf("failed to open store: %w", err)
	}

	id := args[0]
	entry, err := store.Promote(id)
	if err != nil {
		return fmt.Errorf("failed to promote entry: %w", err)
	}

	// Regenerate knowledge.md
	if err := memory.RenderKnowledgeMarkdown(store, memory.DefaultKnowledgePath()); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "%s Warning: failed to regenerate knowledge.md: %v\n", ui.WarningIcon(), err)
	}

	fmt.Printf("%s Promoted: %s (%s)\n", ui.Checkmark(), entry.Title, id)
	return nil
}

func runMemoryDemote(cmd *cobra.Command, args []string) error {
	store, err := getMemoryStore()
	if err != nil {
		return fmt.Errorf("failed to open store: %w", err)
	}

	id := args[0]
	entry, err := store.Demote(id)
	if err != nil {
		return fmt.Errorf("failed to demote entry: %w", err)
	}

	// Regenerate knowledge.md
	if err := memory.RenderKnowledgeMarkdown(store, memory.DefaultKnowledgePath()); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "%s Warning: failed to regenerate knowledge.md: %v\n", ui.WarningIcon(), err)
	}

	fmt.Printf("%s Demoted: %s (%s)\n", ui.Checkmark(), entry.Title, id)
	return nil
}

func runMemoryDelete(cmd *cobra.Command, args []string) error {
	store, err := getMemoryStore()
	if err != nil {
		return fmt.Errorf("failed to open store: %w", err)
	}

	id := args[0]

	// Get entry first to show title in confirmation
	entry, err := store.Get(id)
	if err != nil {
		return fmt.Errorf("entry not found: %w", err)
	}

	if !memoryForceFlag {
		fmt.Printf("Delete entry: %s (%s)? [y/N] ", entry.Title, id)
		var confirm string
		fmt.Scanln(&confirm)
		if strings.ToLower(confirm) != "y" && strings.ToLower(confirm) != "yes" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	wasPromoted := entry.Status == memory.StatusPromoted

	if err := store.Delete(id); err != nil {
		return fmt.Errorf("failed to delete entry: %w", err)
	}

	// Regenerate knowledge.md if deleted entry was promoted
	if wasPromoted {
		if err := memory.RenderKnowledgeMarkdown(store, memory.DefaultKnowledgePath()); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "%s Warning: failed to regenerate knowledge.md: %v\n", ui.WarningIcon(), err)
		}
	}

	fmt.Printf("%s Deleted: %s (%s)\n", ui.Checkmark(), entry.Title, id)
	return nil
}

func runMemorySync(cmd *cobra.Command, args []string) error {
	if memoryProjectIDFlag == "" {
		return fmt.Errorf("--project-id is required for cloud sync")
	}

	accessToken, err := auth.GetValidAccessToken()
	if err != nil {
		return fmt.Errorf("authentication required: %w (run 'sl auth login')", err)
	}

	store, err := getMemoryStore()
	if err != nil {
		return fmt.Errorf("failed to open store: %w", err)
	}

	promoted, err := store.ListPromoted()
	if err != nil {
		return fmt.Errorf("failed to list promoted entries: %w", err)
	}

	if len(promoted) == 0 {
		fmt.Println("No promoted entries to sync.")
		return nil
	}

	client := memory.NewSyncClient()
	result, err := client.Push(accessToken, memoryProjectIDFlag, promoted)
	if err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}

	fmt.Printf("%s Synced: %d pushed, %d errors\n", ui.Checkmark(), result.Pushed, result.Errors)
	return nil
}

func runMemoryPull(cmd *cobra.Command, args []string) error {
	if memoryProjectIDFlag == "" {
		return fmt.Errorf("--project-id is required for cloud pull")
	}

	accessToken, err := auth.GetValidAccessToken()
	if err != nil {
		return fmt.Errorf("authentication required: %w (run 'sl auth login')", err)
	}

	client := memory.NewSyncClient()
	entries, err := client.Pull(accessToken, memoryProjectIDFlag)
	if err != nil {
		return fmt.Errorf("pull failed: %w", err)
	}

	if len(entries) == 0 {
		fmt.Println("No knowledge entries found in cloud.")
		return nil
	}

	store, err := getMemoryStore()
	if err != nil {
		return fmt.Errorf("failed to open store: %w", err)
	}

	var created, skipped int
	for _, entry := range entries {
		if err := store.Create(entry); err != nil {
			skipped++
			continue
		}
		created++
	}

	// Regenerate knowledge.md
	if err := memory.RenderKnowledgeMarkdown(store, memory.DefaultKnowledgePath()); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "%s Warning: failed to regenerate knowledge.md: %v\n", ui.WarningIcon(), err)
	}

	fmt.Printf("%s Pulled: %d created, %d skipped (already exist)\n", ui.Checkmark(), created, skipped)
	return nil
}
