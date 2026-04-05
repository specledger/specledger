package commands

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/specledger/specledger/internal/agent"
	"github.com/specledger/specledger/pkg/cli/metadata"
	"github.com/specledger/specledger/pkg/cli/skills"
	"github.com/specledger/specledger/pkg/version"
	"github.com/spf13/cobra"
)

// Flag variables
var (
	skillSearchJSON  bool
	skillSearchLimit int
	skillAddJSON     bool
	skillAddYes      bool
	skillInfoJSON    bool
	skillListJSON    bool
	skillRemoveJSON  bool
	skillAuditJSON bool
	skillAuditAll  bool
)

// VarSkillCmd represents the skill command
var VarSkillCmd = &cobra.Command{
	Use:   "skill",
	Short: "Manage agent skills from the skills.sh registry",
	Long: `Search, install, and manage agent skills from the skills.sh registry.

Subcommands:
  search   Search for skills by keyword
  add      Install a skill from a repository
  info     Show skill details and security audit
  list     List installed skills
  remove   Remove an installed skill
  audit    Run security audit on installed skills

Examples:
  sl skill search "commit"
  sl skill add vercel-labs/agent-skills@creating-pr
  sl skill list
  sl skill audit`,
	Args:         cobra.NoArgs,
	RunE:         func(cmd *cobra.Command, args []string) error { return cmd.Help() },
	SilenceUsage: true,
}

var skillSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for skills by keyword",
	Long: `Search the skills.sh registry for skills matching the query.

Output formats:
  Default: Compact table with install counts
  --json:  Full JSON array with all fields

Examples:
  sl skill search "commit"
  sl skill search "deploy" --limit 5
  sl skill search "testing" --json`,
	Args:         cobra.ExactArgs(1),
	RunE:         runSkillSearch,
	SilenceUsage: true,
}

var skillAddCmd = &cobra.Command{
	Use:   "add <source>",
	Short: "Install a skill from a repository",
	Long: `Install one or more skills from a GitHub repository or git URL.

Source formats:
  owner/repo              Install all skills from repository
  owner/repo@skill-name   Install a specific skill
  https://github.com/...  Full git URL

Examples:
  sl skill add vercel-labs/agent-skills@creating-pr
  sl skill add vercel-labs/agent-skills -y
  sl skill add https://github.com/org/repo.git`,
	Args:         cobra.ExactArgs(1),
	RunE:         runSkillAdd,
	SilenceUsage: true,
}

var skillInfoCmd = &cobra.Command{
	Use:   "info <source>",
	Short: "Show skill details and security audit",
	Long: `Display metadata and security risk assessments for a skill.

Examples:
  sl skill info vercel-labs/agent-skills@creating-pr
  sl skill info vercel-labs/agent-skills@creating-pr --json`,
	Args:         cobra.ExactArgs(1),
	RunE:         runSkillInfo,
	SilenceUsage: true,
}

var skillListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed skills",
	Long: `List all skills tracked in skills-lock.json.

Output formats:
  Default: Compact list with source info
  --json:  Full JSON array with all lock file fields

Examples:
  sl skill list
  sl skill list --json`,
	Args:         cobra.NoArgs,
	RunE:         runSkillList,
	SilenceUsage: true,
}

var skillRemoveCmd = &cobra.Command{
	Use:   "remove <skill-name>",
	Short: "Remove an installed skill",
	Long: `Remove a skill from agent directories and skills-lock.json.

Examples:
  sl skill remove creating-pr
  sl skill remove creating-pr --json`,
	Args:         cobra.ExactArgs(1),
	RunE:         runSkillRemove,
	SilenceUsage: true,
}

var skillAuditCmd = &cobra.Command{
	Use:   "audit [skill-name]",
	Short: "Run security audit on installed skills",
	Long: `Fetch security risk assessments for installed skills from ATH, Socket, and Snyk.

If a skill name is given, audits only that skill. Otherwise, audits all installed skills.

Examples:
  sl skill audit
  sl skill audit creating-pr
  sl skill audit --json`,
	Args:         cobra.MaximumNArgs(1),
	RunE:         runSkillAudit,
	SilenceUsage: true,
}

