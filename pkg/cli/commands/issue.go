package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/specledger/specledger/pkg/cli/metadata"
	"github.com/specledger/specledger/pkg/cli/ui"
	"github.com/specledger/specledger/pkg/issues"
	"github.com/spf13/cobra"
)

var (
	// Issue command flags
	issueTitleFlag       string
	issueDescFlag        string
	issueTypeFlag        string
	issuePriorityFlag    int
	issueLabelsFlag      string
	issueSpecFlag        string
	issueForceFlag       bool
	issueStatusFlag      string
	issueAllFlag         bool
	issueJSONFlag        bool
	issueTreeFlag        bool
	issueGraphFlag       bool // Show blocking dependency graph
	issueBlockedFlag     bool
	issueReasonFlag      string
	issueAssigneeFlag    string
	issueNotesFlag       string
	issueDesignFlag      string
	issueAcceptFlag      string
	issueAddLabelFlag    string
	issueRemoveLabelFlag string
	issueDryRunFlag      bool
	issueKeepBeadsFlag   bool
	issueDoDFlag         []string // Definition of Done items (repeatable)
	issueCheckDoDFlag    string   // Mark DoD item as checked
	issueUncheckDoDFlag  string   // Mark DoD item as unchecked
	issueParentFlag      string   // Parent issue ID
)

// getArtifactPath loads the artifact_path from specledger.yaml
// Falls back to "specledger/" on error or if not configured
func getArtifactPath() string {
	meta, err := metadata.LoadFromProject(".")
	if err != nil {
		return "specledger/"
	}
	return meta.GetArtifactPath()
}

// VarIssueCmd is the issue command group
var VarIssueCmd = &cobra.Command{
	Use:   "issue",
	Short: "Manage issues for the current spec",
	Long: `Manage issues for tracking work within a spec.

Issues are stored in JSONL format at specledger/<spec>/issues.jsonl.
Each issue has a globally unique ID (SL-xxxxxx format).

Commands:
  sl issue create    Create a new issue
  sl issue list      List issues
  sl issue show      Show issue details
  sl issue update    Update an issue
  sl issue close     Close an issue
  sl issue link      Link issues with dependencies
  sl issue unlink    Remove dependency links
  sl issue migrate   Migrate from Beads format
  sl issue repair    Repair corrupted issues.jsonl

Examples:
  sl issue create --title "Add validation" --type task
  sl issue list --status open
  sl issue show SL-a3f5d8
  sl issue update SL-a3f5d8 --status in_progress
  sl issue close SL-a3f5d8`,
}

// issueCreateCmd creates a new issue
var issueCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new issue",
	Long: `Create a new issue in the current spec.

The issue will be assigned a globally unique ID in the format SL-xxxxxx
derived from SHA-256 hash of (spec_context + title + created_at).`,
	Example: `  sl issue create --title "Add validation" --type task
  sl issue create --title "Fix auth bug" --type bug --priority 0
  sl issue create --title "Feature" --description "Details" --labels "component:api"`,
	RunE: runIssueCreate,
}

// issueListCmd lists issues
var issueListCmd = &cobra.Command{
	Use:   "list",
	Short: "List issues",
	Long: `List issues for the current spec or across all specs.

Supports various filters and output formats.`,
	Example: `  sl issue list
  sl issue list --status open
  sl issue list --all
  sl issue list --spec 010-my-feature`,
	RunE: runIssueList,
}

// issueShowCmd shows issue details
var issueShowCmd = &cobra.Command{
	Use:   "show <issue-id>",
	Short: "Show issue details",
	Long:  `Display full details of an issue including all fields.`,
	Example: `  sl issue show SL-a3f5d8
  sl issue show SL-a3f5d8 --json`,
	Args: cobra.ExactArgs(1),
	RunE: runIssueShow,
}

// issueUpdateCmd updates an issue
var issueUpdateCmd = &cobra.Command{
	Use:   "update <issue-id>",
	Short: "Update an issue",
	Long:  `Update fields of an existing issue.`,
	Example: `  sl issue update SL-a3f5d8 --status in_progress
  sl issue update SL-a3f5d8 --priority 0 --assignee alice`,
	Args: cobra.ExactArgs(1),
	RunE: runIssueUpdate,
}

// issueCloseCmd closes an issue
var issueCloseCmd = &cobra.Command{
	Use:   "close <issue-id>",
	Short: "Close an issue",
	Long: `Close an issue, marking it as complete.

If the issue has a definition_of_done field, all items must be checked
before closing (use --force to bypass).`,
	Example: `  sl issue close SL-a3f5d8
  sl issue close SL-a3f5d8 --reason "Completed in PR #42"`,
	Args: cobra.ExactArgs(1),
	RunE: runIssueClose,
}

// issueMigrateCmd migrates from Beads
var issueMigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate from Beads format",
	Long: `Migrate existing Beads issues to the new per-spec format.

This command reads .beads/issues.jsonl and distributes issues to their
respective spec directories. After successful migration, it removes the
.beads directory and cleans up mise.toml.`,
	Example: `  sl issue migrate
  sl issue migrate --dry-run
  sl issue migrate --keep-beads`,
	RunE: runIssueMigrate,
}

// issueRepairCmd repairs corrupted issues.jsonl files
var issueRepairCmd = &cobra.Command{
	Use:   "repair",
	Short: "Repair corrupted issues.jsonl files",
	Long: `Repair corrupted issues.jsonl files by parsing valid JSON lines
and skipping invalid entries. Creates a backup before modification.`,
	Example: `  sl issue repair
  sl issue repair --spec 010-my-feature`,
	RunE: runIssueRepair,
}

// issueLinkCmd links issues with dependencies
var issueLinkCmd = &cobra.Command{
	Use:   "link <from-issue-id> <type> <to-issue-id>",
	Short: "Create a dependency link between issues",
	Long: `Create a dependency relationship between two issues.

Link types:
  blocks  - from blocks to (from must complete before to can start)
  related - from and to are related (soft link)`,
	Example: `  sl issue link SL-a3f5d8 blocks SL-b4e6f9
  sl issue link SL-a3f5d8 related SL-c7e1a2`,
	Args: cobra.ExactArgs(3),
	RunE: runIssueLink,
}

// issueUnlinkCmd removes dependency links
var issueUnlinkCmd = &cobra.Command{
	Use:     "unlink <from-issue-id> <type> <to-issue-id>",
	Short:   "Remove a dependency link between issues",
	Example: `  sl issue unlink SL-a3f5d8 blocks SL-b4e6f9`,
	Args:    cobra.ExactArgs(3),
	RunE:    runIssueUnlink,
}

// issueReadyCmd lists ready-to-work issues
var issueReadyCmd = &cobra.Command{
	Use:   "ready",
	Short: "List issues ready to work on",
	Long: `List issues that are ready to work on (not blocked by open dependencies).

An issue is "ready" when:
- Status is "open" or "in_progress"
- AND all issues that block it are closed

If all issues are blocked, displays which issues are blocking them.`,
	Example: `  sl issue ready
  sl issue ready --all
  sl issue ready --json`,
	RunE: runIssueReady,
}

