package issues

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Migration-related errors
var (
	ErrBeadsNotFound     = errors.New(".beads/issues.jsonl not found")
	ErrMigrationFailed   = errors.New("migration failed")
	ErrNoIssuesToMigrate = errors.New("no issues to migrate")
)

// BeadsIssue represents the Beads JSONL format for issues
type BeadsIssue struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Status      string   `json:"status"`
	Priority    int      `json:"priority"`
	Type        string   `json:"type"`
	Labels      []string `json:"labels,omitempty"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
	ClosedAt    string   `json:"closed_at,omitempty"`
	Notes       string   `json:"notes,omitempty"`
	Design      string   `json:"design,omitempty"`
	Acceptance  string   `json:"acceptance_criteria,omitempty"`
	BlockedBy   []string `json:"blocked_by,omitempty"`
	Blocks      []string `json:"blocks,omitempty"`
	Assignee    string   `json:"assignee,omitempty"`
}

// MigrationResult contains the results of a migration
type MigrationResult struct {
	TotalIssues      int
	MigratedIssues   int
	SkippedIssues    int
	SpecDistribution map[string]int // spec -> count
	UnmappedIssues   []BeadsIssue
	Errors           []error
	Warnings         []string
}

// Migrator handles migration from Beads to sl issue format
type Migrator struct {
	beadsPath    string
	artifactPath string
	dryRun       bool
	keepBeads    bool
	idMapping    map[string]string // old Beads ID -> new SL ID
}

// MigratorOptions contains options for the migrator
type MigratorOptions struct {
	BeadsPath    string // Path to .beads directory (default: ".beads")
	ArtifactPath string // Path to specledger directory (default: "specledger")
	DryRun       bool
	KeepBeads    bool
}

// NewMigrator creates a new migrator
func NewMigrator(opts MigratorOptions) *Migrator {
	beadsPath := opts.BeadsPath
	if beadsPath == "" {
		beadsPath = ".beads"
	}
	artifactPath := opts.ArtifactPath
	if artifactPath == "" {
		artifactPath = "specledger"
	}

	return &Migrator{
		beadsPath:    beadsPath,
		artifactPath: artifactPath,
		dryRun:       opts.DryRun,
		keepBeads:    opts.KeepBeads,
		idMapping:    make(map[string]string),
	}
}

// Migrate performs the full migration from Beads to sl issue format
func (m *Migrator) Migrate() (*MigrationResult, error) {
	result := &MigrationResult{
		SpecDistribution: make(map[string]int),
	}

	// Check if Beads file exists
	beadsFile := filepath.Join(m.beadsPath, "issues.jsonl")
	if _, err := os.Stat(beadsFile); os.IsNotExist(err) {
		return nil, ErrBeadsNotFound
	}

	// Parse Beads issues
	beadsIssues, err := m.parseBeadsFile(beadsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Beads file: %w", err)
	}

	result.TotalIssues = len(beadsIssues)
	if result.TotalIssues == 0 {
		return nil, ErrNoIssuesToMigrate
	}

	// Group issues by spec context
	issuesBySpec := make(map[string][]BeadsIssue)
	var unmapped []BeadsIssue

	for _, beadsIssue := range beadsIssues {
		specContext := m.extractSpecContext(beadsIssue)
		if specContext == "" {
			specContext = "migrated"
			unmapped = append(unmapped, beadsIssue)
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("Issue %s could not be mapped to a spec", beadsIssue.ID))
		}
		issuesBySpec[specContext] = append(issuesBySpec[specContext], beadsIssue)
	}

	// Convert and write issues to each spec
	for specContext, issues := range issuesBySpec {
		if m.dryRun {
			fmt.Printf("Would migrate %d issues to %s\n", len(issues), specContext)
			result.SpecDistribution[specContext] = len(issues)
			result.MigratedIssues += len(issues)
			continue
		}

		converted, err := m.convertAndWriteIssues(specContext, issues)
		if err != nil {
			result.Errors = append(result.Errors, err)
			result.SkippedIssues += len(issues)
			continue
		}

		result.SpecDistribution[specContext] = converted
		result.MigratedIssues += converted
	}

	result.UnmappedIssues = unmapped

	// Perform cleanup if not dry run and not keeping beads
	if !m.dryRun && !m.keepBeads && result.MigratedIssues > 0 {
		if err := m.Cleanup(); err != nil {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("Migration succeeded but cleanup failed: %v", err))
		}
	}

	return result, nil
}

// parseBeadsFile reads and parses the Beads JSONL file
func (m *Migrator) parseBeadsFile(path string) ([]BeadsIssue, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var issues []BeadsIssue
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var issue BeadsIssue
		if err := json.Unmarshal([]byte(line), &issue); err != nil {
			// Skip invalid lines
			continue
		}

		issues = append(issues, issue)
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		return nil, err
	}

	return issues, nil
}

// extractSpecContext extracts the spec context from a Beads issue
func (m *Migrator) extractSpecContext(issue BeadsIssue) string {
	// Try to extract from labels
	for _, label := range issue.Labels {
		// Look for pattern "spec:###-name"
		if strings.HasPrefix(label, "spec:") {
			specContext := strings.TrimPrefix(label, "spec:")
			if specBranchPattern.MatchString(specContext) {
				return specContext
			}
		}
	}

	// Try to extract from description (branch references)
	// Look for patterns like "branch: 010-my-feature" or "on branch 010-my-feature"
	desc := issue.Description + " " + issue.Notes
	if matches := specBranchPattern.FindAllString(desc, -1); len(matches) > 0 {
		return matches[0]
	}

	// Try to extract from title (some users put spec in title)
	if matches := specBranchPattern.FindAllString(issue.Title, -1); len(matches) > 0 {
		return matches[0]
	}

	return ""
}

// convertAndWriteIssues converts Beads issues to sl format and writes them
func (m *Migrator) convertAndWriteIssues(specContext string, beadsIssues []BeadsIssue) (int, error) {
	// Ensure directory exists
	specDir := filepath.Join(m.artifactPath, specContext)
	if err := os.MkdirAll(specDir, 0755); err != nil {
		return 0, fmt.Errorf("failed to create spec directory: %w", err)
	}

	// Create store
	store, err := NewStore(StoreOptions{
		BasePath:    m.artifactPath,
		SpecContext: specContext,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to create store: %w", err)
	}

	converted := 0
	now := time.Now()

	for _, beadsIssue := range beadsIssues {
		issue := m.convertBeadsIssue(beadsIssue, specContext, now)

		if err := store.Create(issue); err != nil {
			// Skip if already exists or other error
			continue
		}

		// Store ID mapping for dependency resolution
		m.idMapping[beadsIssue.ID] = issue.ID
		converted++
	}

	return converted, nil
}

// convertBeadsIssue converts a Beads issue to sl issue format
func (m *Migrator) convertBeadsIssue(beadsIssue BeadsIssue, specContext string, now time.Time) *Issue {
	issue := &Issue{
		Title:              beadsIssue.Title,
		Description:        beadsIssue.Description,
		Status:             m.convertStatus(beadsIssue.Status),
		Priority:           beadsIssue.Priority,
		IssueType:          m.convertType(beadsIssue.Type),
		SpecContext:        specContext,
		Labels:             beadsIssue.Labels,
		Assignee:           beadsIssue.Assignee,
		Notes:              beadsIssue.Notes,
		Design:             beadsIssue.Design,
		AcceptanceCriteria: beadsIssue.Acceptance,
	}

	// Parse timestamps
	if beadsIssue.CreatedAt != "" {
		if t, err := time.Parse(time.RFC3339, beadsIssue.CreatedAt); err == nil {
			issue.CreatedAt = t
		} else {
			issue.CreatedAt = now
		}
	} else {
		issue.CreatedAt = now
	}

	if beadsIssue.UpdatedAt != "" {
		if t, err := time.Parse(time.RFC3339, beadsIssue.UpdatedAt); err == nil {
			issue.UpdatedAt = t
		} else {
			issue.UpdatedAt = now
		}
	} else {
		issue.UpdatedAt = now
	}

	if beadsIssue.ClosedAt != "" {
		if t, err := time.Parse(time.RFC3339, beadsIssue.ClosedAt); err == nil {
			issue.ClosedAt = &t
		}
	}

	// Generate new ID (use beads creation time for deterministic ID)
	issue.ID = GenerateIssueID(specContext, issue.Title, issue.CreatedAt)

	// Add migration metadata
	issue.BeadsMigration = &BeadsMigration{
		OriginalID: beadsIssue.ID,
		MigratedAt: now,
	}

	return issue
}

// convertStatus converts Beads status to sl status
func (m *Migrator) convertStatus(status string) IssueStatus {
	switch strings.ToLower(status) {
	case "open":
		return StatusOpen
	case "in_progress", "in-progress", "inprogress":
		return StatusInProgress
	case "closed", "done", "complete":
		return StatusClosed
	default:
		return StatusOpen
	}
}

// convertType converts Beads type to sl type
func (m *Migrator) convertType(issueType string) IssueType {
	switch strings.ToLower(issueType) {
	case "epic":
		return TypeEpic
	case "feature":
		return TypeFeature
	case "bug":
		return TypeBug
	case "task", "chore", "":
		return TypeTask
	default:
		return TypeTask
	}
}

// Cleanup removes Beads dependencies after successful migration
func (m *Migrator) Cleanup() error {
	if m.keepBeads {
		return nil
	}

	// 1. Remove .beads directory
	if err := os.RemoveAll(m.beadsPath); err != nil {
		return fmt.Errorf("failed to remove .beads directory: %w", err)
	}

	// 2. Update mise.toml
	if err := m.removeFromMiseToml(); err != nil {
		return fmt.Errorf("failed to update mise.toml: %w", err)
	}

	// 3. Write migration log
	if err := m.writeMigrationLog(); err != nil {
		return fmt.Errorf("failed to write migration log: %w", err)
	}

	return nil
}

// removeFromMiseToml removes beads and perles from mise.toml
func (m *Migrator) removeFromMiseToml() error {
	misePath := "mise.toml"
	if _, err := os.Stat(misePath); os.IsNotExist(err) {
		return nil // No mise.toml, nothing to do
	}

	content, err := os.ReadFile(misePath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string
	inToolsSection := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Track section
		if strings.HasPrefix(trimmed, "[tools]") {
			inToolsSection = true
			newLines = append(newLines, line)
			continue
		}
		if strings.HasPrefix(trimmed, "[") && !strings.HasPrefix(trimmed, "[tools") {
			inToolsSection = false
		}

		// Skip beads and perles lines in tools section
		if inToolsSection {
			if strings.Contains(line, "beads") || strings.Contains(line, "perles") {
				continue
			}
		}

		newLines = append(newLines, line)
	}

	// #nosec G306 -- mise.toml needs to be readable by mise tool
	return os.WriteFile(misePath, []byte(strings.Join(newLines, "\n")), 0644)
}

// writeMigrationLog creates a log file of the migration
func (m *Migrator) writeMigrationLog() error {
	logPath := filepath.Join(m.artifactPath, ".migration-log")

	var sb strings.Builder
	fmt.Fprintf(&sb, "# Migration Log\n")
	fmt.Fprintf(&sb, "# Date: %s\n\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(&sb, "Total issues migrated: %d\n", len(m.idMapping))
	fmt.Fprintf(&sb, "\n## ID Mapping (Beads -> SL)\n\n")

	for oldID, newID := range m.idMapping {
		fmt.Fprintf(&sb, "%s -> %s\n", oldID, newID)
	}

	// #nosec G306 -- migration log needs to be readable by user
	return os.WriteFile(logPath, []byte(sb.String()), 0644)
}

// GetIDMapping returns the ID mapping from old Beads IDs to new SL IDs
func (m *Migrator) GetIDMapping() map[string]string {
	return m.idMapping
}