func init() {
	VarSkillCmd.AddCommand(skillSearchCmd)
	VarSkillCmd.AddCommand(skillAddCmd)
	VarSkillCmd.AddCommand(skillInfoCmd)
	VarSkillCmd.AddCommand(skillListCmd)
	VarSkillCmd.AddCommand(skillRemoveCmd)
	VarSkillCmd.AddCommand(skillAuditCmd)

	skillSearchCmd.Flags().BoolVar(&skillSearchJSON, "json", false, "Output as JSON array")
	skillSearchCmd.Flags().IntVar(&skillSearchLimit, "limit", 10, "Maximum results to return")

	skillAddCmd.Flags().BoolVar(&skillAddJSON, "json", false, "Output as JSON")
	skillAddCmd.Flags().BoolVarP(&skillAddYes, "yes", "y", false, "Skip confirmation prompt")

	skillInfoCmd.Flags().BoolVar(&skillInfoJSON, "json", false, "Output as JSON")

	skillListCmd.Flags().BoolVar(&skillListJSON, "json", false, "Output as JSON array")

	skillRemoveCmd.Flags().BoolVar(&skillRemoveJSON, "json", false, "Output as JSON")

	skillAuditCmd.Flags().BoolVar(&skillAuditJSON, "json", false, "Output as JSON")
	skillAuditCmd.Flags().BoolVar(&skillAuditAll, "all", false, "Audit all installed skills (default behavior)")
}

// --- Search ---

func runSkillSearch(_ *cobra.Command, args []string) error {
	query := args[0]
	client := skills.NewClient()

	results, err := client.Search(query, skillSearchLimit)
	if err != nil {
		return err
	}

	if skillSearchJSON {
		if results == nil {
			results = []skills.SkillSearchResult{}
		}
		return json.NewEncoder(os.Stdout).Encode(results)
	}

	if len(results) == 0 {
		fmt.Fprintf(os.Stderr, "No skills found for %q\n", query)
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for _, r := range results {
		fmt.Fprintf(w, "%s\t%s\t%s\n", r.Name, r.Source, formatInstalls(r.Installs))
	}
	w.Flush()
	fmt.Fprintf(os.Stderr, "→ sl skill add <owner/repo@skill> to install\n")

	return nil
}

// --- Add ---

func runSkillAdd(_ *cobra.Command, args []string) error {
	source, err := skills.ParseSource(args[0])
	if err != nil {
		return err
	}

	projectRoot, err := metadata.FindProjectRoot()
	if err != nil {
		return fmt.Errorf("not in a SpecLedger project\n→ Run 'sl init' first")
	}

	lockPath := filepath.Join(projectRoot, "skills-lock.json")
	client := skills.NewClient()

	// Discover skills
	discovered, err := skills.DiscoverSkills(client, source)
	if err != nil {
		return err
	}

	// Filter out internal skills unless INSTALL_INTERNAL_SKILLS=1
	var toInstall []skills.SkillMetadata
	for _, s := range discovered {
		if s.Internal && os.Getenv("INSTALL_INTERNAL_SKILLS") != "1" {
			continue
		}
		toInstall = append(toInstall, s)
	}

	if len(toInstall) == 0 {
		return fmt.Errorf("no installable skills found in %s", source.SourceString())
	}

	// Fetch audit data (non-blocking, 3s timeout)
	slugs := make([]string, len(toInstall))
	for i, s := range toInstall {
		slugs[i] = s.Name
	}

	auditCh := make(chan map[string]*skills.SkillAuditResult, 1)
	go func() {
		auditClient := skills.NewClient()
		auditClient.HTTPClient.Timeout = 3 * time.Second
		result, _ := auditClient.FetchAudit(source.SourceString(), slugs)
		auditCh <- result
	}()

	// Wait for audit with timeout
	var auditResults map[string]*skills.SkillAuditResult
	select {
	case auditResults = <-auditCh:
	case <-time.After(4 * time.Second):
		// Proceed without audit
	}

	if !skillAddJSON {
		// Display audit table
		if len(auditResults) > 0 {
			fmt.Println("\nSecurity Risk Assessments")
			printAuditTable(auditResults)
			fmt.Println()
		}

		// Check overwrite — filter out declined skills
		var confirmed []skills.SkillMetadata
		for _, s := range toInstall {
			if skills.IsSkillInstalled(s.Name, lockPath) {
				if !skillAddYes {
					fmt.Printf("%s is already installed. Overwrite? [y/N] ", s.Name)
					if !confirm() {
						fmt.Fprintf(os.Stderr, "Skipped %s\n", s.Name)
						continue
					}
				}
			}
			confirmed = append(confirmed, s)
		}
		toInstall = confirmed

		if len(toInstall) == 0 {
			return nil
		}

		// Confirmation prompt
		if !skillAddYes {
			names := make([]string, len(toInstall))
			for i, s := range toInstall {
				names[i] = s.Name
			}
			fmt.Printf("Install %s from %s? [Y/n] ", strings.Join(names, ", "), source.SourceString())
			if !confirmDefault(true) {
				return nil
			}
		}
	}

	// Resolve agent paths
	agentPaths, err := resolveAgentSkillPaths(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to resolve agent paths: %w", err)
	}

	// Install each skill
	type installResult struct {
		Name   string `json:"name"`
		Path   string `json:"path"`
		Source string `json:"source"`
	}
	var results []installResult

	for _, s := range toInstall {
		// Use the discovered repo path if available, fallback to conventional layout
		fetchPath := s.RepoPath
		if fetchPath == "" {
			fetchPath = fmt.Sprintf("skills/%s/SKILL.md", s.Name)
		}
		content, err := client.FetchSkillContent(source.Owner, source.Repo, source.Ref, fetchPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to fetch %s: %v\n", s.Name, err)
			continue
		}

		if err := skills.InstallSkill(s.Name, content, agentPaths, lockPath, source); err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to install %s: %v\n", s.Name, err)
			continue
		}

		skillPath := filepath.Join(agentPaths[0], s.Name, "SKILL.md")
		results = append(results, installResult{
			Name:   s.Name,
			Path:   skillPath,
			Source: source.SourceString(),
		})

		if !skillAddJSON {
			fmt.Printf("✓ Installed %s to %s\n", s.Name, skillPath)
		}
	}

	if !skillAddJSON {
		fmt.Println("✓ Updated skills-lock.json")
	}

	// Telemetry
	skillNames := make([]string, len(results))
	for i, r := range results {
		skillNames[i] = r.Name
	}
	skills.Track(client.AuditURL, "install",
		skills.BuildTelemetryParams(source.SourceString(), skillNames, nil),
		version.GetVersion(),
	)

	if skillAddJSON {
		return json.NewEncoder(os.Stdout).Encode(results)
	}

	return nil
}