func init() {
	// Add issue command to root
	VarIssueCmd.AddCommand(issueCreateCmd)
	VarIssueCmd.AddCommand(issueListCmd)
	VarIssueCmd.AddCommand(issueShowCmd)
	VarIssueCmd.AddCommand(issueUpdateCmd)
	VarIssueCmd.AddCommand(issueCloseCmd)
	VarIssueCmd.AddCommand(issueLinkCmd)
	VarIssueCmd.AddCommand(issueUnlinkCmd)
	VarIssueCmd.AddCommand(issueReadyCmd)
	VarIssueCmd.AddCommand(issueMigrateCmd)
	VarIssueCmd.AddCommand(issueRepairCmd)

	// Create command flags
	issueCreateCmd.Flags().StringVar(&issueTitleFlag, "title", "", "Issue title (required)")
	issueCreateCmd.Flags().StringVar(&issueDescFlag, "description", "", "Issue description")
	issueCreateCmd.Flags().StringVar(&issueTypeFlag, "type", "task", "Issue type (epic, feature, task, bug)")
	issueCreateCmd.Flags().IntVarP(&issuePriorityFlag, "priority", "p", 2, "Priority (0-5, 0=highest)")
	issueCreateCmd.Flags().StringVar(&issueLabelsFlag, "labels", "", "Comma-separated labels")
	issueCreateCmd.Flags().StringVar(&issueSpecFlag, "spec", "", "Override spec context")
	issueCreateCmd.Flags().BoolVar(&issueForceFlag, "force", false, "Skip duplicate detection")
	issueCreateCmd.Flags().StringVar(&issueAcceptFlag, "acceptance-criteria", "", "Acceptance criteria text")
	issueCreateCmd.Flags().StringArrayVar(&issueDoDFlag, "dod", []string{}, "Definition of Done items (can be repeated)")
	issueCreateCmd.Flags().StringVar(&issueDesignFlag, "design", "", "Design notes/approach")
	issueCreateCmd.Flags().StringVar(&issueNotesFlag, "notes", "", "Implementation notes")
	issueCreateCmd.Flags().StringVar(&issueParentFlag, "parent", "", "Parent issue ID")
	if err := issueCreateCmd.MarkFlagRequired("title"); err != nil {
		panic(fmt.Sprintf("failed to mark title flag as required: %v", err))
	}

	// List command flags
	issueListCmd.Flags().StringVar(&issueStatusFlag, "status", "", "Filter by status (open, in_progress, closed)")
	issueListCmd.Flags().StringVar(&issueTypeFlag, "type", "", "Filter by type")
	issueListCmd.Flags().IntVarP(&issuePriorityFlag, "priority", "p", -1, "Filter by priority")
	issueListCmd.Flags().StringVar(&issueLabelsFlag, "label", "", "Filter by label")
	issueListCmd.Flags().StringVar(&issueSpecFlag, "spec", "", "Filter by spec context")
	issueListCmd.Flags().BoolVar(&issueAllFlag, "all", false, "List across all specs")
	issueListCmd.Flags().BoolVar(&issueJSONFlag, "json", false, "Output as JSON")
	issueListCmd.Flags().BoolVar(&issueTreeFlag, "tree", false, "Show parent-child hierarchy tree")
	issueListCmd.Flags().BoolVar(&issueGraphFlag, "graph", false, "Show blocking dependency graph")
	issueListCmd.Flags().BoolVar(&issueBlockedFlag, "blocked", false, "Show only blocked issues")

	// Show command flags
	issueShowCmd.Flags().BoolVar(&issueJSONFlag, "json", false, "Output as JSON")
	issueShowCmd.Flags().BoolVar(&issueTreeFlag, "tree", false, "Show dependency tree")

	// Update command flags
	issueUpdateCmd.Flags().StringVar(&issueTitleFlag, "title", "", "Update title")
	issueUpdateCmd.Flags().StringVar(&issueDescFlag, "description", "", "Update description")
	issueUpdateCmd.Flags().StringVar(&issueStatusFlag, "status", "", "Update status")
	issueUpdateCmd.Flags().IntVarP(&issuePriorityFlag, "priority", "p", -1, "Update priority")
	issueUpdateCmd.Flags().StringVar(&issueAssigneeFlag, "assignee", "", "Update assignee")
	issueUpdateCmd.Flags().StringVar(&issueNotesFlag, "notes", "", "Update notes")
	issueUpdateCmd.Flags().StringVar(&issueDesignFlag, "design", "", "Update design notes")
	issueUpdateCmd.Flags().StringVar(&issueAcceptFlag, "acceptance-criteria", "", "Update acceptance criteria")
	issueUpdateCmd.Flags().StringVar(&issueAddLabelFlag, "add-label", "", "Add a label")
	issueUpdateCmd.Flags().StringVar(&issueRemoveLabelFlag, "remove-label", "", "Remove a label")
	issueUpdateCmd.Flags().StringArrayVar(&issueDoDFlag, "dod", []string{}, "Replace Definition of Done items (can be repeated)")
	issueUpdateCmd.Flags().StringVar(&issueCheckDoDFlag, "check-dod", "", "Mark DoD item as checked (exact match)")
	issueUpdateCmd.Flags().StringVar(&issueUncheckDoDFlag, "uncheck-dod", "", "Mark DoD item as unchecked (exact match)")
	issueUpdateCmd.Flags().StringVar(&issueParentFlag, "parent", "", "Set parent issue ID (empty string to clear)")

	// Close command flags
	issueCloseCmd.Flags().StringVar(&issueReasonFlag, "reason", "", "Close reason")
	issueCloseCmd.Flags().BoolVar(&issueForceFlag, "force", false, "Skip definition of done check")

	// Migrate command flags
	issueMigrateCmd.Flags().BoolVar(&issueDryRunFlag, "dry-run", false, "Show what would be migrated")
	issueMigrateCmd.Flags().BoolVar(&issueKeepBeadsFlag, "keep-beads", false, "Keep .beads folder after migration")

	// Ready command flags
	issueReadyCmd.Flags().BoolVar(&issueAllFlag, "all", false, "List ready issues across all specs")
	issueReadyCmd.Flags().BoolVar(&issueJSONFlag, "json", false, "Output as JSON")
}