// --- Info ---

func runSkillInfo(_ *cobra.Command, args []string) error {
	source, err := skills.ParseSource(args[0])
	if err != nil {
		return err
	}

	if source.SkillFilter == "" {
		return fmt.Errorf("skill name required\n→ Use format: owner/repo@skill-name")
	}

	client := skills.NewClient()

	// Discover to get metadata
	discovered, err := skills.DiscoverSkills(client, source)
	if err != nil {
		return err
	}

	if len(discovered) == 0 {
		return fmt.Errorf("skill %q not found", source.SkillFilter)
	}

	meta := discovered[0]

	// Fetch audit
	auditResults, _ := client.FetchAudit(source.SourceString(), []string{meta.Name})

	if skillInfoJSON {
		output := struct {
			skills.SkillMetadata
			Audit *skills.SkillAuditResult `json:"audit,omitempty"`
		}{
			SkillMetadata: meta,
		}
		if ar, ok := auditResults[meta.Name]; ok {
			output.Audit = ar
		}
		return json.NewEncoder(os.Stdout).Encode(output)
	}

	fmt.Printf("%s (%s)\n", meta.Name, meta.Source)
	fmt.Printf("Description: %s\n", meta.Description)

	if len(auditResults) > 0 {
		fmt.Println("\nSecurity Risk Assessments")
		printAuditSingle(meta.Name, auditResults)

		if ar, ok := auditResults[meta.Name]; ok && isHighOrCritical(ar) {
			fmt.Fprintf(os.Stderr, "\n⚠ Warning: %s has HIGH or CRITICAL risk. Review before using.\n", meta.Name)
			fmt.Fprintf(os.Stderr, "→ https://skills.sh/%s\n", meta.Slug)
		}
	}

	return nil
}

// --- List ---