func runIssueCreate(cmd *cobra.Command, args []string) error {
	// Get spec context
	specContext := issueSpecFlag
	if specContext == "" {
		detector := issues.NewContextDetector(".")
		var err error
		specContext, err = detector.DetectSpecContext()
		if err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	// Create issue
	issueType := issues.IssueType(issueTypeFlag)
	if !issues.IsValidIssueType(issueType) {
		return fmt.Errorf("invalid issue type: %s (must be epic, feature, task, or bug)", issueTypeFlag)
	}

	issue := issues.NewIssue(issueTitleFlag, issueDescFlag, specContext, issueType, issuePriorityFlag)

	// Add labels
	if issueLabelsFlag != "" {
		issue.Labels = strings.Split(issueLabelsFlag, ",")
		for i, l := range issue.Labels {
			issue.Labels[i] = strings.TrimSpace(l)
		}
	}

	// Add structured fields from flags
	if issueAcceptFlag != "" {
		issue.AcceptanceCriteria = issueAcceptFlag
	}
	if len(issueDoDFlag) > 0 {
		items := make([]issues.ChecklistItem, len(issueDoDFlag))
		for i, item := range issueDoDFlag {
			items[i] = issues.ChecklistItem{
				Item:    item,
				Checked: false,
			}
		}
		issue.DefinitionOfDone = &issues.DefinitionOfDone{Items: items}
	}
	if issueDesignFlag != "" {
		issue.Design = issueDesignFlag
	}
	if issueNotesFlag != "" {
		issue.Notes = issueNotesFlag
	}
	if issueParentFlag != "" {
		issue.ParentID = &issueParentFlag
	}

	// Create store and save
	store, err := issues.NewStore(issues.StoreOptions{
		BasePath:    getArtifactPath(),
		SpecContext: specContext,
	})
	if err != nil {
		return fmt.Errorf("failed to create store: %w", err)
	}

	if err := store.Create(issue); err != nil {
		return fmt.Errorf("failed to create issue: %w", err)
	}

	if issueJSONFlag {
		data, _ := json.MarshalIndent(issue, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Printf("%s Created issue %s\n", ui.Checkmark(), issue.ID)
		fmt.Printf("  Title: %s\n", issue.Title)
		fmt.Printf("  Type: %s\n", issue.IssueType)
		fmt.Printf("  Priority: %d\n", issue.Priority)
		fmt.Printf("  Spec: %s\n", issue.SpecContext)
		fmt.Println()
		fmt.Printf("View: sl issue show %s\n", issue.ID)
	}

	return nil
}

func runIssueList(cmd *cobra.Command, args []string) error {
	// Determine spec context
	specContext := issueSpecFlag
	if specContext == "" && !issueAllFlag {
		detector := issues.NewContextDetector(".")
		var err error
		specContext, err = detector.DetectSpecContext()
		if err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	// Build filter
	filter := issues.ListFilter{}
	if issueStatusFlag != "" {
		status := issues.IssueStatus(issueStatusFlag)
		filter.Status = &status
	}
	if issueTypeFlag != "" {
		issueType := issues.IssueType(issueTypeFlag)
		filter.IssueType = &issueType
	}
	if issuePriorityFlag >= 0 {
		filter.Priority = &issuePriorityFlag
	}
	if issueLabelsFlag != "" {
		filter.Labels = []string{issueLabelsFlag}
	}
	filter.SpecContext = specContext
	filter.All = issueAllFlag
	filter.Blocked = issueBlockedFlag

	var issueList []issues.Issue
	var err error

	// Get issues - use cross-spec listing if --all flag is set
	artifactPath := getArtifactPath()
	if issueAllFlag {
		issueList, err = issues.ListAllSpecs(artifactPath, filter)
		if err != nil {
			return fmt.Errorf("failed to list issues across specs: %w", err)
		}
	} else {
		store, storeErr := issues.NewStore(issues.StoreOptions{
			BasePath:    artifactPath,
			SpecContext: specContext,
		})
		if storeErr != nil {
			return fmt.Errorf("failed to create store: %w", storeErr)
		}

		issueList, err = store.List(filter)
		if err != nil {
			return fmt.Errorf("failed to list issues: %w", err)
		}
	}

	if issueJSONFlag {
		data, _ := json.MarshalIndent(issueList, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	if len(issueList) == 0 {
		fmt.Println("No issues found.")
		return nil
	}

	// Handle tree view (parent-child hierarchy)
	if issueTreeFlag {
		return renderIssueTree(issueList, specContext, issueAllFlag, artifactPath)
	}

	// Handle graph view (blocking dependencies)
	if issueGraphFlag {
		return renderDependencyGraph(issueList, specContext, issueAllFlag, artifactPath)
	}

	// Print table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTITLE\tSTATUS\tTYPE\tPRIORITY\tSPEC")
	for _, issue := range issueList {
		title := issue.Title
		if len(title) > 40 {
			title = title[:37] + "..."
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%s\n",
			issue.ID, title, issue.Status, issue.IssueType, issue.Priority, issue.SpecContext)
	}
	w.Flush()

	return nil
}

// renderIssueTree renders issues in a hierarchical tree format
func renderIssueTree(issueList []issues.Issue, specContext string, allFlag bool, artifactPath string) error {
	// Group issues by spec if --all flag is set
	if allFlag {
		return renderCrossSpecTree(issueList, artifactPath)
	}

	// Build dependency trees for single spec
	store, err := issues.NewStore(issues.StoreOptions{
		BasePath:    artifactPath,
		SpecContext: specContext,
	})
	if err != nil {
		return fmt.Errorf("failed to create store: %w", err)
	}

	return renderSingleSpecTree(issueList, store, specContext)
}

// renderSingleSpecTree renders issues for a single spec in tree format
func renderSingleSpecTree(issueList []issues.Issue, store *issues.Store, specContext string) error {
	// Build hierarchy trees based on parent-child relationships
	trees, err := store.GetHierarchyForest()
	if err != nil {
		return fmt.Errorf("failed to build hierarchy tree: %w", err)
	}

	// Create renderer
	renderer := issues.NewTreeRenderer(issues.DefaultTreeRenderOptions())

	// Render hierarchy tree
	output := renderer.RenderHierarchyForest(specContext, trees, len(issueList))
	fmt.Print(output)

	return nil
}

// renderDependencyGraph renders issues as a blocking dependency graph
func renderDependencyGraph(issueList []issues.Issue, specContext string, allFlag bool, artifactPath string) error {
	// Group issues by spec if --all flag is set
	if allFlag {
		return renderCrossSpecGraph(issueList, artifactPath)
	}

	// Build store for single spec
	store, err := issues.NewStore(issues.StoreOptions{
		BasePath:    artifactPath,
		SpecContext: specContext,
	})
	if err != nil {
		return fmt.Errorf("failed to create store: %w", err)
	}

	return renderSingleSpecGraph(issueList, store, specContext)
}

// renderSingleSpecGraph renders blocking dependencies as a graph
func renderSingleSpecGraph(issueList []issues.Issue, store *issues.Store, specContext string) error {
	// Build issue map
	issueMap := make(map[string]*issues.Issue)
	for i := range issueList {
		issueMap[issueList[i].ID] = &issueList[i]
	}

	// Collect edges
	edges := collectAllEdges(issueList, issueMap)

	// Find isolated nodes
	nodesWithEdges := make(map[string]bool)
	for _, edge := range edges {
		nodesWithEdges[edge.from] = true
		nodesWithEdges[edge.to] = true
	}

	var isolatedNodes []issues.Issue
	for _, issue := range issueList {
		if !nodesWithEdges[issue.ID] {
			isolatedNodes = append(isolatedNodes, issue)
		}
	}

	// Find connected components
	components := findConnectedComponents(issueList, issueMap)

	// Output header
	fmt.Printf("%s (dependency graph)\n\n", specContext)

	if len(edges) == 0 && len(isolatedNodes) == 0 {
		fmt.Println("No issues found.")
		return nil
	}

	renderer := issues.NewTreeRenderer(issues.DefaultTreeRenderOptions())

	// Summary line
	parts := []string{}
	if len(edges) > 0 {
		parts = append(parts, fmt.Sprintf("%d edges", len(edges)))
	}
	if len(isolatedNodes) > 0 {
		parts = append(parts, fmt.Sprintf("%d isolated", len(isolatedNodes)))
	}
	if len(components) > 1 {
		parts = append(parts, fmt.Sprintf("%d components", len(components)))
	}
	if len(parts) > 0 {
		fmt.Printf("[%s]\n\n", strings.Join(parts, ", "))
	}

	// Render each component
	for i, component := range components {
		if len(components) > 1 {
			fmt.Printf("── Component %d (%d nodes) ──\n", i+1, len(component))
		}

		// Get edges for this component only
		componentSet := make(map[string]bool)
		for _, iss := range component {
			componentSet[iss.ID] = true
		}

		var componentEdges []edge
		for _, e := range edges {
			if componentSet[e.from] && componentSet[e.to] {
				componentEdges = append(componentEdges, e)
			}
		}

		// Group edges by source
		outgoing := make(map[string][]string)
		for _, e := range componentEdges {
			outgoing[e.from] = append(outgoing[e.from], e.to)
		}

		// Find root nodes in this component (not blocked by anything in component)
		blocked := make(map[string]bool)
		for _, iss := range component {
			for _, blockerID := range iss.BlockedBy {
				if componentSet[blockerID] {
					blocked[iss.ID] = true
				}
			}
		}

		// Render starting from root nodes
		visited := make(map[string]bool)
		for _, iss := range component {
			if !blocked[iss.ID] && len(outgoing[iss.ID]) > 0 {
				renderGraphNode(iss.ID, issueMap, outgoing, renderer, visited, "")
			}
		}

		fmt.Println()
	}

	// Render isolated nodes
	if len(isolatedNodes) > 0 {
		fmt.Printf("── Isolated (%d) ──\n", len(isolatedNodes))
		for _, node := range isolatedNodes {
			fmt.Printf("  %s\n", renderer.FormatIssueSimple(node))
		}
	}

	return nil
}

// renderGraphNode renders a node and what it blocks recursively with topological ordering
func renderGraphNode(nodeID string, issueMap map[string]*issues.Issue, outgoing map[string][]string, renderer *issues.TreeRenderer, visited map[string]bool, prefix string) {
	node, exists := issueMap[nodeID]
	if !exists {
		return
	}

	// If already visited, don't render again
	if visited[nodeID] {
		return
	}
	visited[nodeID] = true

	// Render this node
	fmt.Printf("%s%s\n", prefix, renderer.FormatIssueSimple(*node))

	// Get targets this node blocks
	targets := outgoing[nodeID]
	if len(targets) == 0 {
		return
	}

	// Sort targets: sinks (no outgoing) first
	sortedTargets := make([]string, len(targets))
	copy(sortedTargets, targets)
	for i := 0; i < len(sortedTargets)-1; i++ {
		for j := i + 1; j < len(sortedTargets); j++ {
			outI := len(outgoing[sortedTargets[i]])
			outJ := len(outgoing[sortedTargets[j]])
			// Sinks first
			if outJ == 0 && outI > 0 {
				sortedTargets[i], sortedTargets[j] = sortedTargets[j], sortedTargets[i]
			}
		}
	}

	for i, targetID := range sortedTargets {
		isLast := i == len(sortedTargets)-1

		// Determine connector and child prefix
		var connector, childPrefix string
		if isLast {
			connector = "└─>"
			childPrefix = prefix + "   "
		} else {
			connector = "├─>"
			childPrefix = prefix + "│  "
		}

		// Add blank line with vertical connector for non-last children
		if isLast {
			fmt.Println()
		} else {
			fmt.Printf("%s│\n", prefix)
		}

		// Check if already visited
		if visited[targetID] {
			childNode := issueMap[targetID]
			fmt.Printf("%s%s %s (see above)\n", prefix, connector, renderer.FormatIssueSimple(*childNode))
		} else {
			// Print connector + node on same line
			childNode := issueMap[targetID]
			fmt.Printf("%s%s %s\n", prefix, connector, renderer.FormatIssueSimple(*childNode))
			// Mark as visited before recursing
			visited[targetID] = true
			// Recurse for children
			renderGraphNode(targetID, issueMap, outgoing, renderer, visited, childPrefix)
		}
	}
}

// renderCrossSpecGraph renders blocking dependencies across all specs as a graph
func renderCrossSpecGraph(issueList []issues.Issue, artifactPath string) error {
	// Group issues by spec
	specIssues := make(map[string][]issues.Issue)
	for _, issue := range issueList {
		specIssues[issue.SpecContext] = append(specIssues[issue.SpecContext], issue)
	}

	// Sort specs for consistent output
	specNames := make([]string, 0, len(specIssues))
	for spec := range specIssues {
		specNames = append(specNames, spec)
	}

	fmt.Println("All Specs (dependency graph)")
	fmt.Println()

	renderer := issues.NewTreeRenderer(issues.DefaultTreeRenderOptions())

	for _, spec := range specNames {
		issuesInSpec := specIssues[spec]
		issueMap := make(map[string]*issues.Issue)
		for i := range issuesInSpec {
			issueMap[issuesInSpec[i].ID] = &issuesInSpec[i]
		}

		edges := collectAllEdges(issuesInSpec, issueMap)
		components := findConnectedComponents(issuesInSpec, issueMap)

		// Count isolated
		nodesWithEdges := make(map[string]bool)
		for _, edge := range edges {
			nodesWithEdges[edge.from] = true
			nodesWithEdges[edge.to] = true
		}
		isolated := 0
		for _, iss := range issuesInSpec {
			if !nodesWithEdges[iss.ID] {
				isolated++
			}
		}

		fmt.Printf("=== %s ===\n", spec)
		fmt.Printf("%d issues, %d edges", len(issuesInSpec), len(edges))
		if len(components) > 1 {
			fmt.Printf(", %d components", len(components))
		}
		if isolated > 0 {
			fmt.Printf(", %d isolated", isolated)
		}
		fmt.Println()
		fmt.Println()

		// Render each component
		for i, component := range components {
			if len(components) > 1 {
				fmt.Printf("── Component %d (%d nodes) ──\n", i+1, len(component))
			}

			// Get edges for this component
			componentSet := make(map[string]bool)
			for _, iss := range component {
				componentSet[iss.ID] = true
			}

			var componentEdges []edge
			for _, e := range edges {
				if componentSet[e.from] && componentSet[e.to] {
					componentEdges = append(componentEdges, e)
				}
			}

			// Group edges by source
			outgoing := make(map[string][]string)
			for _, e := range componentEdges {
				outgoing[e.from] = append(outgoing[e.from], e.to)
			}

			// Find root nodes
			blocked := make(map[string]bool)
			for _, iss := range component {
				for _, blockerID := range iss.BlockedBy {
					if componentSet[blockerID] {
						blocked[iss.ID] = true
					}
				}
			}

			// Render starting from roots
			visited := make(map[string]bool)
			for _, iss := range component {
				if !blocked[iss.ID] && len(outgoing[iss.ID]) > 0 {
					renderGraphNode(iss.ID, issueMap, outgoing, renderer, visited, "")
				}
			}

			fmt.Println()
		}
	}

	return nil
}

// edge represents a dependency edge
type edge struct {
	from string
	to   string
}

// collectAllEdges collects all edges from the issue list
func collectAllEdges(issueList []issues.Issue, issueMap map[string]*issues.Issue) []edge {
	var edges []edge
	seen := make(map[string]bool)

	for _, issue := range issueList {
		for _, blockerID := range issue.BlockedBy {
			if _, exists := issueMap[blockerID]; exists {
				key := blockerID + "->" + issue.ID
				if !seen[key] {
					edges = append(edges, edge{from: blockerID, to: issue.ID})
					seen[key] = true
				}
			}
		}
	}

	return edges
}

// findConnectedComponents finds all connected components in the dependency graph
func findConnectedComponents(issueList []issues.Issue, issueMap map[string]*issues.Issue) [][]issues.Issue {
	// Build adjacency list (undirected for connected components)
	adj := make(map[string]map[string]bool)
	for _, issue := range issueList {
		adj[issue.ID] = make(map[string]bool)
	}

	for _, issue := range issueList {
		for _, blockerID := range issue.BlockedBy {
			if _, exists := issueMap[blockerID]; exists {
				adj[issue.ID][blockerID] = true
				adj[blockerID][issue.ID] = true
			}
		}
		for _, blocksID := range issue.Blocks {
			if _, exists := issueMap[blocksID]; exists {
				adj[issue.ID][blocksID] = true
				adj[blocksID][issue.ID] = true
			}
		}
	}

	// Find connected components using BFS
	visited := make(map[string]bool)
	var components [][]issues.Issue

	for _, issue := range issueList {
		if visited[issue.ID] {
			continue
		}

		// BFS to find all nodes in this component
		var component []issues.Issue
		queue := []string{issue.ID}
		visited[issue.ID] = true

		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]

			if iss, exists := issueMap[current]; exists {
				component = append(component, *iss)
			}

			for neighbor := range adj[current] {
				if !visited[neighbor] {
					visited[neighbor] = true
					queue = append(queue, neighbor)
				}
			}
		}

		// Only include components with edges (more than 1 node connected)
		if len(component) > 1 {
			// Verify there are actual edges
			hasEdges := false
			for _, iss := range component {
				for _, blockerID := range iss.BlockedBy {
					if _, exists := issueMap[blockerID]; exists {
						hasEdges = true
						break
					}
				}
				if hasEdges {
					break
				}
			}
			if hasEdges {
				components = append(components, component)
			}
		}
	}

	return components
}

// renderCrossSpecTree renders issues grouped by spec with hierarchy trees within each spec
func renderCrossSpecTree(issueList []issues.Issue, artifactPath string) error {
	// Group issues by spec
	specIssues := make(map[string][]issues.Issue)
	for _, issue := range issueList {
		specIssues[issue.SpecContext] = append(specIssues[issue.SpecContext], issue)
	}

	// Create renderer (no spec context since we're grouping by spec)
	opts := issues.DefaultTreeRenderOptions()
	opts.ShowSpec = false
	renderer := issues.NewTreeRenderer(opts)

	// Sort specs for consistent output
	specNames := make([]string, 0, len(specIssues))
	for spec := range specIssues {
		specNames = append(specNames, spec)
	}

	// Output
	fmt.Println("All Specs")

	for specIdx, spec := range specNames {
		issuesInSpec := specIssues[spec]
		isLastSpec := specIdx == len(specNames)-1

		// Spec header
		var specPrefix string
		if isLastSpec {
			specPrefix = "└── "
		} else {
			specPrefix = "├── "
		}
		fmt.Printf("%s%s (%d issues)\n", specPrefix, spec, len(issuesInSpec))

		// Create store for this spec to get hierarchy
		store, err := issues.NewStore(issues.StoreOptions{
			BasePath:    artifactPath,
			SpecContext: spec,
		})
		if err != nil {
			continue
		}

		// Build hierarchy trees
		trees, err := store.GetHierarchyForest()
		if err != nil {
			continue
		}

		// Render each hierarchy tree with spec prefix
		childPrefix := "    "
		if !isLastSpec {
			childPrefix = "│   "
		}

		for treeIdx, tree := range trees {
			isLastTree := treeIdx == len(trees)-1
			treeOutput := renderHierarchyTreeWithPrefix(renderer, tree, childPrefix, isLastTree)
			fmt.Print(treeOutput)
		}
	}

	return nil
}

// renderHierarchyTreeWithPrefix renders a hierarchy tree with a custom prefix
func renderHierarchyTreeWithPrefix(renderer *issues.TreeRenderer, tree *issues.DependencyTree, prefix string, isLast bool) string {
	var sb strings.Builder

	// Render current node
	if isLast {
		sb.WriteString(prefix + "└── ")
	} else {
		sb.WriteString(prefix + "├── ")
	}
	sb.WriteString(renderer.FormatIssueSimple(tree.Issue))
	sb.WriteString("\n")

	// Render children
	if len(tree.Children) > 0 {
		var childPrefix string
		if isLast {
			childPrefix = prefix + "    "
		} else {
			childPrefix = prefix + "│   "
		}

		for i, child := range tree.Children {
			childIsLast := i == len(tree.Children)-1
			sb.WriteString(renderHierarchyTreeWithPrefix(renderer, child, childPrefix, childIsLast))
		}
	}

	return sb.String()
}

// truncateTitle truncates a title to maxLen characters
func truncateTitle(title string, maxLen int) string {
	if len(title) <= maxLen {
		return title
	}
	if maxLen <= 3 {
		return title[:maxLen]
	}
	return title[:maxLen-3] + "..."
}

func runIssueReady(cmd *cobra.Command, args []string) error {
	// Determine spec context
	specContext := issueSpecFlag
	if specContext == "" && !issueAllFlag {
		detector := issues.NewContextDetector(".")
		var err error
		specContext, err = detector.DetectSpecContext()
		if err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	artifactPath := getArtifactPath()
	var readyIssues []issues.ReadyIssue
	var blockedIssues []issues.ReadyIssue
	var err error

	if issueAllFlag {
		// Get ready issues across all specs
		readyIssues, err = issues.ListReadyAcrossSpecs(artifactPath, issues.ListFilter{})
		if err != nil {
			return fmt.Errorf("failed to list ready issues: %w", err)
		}

		// Also get blocked issues for context
		// For cross-spec, we need to iterate through each spec
		specs, _ := issues.ListAllSpecs(artifactPath, issues.ListFilter{})
		specSet := make(map[string]bool)
		for _, issue := range specs {
			specSet[issue.SpecContext] = true
		}

		for spec := range specSet {
			store, storeErr := issues.NewStore(issues.StoreOptions{
				BasePath:    artifactPath,
				SpecContext: spec,
			})
			if storeErr != nil {
				continue
			}
			blocked, _ := store.GetBlockedIssuesWithBlockers()
			blockedIssues = append(blockedIssues, blocked...)
		}
	} else {
		store, storeErr := issues.NewStore(issues.StoreOptions{
			BasePath:    artifactPath,
			SpecContext: specContext,
		})
		if storeErr != nil {
			return fmt.Errorf("failed to create store: %w", storeErr)
		}

		readyIssues, err = store.ListReady(issues.ListFilter{})
		if err != nil {
			return fmt.Errorf("failed to list ready issues: %w", err)
		}

		blockedIssues, _ = store.GetBlockedIssuesWithBlockers()
	}

	// JSON output
	if issueJSONFlag {
		data, _ := json.MarshalIndent(readyIssues, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	// No ready issues found
	if len(readyIssues) == 0 {
		fmt.Println("No ready issues found.")
		if len(blockedIssues) > 0 {
			fmt.Println()
			fmt.Println("Blocked issues:")
			for _, bi := range blockedIssues {
				fmt.Printf("  %s \"%s\" is blocked by:\n", bi.Issue.ID, truncateTitle(bi.Issue.Title, 40))
				for _, blocker := range bi.BlockedBy {
					fmt.Printf("    - %s \"%s\" (%s)\n", blocker.ID, truncateTitle(blocker.Title, 30), blocker.Status)
				}
			}
		}
		return nil
	}

	// Table output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	if issueAllFlag {
		fmt.Fprintln(w, "ID\tTITLE\tSTATUS\tPRIORITY\tSPEC")
		for _, ri := range readyIssues {
			fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\n",
				ri.Issue.ID, truncateTitle(ri.Issue.Title, 40), ri.Issue.Status, ri.Issue.Priority, ri.Issue.SpecContext)
		}
	} else {
		fmt.Fprintln(w, "ID\tTITLE\tSTATUS\tPRIORITY")
		for _, ri := range readyIssues {
			fmt.Fprintf(w, "%s\t%s\t%s\t%d\n",
				ri.Issue.ID, truncateTitle(ri.Issue.Title, 40), ri.Issue.Status, ri.Issue.Priority)
		}
	}
	w.Flush()

	return nil
}

func runIssueShow(cmd *cobra.Command, args []string) error {
	issueID := args[0]

	// Validate ID format
	if _, err := issues.ParseIssueID(issueID); err != nil {
		return fmt.Errorf("invalid issue ID: %w", err)
	}

	// Detect spec context for searching
	detector := issues.NewContextDetector(".")
	specContext, err := detector.DetectSpecContext()
	if err != nil {
		specContext = ""
	}

	// Try to find the issue
	store, err := issues.NewStore(issues.StoreOptions{
		BasePath:    getArtifactPath(),
		SpecContext: specContext,
	})
	if err != nil {
		return fmt.Errorf("failed to create store: %w", err)
	}

	issue, err := store.Get(issueID)
	if err != nil {
		return fmt.Errorf("failed to get issue: %w", err)
	}

	if issueJSONFlag {
		data, _ := json.MarshalIndent(issue, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	// Handle tree view for single issue
	if issueTreeFlag {
		return renderIssueShowTree(store, issue)
	}

	// Print issue details
	fmt.Printf("Issue: %s\n", issue.ID)
	fmt.Printf("  Title: %s\n", issue.Title)
	fmt.Printf("  Type: %s\n", issue.IssueType)
	fmt.Printf("  Status: %s\n", issue.Status)
	fmt.Printf("  Priority: %d", issue.Priority)
	switch issue.Priority {
	case 0:
		fmt.Printf(" (critical)\n")
	case 1:
		fmt.Printf(" (high)\n")
	default:
		fmt.Println()
	}
	fmt.Printf("  Spec: %s\n", issue.SpecContext)
	if issue.ParentID != nil && *issue.ParentID != "" {
		fmt.Printf("  Parent: %s\n", *issue.ParentID)
	}
	fmt.Println()

	if issue.Description != "" {
		fmt.Println("Description:")
		fmt.Printf("  %s\n", strings.ReplaceAll(issue.Description, "\n", "\n  "))
		fmt.Println()
	}

	if issue.AcceptanceCriteria != "" {
		fmt.Println("Acceptance Criteria:")
		fmt.Printf("  %s\n", strings.ReplaceAll(issue.AcceptanceCriteria, "\n", "\n  "))
		fmt.Println()
	}

	if issue.Design != "" {
		fmt.Println("Design:")
		fmt.Printf("  %s\n", strings.ReplaceAll(issue.Design, "\n", "\n  "))
		fmt.Println()
	}

	if issue.DefinitionOfDone != nil && len(issue.DefinitionOfDone.Items) > 0 {
		fmt.Println("Definition of Done:")
		for _, item := range issue.DefinitionOfDone.Items {
			if item.Checked {
				if item.VerifiedAt != nil {
					fmt.Printf("  [x] %s (verified: %s)\n", item.Item, item.VerifiedAt.Format("2006-01-02 15:04:05"))
				} else {
					fmt.Printf("  [x] %s\n", item.Item)
				}
			} else {
				fmt.Printf("  [ ] %s\n", item.Item)
			}
		}
		fmt.Println()
	}

	if issue.Notes != "" {
		fmt.Println("Notes:")
		fmt.Printf("  %s\n", strings.ReplaceAll(issue.Notes, "\n", "\n  "))
		fmt.Println()
	}

	if len(issue.Labels) > 0 {
		fmt.Println("Labels:")
		for _, label := range issue.Labels {
			fmt.Printf("  - %s\n", label)
		}
		fmt.Println()
	}

	// Get and display children
	children, err := store.GetChildren(issue.ID)
	if err == nil && len(children) > 0 {
		fmt.Println("Children:")
		for _, child := range children {
			fmt.Printf("  - %s (%s) [P%d]\n", child.ID, child.IssueType, child.Priority)
		}
		fmt.Println()
	}

	fmt.Printf("Created: %s\n", issue.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Updated: %s\n", issue.UpdatedAt.Format("2006-01-02 15:04:05"))

	if issue.ClosedAt != nil {
		fmt.Printf("Closed: %s\n", issue.ClosedAt.Format("2006-01-02 15:04:05"))
	}

	return nil
}

// renderIssueShowTree renders a tree showing parent-child hierarchy and blocking relationships
func renderIssueShowTree(store *issues.Store, issue *issues.Issue) error {
	tree, err := store.GetDependencyTree(issue.ID)
	if err != nil {
		return fmt.Errorf("failed to get dependency tree: %w", err)
	}

	renderer := issues.NewTreeRenderer(issues.DefaultTreeRenderOptions())

	// Show parent hierarchy (if any)
	if issue.ParentID != nil && *issue.ParentID != "" {
		fmt.Println("Parent:")
		parent, err := store.Get(*issue.ParentID)
		if err == nil {
			fmt.Printf("└── %s (%s) [P%d]\n", parent.ID, parent.IssueType, parent.Priority)
		} else {
			fmt.Printf("└── %s (not found)\n", *issue.ParentID)
		}
		fmt.Println()
	}

	// Show what blocks this issue
	if len(tree.BlockedBy) > 0 {
		fmt.Println("Blocked by:")
		for i, blocker := range tree.BlockedBy {
			isLast := i == len(tree.BlockedBy)-1
			prefix := "├── "
			if isLast {
				prefix = "└── "
			}
			fmt.Printf("%s%s\n", prefix, renderer.FormatIssueSimple(blocker.Issue))
		}
		fmt.Println()
	}

	// Show the issue itself (centered)
	fmt.Printf("%s (%s) [P%d]\n", issue.ID, issue.IssueType, issue.Priority)
	fmt.Println()

	// Show children recursively (parent-child hierarchy)
	children, err := store.GetChildren(issue.ID)
	if err == nil && len(children) > 0 {
		fmt.Println("Children:")
		for i, child := range children {
			isLast := i == len(children)-1
			renderChildTree(store, child, "", isLast)
		}
		fmt.Println()
	}

	// Show what this issue blocks
	if len(tree.Blocks) > 0 {
		fmt.Println("Blocks:")
		for i, blocked := range tree.Blocks {
			isLast := i == len(tree.Blocks)-1
			prefix := "├── "
			if isLast {
				prefix = "└── "
			}
			fmt.Printf("%s%s\n", prefix, renderer.FormatIssueSimple(blocked.Issue))
		}
	}

	// No dependencies or children
	if len(tree.BlockedBy) == 0 && len(tree.Blocks) == 0 && len(children) == 0 && issue.ParentID == nil {
		fmt.Println("(No dependencies or hierarchy)")
	}

	return nil
}

// renderChildTree recursively renders children with proper tree characters
func renderChildTree(store *issues.Store, issue issues.Issue, prefix string, isLast bool) {
	// Determine the connector for this item
	connector := "├── "
	if isLast {
		connector = "└── "
	}

	// Print this item
	fmt.Printf("%s%s%s (%s) [P%d]\n", prefix, connector, issue.ID, issue.IssueType, issue.Priority)

	// Get children of this issue
	children, err := store.GetChildren(issue.ID)
	if err != nil || len(children) == 0 {
		return
	}

	// Determine the prefix for children
	childPrefix := prefix
	if isLast {
		childPrefix += "    "
	} else {
		childPrefix += "│   "
	}

	// Render children
	for i, child := range children {
		childIsLast := i == len(children)-1
		renderChildTree(store, child, childPrefix, childIsLast)
	}
}

func runIssueUpdate(cmd *cobra.Command, args []string) error {
	issueID := args[0]

	// Validate ID format
	if _, err := issues.ParseIssueID(issueID); err != nil {
		return fmt.Errorf("invalid issue ID: %w", err)
	}

	// Detect spec context
	detector := issues.NewContextDetector(".")
	specContext, err := detector.DetectSpecContext()
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	store, err := issues.NewStore(issues.StoreOptions{
		BasePath:    getArtifactPath(),
		SpecContext: specContext,
	})
	if err != nil {
		return fmt.Errorf("failed to create store: %w", err)
	}

	// Build update
	update := issues.IssueUpdate{}
	if cmd.Flags().Changed("title") {
		update.Title = &issueTitleFlag
	}
	if cmd.Flags().Changed("description") {
		update.Description = &issueDescFlag
	}
	if cmd.Flags().Changed("status") {
		status := issues.IssueStatus(issueStatusFlag)
		if !issues.IsValidStatus(status) {
			return fmt.Errorf("invalid status: %s", issueStatusFlag)
		}
		update.Status = &status
	}
	if cmd.Flags().Changed("priority") {
		update.Priority = &issuePriorityFlag
	}
	if cmd.Flags().Changed("assignee") {
		update.Assignee = &issueAssigneeFlag
	}
	if cmd.Flags().Changed("notes") {
		update.Notes = &issueNotesFlag
	}
	if cmd.Flags().Changed("design") {
		update.Design = &issueDesignFlag
	}
	if cmd.Flags().Changed("acceptance-criteria") {
		update.AcceptanceCriteria = &issueAcceptFlag
	}
	if issueAddLabelFlag != "" {
		update.AddLabels = []string{issueAddLabelFlag}
	}
	if issueRemoveLabelFlag != "" {
		update.RemoveLabels = []string{issueRemoveLabelFlag}
	}

	// Handle DoD operations
	if cmd.Flags().Changed("dod") {
		items := make([]issues.ChecklistItem, len(issueDoDFlag))
		for i, item := range issueDoDFlag {
			items[i] = issues.ChecklistItem{
				Item:    item,
				Checked: false,
			}
		}
		update.DefinitionOfDone = &issues.DefinitionOfDone{Items: items}
	}
	if issueCheckDoDFlag != "" {
		update.CheckDoDItem = issueCheckDoDFlag
	}
	if issueUncheckDoDFlag != "" {
		update.UncheckDoDItem = issueUncheckDoDFlag
	}

	// Handle parent update
	if cmd.Flags().Changed("parent") {
		if issueParentFlag == "" {
			// Clear parent
			update.ParentID = &issueParentFlag // empty string clears parent
		} else {
			update.ParentID = &issueParentFlag
		}
	}

	issue, err := store.Update(issueID, update)
	if err != nil {
		return fmt.Errorf("failed to update issue: %w", err)
	}

	fmt.Printf("%s Updated issue %s\n", ui.Checkmark(), issue.ID)
	return nil
}

func runIssueClose(cmd *cobra.Command, args []string) error {
	issueID := args[0]

	// Validate ID format
	if _, err := issues.ParseIssueID(issueID); err != nil {
		return fmt.Errorf("invalid issue ID: %w", err)
	}

	// Detect spec context
	detector := issues.NewContextDetector(".")
	specContext, err := detector.DetectSpecContext()
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	store, err := issues.NewStore(issues.StoreOptions{
		BasePath:    getArtifactPath(),
		SpecContext: specContext,
	})
	if err != nil {
		return fmt.Errorf("failed to create store: %w", err)
	}

	// Get issue to check DoD
	issue, err := store.Get(issueID)
	if err != nil {
		return fmt.Errorf("failed to get issue: %w", err)
	}

	// Check definition of done
	if issue.DefinitionOfDone != nil && !issue.DefinitionOfDone.IsComplete() && !issueForceFlag {
		unchecked := issue.DefinitionOfDone.GetUncheckedItems()
		fmt.Println("Definition of done not met:")
		for _, item := range unchecked {
			fmt.Printf("  [ ] %s\n", item)
		}
		fmt.Println()
		return fmt.Errorf("use --force to close anyway")
	}

	// Close the issue
	status := issues.StatusClosed
	update := issues.IssueUpdate{
		Status: &status,
	}

	_, err = store.Update(issueID, update)
	if err != nil {
		return fmt.Errorf("failed to close issue: %w", err)
	}

	fmt.Printf("%s Closed issue %s\n", ui.Checkmark(), issueID)
	if issueReasonFlag != "" {
		fmt.Printf("  Reason: %s\n", issueReasonFlag)
	}
	return nil
}

func runIssueMigrate(cmd *cobra.Command, args []string) error {
	ui.PrintSection("Migrating Beads Issues")

	// Create migrator
	migrator := issues.NewMigrator(issues.MigratorOptions{
		DryRun:    issueDryRunFlag,
		KeepBeads: issueKeepBeadsFlag,
	})

	// Perform migration
	result, err := migrator.Migrate()
	if err != nil {
		if errors.Is(err, issues.ErrBeadsNotFound) {
			fmt.Printf("%s No .beads/issues.jsonl found - nothing to migrate\n", ui.Checkmark())
			return nil
		}
		return fmt.Errorf("migration failed: %w", err)
	}

	// Print results
	fmt.Println()
	fmt.Printf("Total issues found: %d\n", result.TotalIssues)
	fmt.Println()

	// Print distribution
	fmt.Println("Issues by spec:")
	for spec, count := range result.SpecDistribution {
		fmt.Printf("  %s: %d issues\n", spec, count)
	}

	if len(result.UnmappedIssues) > 0 {
		fmt.Println()
		fmt.Printf("%s %d issues could not be mapped to a spec\n",
			ui.WarningIcon(), len(result.UnmappedIssues))
		fmt.Println("  These will be placed in specledger/migrated/")
	}

	if len(result.Errors) > 0 {
		fmt.Println()
		fmt.Printf("❌ %d errors occurred during migration\n", len(result.Errors))
		for _, e := range result.Errors {
			fmt.Printf("  - %v\n", e)
		}
	}

	if len(result.Warnings) > 0 {
		fmt.Println()
		fmt.Printf("%s %d warnings during migration\n", ui.WarningIcon(), len(result.Warnings))
		for _, w := range result.Warnings {
			fmt.Printf("  - %s\n", w)
		}
	}

	if issueDryRunFlag {
		fmt.Println()
		fmt.Println("Dry run complete. No changes were made.")
		fmt.Println("Run without --dry-run to perform actual migration.")
		return nil
	}

	fmt.Println()
	fmt.Printf("%s Migration complete!\n", ui.Checkmark())
	fmt.Printf("  %d issues migrated\n", result.MigratedIssues)

	if !issueKeepBeadsFlag && result.MigratedIssues > 0 {
		// Check if cleanup actually succeeded by checking if .beads still exists
		if _, err := os.Stat(".beads"); os.IsNotExist(err) {
			fmt.Println("  .beads/ directory removed")
			fmt.Println("  mise.toml updated")
			fmt.Println("  Migration log written to specledger/.migration-log")
		} else {
			fmt.Printf("  %s .beads/ directory may still exist (check warnings above)\n", ui.WarningIcon())
		}
	}

	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  sl issue list --all")
	fmt.Println("  sl issue show <issue-id>")

	return nil
}

func runIssueLink(cmd *cobra.Command, args []string) error {
	fromID := args[0]
	linkTypeStr := args[1]
	toID := args[2]

	// Validate IDs
	if _, err := issues.ParseIssueID(fromID); err != nil {
		return fmt.Errorf("invalid from issue ID: %w", err)
	}
	if _, err := issues.ParseIssueID(toID); err != nil {
		return fmt.Errorf("invalid to issue ID: %w", err)
	}

	// Validate link type
	linkType := issues.LinkType(linkTypeStr)
	if !issues.IsValidLinkType(linkType) {
		return fmt.Errorf("invalid link type: %s (must be 'blocks' or 'related')", linkTypeStr)
	}

	// Detect spec context
	detector := issues.NewContextDetector(".")
	specContext, err := detector.DetectSpecContext()
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	store, err := issues.NewStore(issues.StoreOptions{
		BasePath:    getArtifactPath(),
		SpecContext: specContext,
	})
	if err != nil {
		return fmt.Errorf("failed to create store: %w", err)
	}

	if err := store.AddDependency(fromID, toID, linkType); err != nil {
		if errors.Is(err, issues.ErrCyclicDependency) {
			return fmt.Errorf("cannot create dependency: %w", err)
		}
		return fmt.Errorf("failed to create dependency: %w", err)
	}

	fmt.Printf("%s Created dependency: %s %s %s\n", ui.Checkmark(), fromID, linkTypeStr, toID)
	return nil
}

func runIssueUnlink(cmd *cobra.Command, args []string) error {
	fromID := args[0]
	linkTypeStr := args[1]
	toID := args[2]

	// Validate IDs
	if _, err := issues.ParseIssueID(fromID); err != nil {
		return fmt.Errorf("invalid from issue ID: %w", err)
	}
	if _, err := issues.ParseIssueID(toID); err != nil {
		return fmt.Errorf("invalid to issue ID: %w", err)
	}

	// Validate link type
	linkType := issues.LinkType(linkTypeStr)
	if !issues.IsValidLinkType(linkType) {
		return fmt.Errorf("invalid link type: %s (must be 'blocks' or 'related')", linkTypeStr)
	}

	// Detect spec context
	detector := issues.NewContextDetector(".")
	specContext, err := detector.DetectSpecContext()
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	store, err := issues.NewStore(issues.StoreOptions{
		BasePath:    getArtifactPath(),
		SpecContext: specContext,
	})
	if err != nil {
		return fmt.Errorf("failed to create store: %w", err)
	}

	if err := store.RemoveDependency(fromID, toID, linkType); err != nil {
		return fmt.Errorf("failed to remove dependency: %w", err)
	}

	fmt.Printf("%s Removed dependency: %s %s %s\n", ui.Checkmark(), fromID, linkTypeStr, toID)
	return nil
}

func runIssueRepair(cmd *cobra.Command, args []string) error {
	// Detect spec context
	detector := issues.NewContextDetector(".")
	specContext, err := detector.DetectSpecContext()
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	ui.PrintSection("Repairing Issues File")

	store, err := issues.NewStore(issues.StoreOptions{
		BasePath:    getArtifactPath(),
		SpecContext: specContext,
	})
	if err != nil {
		return fmt.Errorf("failed to create store: %w", err)
	}

	// Repair the issues file
	result, err := issues.RepairIssuesFile(store.Path())
	if err != nil {
		return fmt.Errorf("repair failed: %w", err)
	}

	fmt.Printf("Spec: %s\n", specContext)
	fmt.Printf("  Valid lines: %d\n", result.ValidLines)
	fmt.Printf("  Invalid lines: %d\n", result.InvalidLines)
	fmt.Printf("  Recovered issues: %d\n", result.RecoveredIssues)

	if result.InvalidLines > 0 {
		fmt.Println()
		fmt.Println("Invalid lines skipped:")
		for _, line := range result.SkippedLines {
			fmt.Printf("  Line %d: %s\n", line.LineNum, line.Reason)
		}
	}

	if result.BackupPath != "" {
		fmt.Println()
		fmt.Printf("Backup saved to: %s\n", result.BackupPath)
	}

	fmt.Printf("\n%s Repair complete\n", ui.Checkmark())
	return nil
}