func runSkillList(_ *cobra.Command, _ []string) error {
	projectRoot, err := metadata.FindProjectRoot()
	if err != nil {
		return fmt.Errorf("not in a SpecLedger project\n→ Run 'sl init' first")
	}

	lockPath := filepath.Join(projectRoot, "skills-lock.json")
	lock, err := skills.ReadLocalLock(lockPath)
	if err != nil {
		return err
	}

	if skillListJSON {
		type jsonEntry struct {
			Name         string `json:"name"`
			Source       string `json:"source"`
			SourceType   string `json:"sourceType"`
			ComputedHash string `json:"computedHash"`
		}
		entries := make([]jsonEntry, 0, len(lock.Skills))
		for name, entry := range lock.Skills {
			entries = append(entries, jsonEntry{
				Name:         name,
				Source:       entry.Source,
				SourceType:   entry.SourceType,
				ComputedHash: entry.ComputedHash,
			})
		}
		sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })
		return json.NewEncoder(os.Stdout).Encode(entries)
	}

	if len(lock.Skills) == 0 {
		fmt.Println("No skills installed.")
		fmt.Fprintf(os.Stderr, "→ Use 'sl skill search' to discover skills.\n")
		return nil
	}

	// Sort by name
	names := make([]string, 0, len(lock.Skills))
	for name := range lock.Skills {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		entry := lock.Skills[name]
		fmt.Printf("%-30s %s\n", name, entry.Source)
	}
	fmt.Fprintf(os.Stderr, "→ %d skill(s) installed. Use 'sl skill audit' to check security.\n", len(lock.Skills))

	return nil
}

// --- Remove ---

func runSkillRemove(_ *cobra.Command, args []string) error {
	name := args[0]

	projectRoot, err := metadata.FindProjectRoot()
	if err != nil {
		return fmt.Errorf("not in a SpecLedger project\n→ Run 'sl init' first")
	}

	lockPath := filepath.Join(projectRoot, "skills-lock.json")

	if !skills.IsSkillInstalled(name, lockPath) {
		return fmt.Errorf("skill %q is not installed.\n→ Use 'sl skill list' to see installed skills", name)
	}

	agentPaths, err := resolveAgentSkillPaths(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to resolve agent paths: %w", err)
	}

	// Read source before removing for telemetry
	lock, _ := skills.ReadLocalLock(lockPath)
	entry := lock.Skills[name]

	if err := skills.UninstallSkill(name, agentPaths, lockPath); err != nil {
		return err
	}

	// Telemetry
	client := skills.NewClient()
	skills.Track(client.AuditURL, "remove",
		skills.BuildTelemetryParams(entry.Source, []string{name}, nil),
		version.GetVersion(),
	)

	if skillRemoveJSON {
		return json.NewEncoder(os.Stdout).Encode(map[string]string{
			"name":   name,
			"status": "removed",
		})
	}

	fmt.Printf("✓ Removed %s from agent skills directories\n", name)
	fmt.Println("✓ Updated skills-lock.json")

	return nil
}

// --- Audit ---

func runSkillAudit(_ *cobra.Command, args []string) error {
	projectRoot, err := metadata.FindProjectRoot()
	if err != nil {
		return fmt.Errorf("not in a SpecLedger project\n→ Run 'sl init' first")
	}

	lockPath := filepath.Join(projectRoot, "skills-lock.json")
	lock, err := skills.ReadLocalLock(lockPath)
	if err != nil {
		return err
	}

	client := skills.NewClient()

	// Determine which skills to audit
	var toAudit []string
	if len(args) > 0 {
		name := args[0]
		if _, ok := lock.Skills[name]; !ok {
			return fmt.Errorf("skill %q is not installed.\n→ Use 'sl skill list' to see installed skills", name)
		}
		toAudit = []string{name}
	} else {
		if len(lock.Skills) == 0 {
			fmt.Println("No skills installed.")
			fmt.Fprintf(os.Stderr, "→ Use 'sl skill search' to discover and install skills.\n")
			return nil
		}
		for name := range lock.Skills {
			toAudit = append(toAudit, name)
		}
		sort.Strings(toAudit)
	}

	// Group by source for batch API calls
	sourceSkills := make(map[string][]string)
	for _, name := range toAudit {
		entry := lock.Skills[name]
		sourceSkills[entry.Source] = append(sourceSkills[entry.Source], name)
	}

	allResults := make(map[string]*skills.SkillAuditResult)
	for source, slugs := range sourceSkills {
		results, err := client.FetchAudit(source, slugs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: audit failed for %s: %v\n", source, err)
			continue
		}
		for k, v := range results {
			allResults[k] = v
		}
	}

	if skillAuditJSON {
		return json.NewEncoder(os.Stdout).Encode(allResults)
	}

	fmt.Printf("Security Risk Assessments for %d installed skill(s)\n\n", len(toAudit))
	printAuditTable(allResults)

	// Warning summary
	hasHighRisk := false
	for _, r := range allResults {
		if isHighOrCritical(r) {
			hasHighRisk = true
			break
		}
	}

	fmt.Println()
	if hasHighRisk {
		fmt.Fprintf(os.Stderr, "⚠ Some skills have HIGH or CRITICAL risk. Review before using.\n")
	} else {
		fmt.Println("✓ No high or critical risks detected.")
	}

	return nil
}

// --- Agent Path Resolution ---

// resolveAgentSkillPaths returns the skill installation directories for all
// configured agents in the project. Reads agent selection from the constitution
// file and maps to ConfigDir via the agent registry.
func resolveAgentSkillPaths(projectRoot string) ([]string, error) {
	constitutionPath := filepath.Join(projectRoot, "specledger", "constitution.md")
	agents, err := ReadSelectedAgents(constitutionPath)
	if err != nil || len(agents) == 0 {
		agents = []string{"claude"}
	}

	var paths []string
	for _, name := range agents {
		a, ok := agent.Lookup(name)
		if !ok {
			continue
		}
		paths = append(paths, filepath.Join(projectRoot, a.ConfigDir, "skills"))
	}

	if len(paths) == 0 {
		paths = []string{filepath.Join(projectRoot, ".claude", "skills")}
	}

	return paths, nil
}

// --- Helpers ---

func formatInstalls(n int) string {
	switch {
	case n >= 1000:
		return fmt.Sprintf("%.1fK installs", float64(n)/1000)
	default:
		return fmt.Sprintf("%d installs", n)
	}
}

func printAuditTable(results map[string]*skills.SkillAuditResult) {
	// Print header
	fmt.Printf("%-20s %-12s %-12s %-12s\n", "", "Gen", "Socket", "Snyk")

	// Sort by skill name
	names := make([]string, 0, len(results))
	for name := range results {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		r := results[name]
		fmt.Printf("%-20s %-12s %-12s %-12s\n",
			name,
			formatPartner(r.ATH),
			formatSocket(r.Socket),
			formatPartner(r.Snyk),
		)
	}
}

func printAuditSingle(name string, results map[string]*skills.SkillAuditResult) {
	r, ok := results[name]
	if !ok {
		fmt.Println("  No audit data available")
		return
	}
	if r.ATH != nil {
		fmt.Printf("  Gen:    %s (score: %d, analyzed: %s)\n", formatRisk(r.ATH.Risk), r.ATH.Score, r.ATH.AnalyzedAt.Format("2006-01-02"))
	} else {
		fmt.Println("  Gen:    --")
	}
	if r.Socket != nil {
		fmt.Printf("  Socket: %d alerts (score: %d, analyzed: %s)\n", r.Socket.Alerts, r.Socket.Score, r.Socket.AnalyzedAt.Format("2006-01-02"))
	} else {
		fmt.Println("  Socket: --")
	}
	if r.Snyk != nil {
		fmt.Printf("  Snyk:   %s (score: %d, analyzed: %s)\n", formatRisk(r.Snyk.Risk), r.Snyk.Score, r.Snyk.AnalyzedAt.Format("2006-01-02"))
	} else {
		fmt.Println("  Snyk:   --")
	}
}

func formatPartner(p *skills.PartnerAudit) string {
	if p == nil {
		return "--"
	}
	return formatRisk(p.Risk)
}

func formatSocket(p *skills.PartnerAudit) string {
	if p == nil {
		return "--"
	}
	return fmt.Sprintf("%d alerts", p.Alerts)
}

func formatRisk(risk string) string {
	switch strings.ToLower(risk) {
	case "safe":
		return "Safe"
	case "low":
		return "Low Risk"
	case "medium":
		return "Med Risk"
	case "high":
		return "High Risk"
	case "critical":
		return "Critical"
	default:
		return strings.Title(risk) //nolint:staticcheck
	}
}

func isHighOrCritical(r *skills.SkillAuditResult) bool {
	for _, p := range []*skills.PartnerAudit{r.ATH, r.Socket, r.Snyk} {
		if p == nil {
			continue
		}
		risk := strings.ToLower(p.Risk)
		if risk == "high" || risk == "critical" {
			return true
		}
	}
	return false
}

func confirm() bool {
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(strings.ToLower(text))
	return text == "y" || text == "yes"
}

func confirmDefault(defaultYes bool) bool {
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(strings.ToLower(text))
	if text == "" {
		return defaultYes
	}
	return text == "y" || text == "yes"
}
