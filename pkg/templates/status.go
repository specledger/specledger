// Package templates provides template comparison and update utilities.
package templates

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/specledger/specledger/pkg/cli/metadata"
	"github.com/specledger/specledger/pkg/cli/playbooks"
	"github.com/specledger/specledger/pkg/embedded"
)

// TemplateStatus represents the state of project templates relative to current CLI.
type TemplateStatus struct {
	ProjectTemplateVersion string   `json:"project_template_version"` // Version stored in specledger.yaml
	CLIVersion             string   `json:"cli_version"`              // Current CLI version
	UpdateAvailable        bool     `json:"update_available"`         // true if versions differ
	CustomizedFiles        []string `json:"customized_files"`         // Files that differ from embedded originals
	TotalFiles             int      `json:"total_files"`              // Total template files in project
	NeedsUpdate            bool     `json:"needs_update"`             // true if UpdateAvailable and in project
	InProject              bool     `json:"in_project"`               // true if running in a SpecLedger project
}

// CheckTemplateStatus checks the template status for a project.
// Returns TemplateStatus with information about template version and customization.
func CheckTemplateStatus(projectDir, cliVersion string) (*TemplateStatus, error) {
	status := &TemplateStatus{
		CLIVersion:      cliVersion,
		CustomizedFiles: []string{},
		InProject:       false,
	}

	// Check if we're in a SpecLedger project
	if !metadata.HasYAMLMetadata(projectDir) {
		return status, nil
	}

	status.InProject = true

	// Load project metadata
	meta, err := metadata.LoadFromProject(projectDir)
	if err != nil {
		return nil, err
	}

	status.ProjectTemplateVersion = meta.TemplateVersion

	// Determine if update is needed
	// If template_version is empty, assume update is needed
	if meta.TemplateVersion == "" {
		status.UpdateAvailable = true
		status.NeedsUpdate = true
	} else {
		status.UpdateAvailable = meta.TemplateVersion != cliVersion
		status.NeedsUpdate = status.UpdateAvailable
	}

	// Load playbook manifest to know what files to check
	source, err := playbooks.NewEmbeddedSource()
	if err != nil {
		return status, nil // Can't load manifest, return status as-is
	}

	playbook, err := source.GetDefaultPlaybook()
	if err != nil {
		return status, nil // Can't get playbook, return status as-is
	}

	// Check template files based on manifest
	// 1. Check structure items
	for _, structureItem := range playbook.Structure {
		checkStructureItem(projectDir, structureItem, status)
	}

	// 2. Check commands in .claude/commands/
	for _, cmd := range playbook.Commands {
		// cmd.Path is like "commands/specledger.specify.md"
		// Target in project is .claude/commands/specledger.specify.md
		relPath := filepath.Join(".claude", cmd.Path)
		checkFileCustomization(projectDir, relPath, status)
	}

	// 3. Check skills in .claude/skills/
	for _, skill := range playbook.Skills {
		// skill.Path is like "skills/sl-audit/skill.md"
		// Target in project is .claude/skills/sl-audit/skill.md
		relPath := filepath.Join(".claude", skill.Path)
		checkFileCustomization(projectDir, relPath, status)
	}

	return status, nil
}

// checkStructureItem checks if a structure item (file or directory) exists and is customized.
func checkStructureItem(projectDir, structureItem string, status *TemplateStatus) {
	itemPath := filepath.Join(projectDir, structureItem)

	// Check if it's a directory or file
	info, err := os.Stat(itemPath)
	if err != nil {
		return // Item doesn't exist, skip
	}

	if info.IsDir() {
		// It's a directory - walk and check all files
		_ = filepath.Walk(itemPath, func(path string, fi os.FileInfo, err error) error {
			if err != nil || fi.IsDir() {
				return nil
			}

			relPath, err := filepath.Rel(projectDir, path)
			if err != nil {
				return nil
			}

			checkFileCustomization(projectDir, relPath, status)
			return nil
		})
	} else {
		// It's a single file
		checkFileCustomization(projectDir, structureItem, status)
	}
}

// checkFileCustomization checks if a single file is customized compared to the embedded version.
func checkFileCustomization(projectDir, relPath string, status *TemplateStatus) {
	status.TotalFiles++

	// Path in TemplatesFS is templates/specledger/<relPath>
	// But for .claude files, the embedded path doesn't have .claude prefix
	// Commands are at templates/specledger/commands/...
	// Skills are at templates/specledger/skills/...
	// Structure items are at templates/specledger/<item>

	// Determine the path in embedded FS
	// Strip .claude/ prefix for embedded path lookup if present
	var relPathInFS string
	prefix := ".claude" + string(filepath.Separator)
	if rest, found := strings.CutPrefix(relPath, prefix); found {
		relPathInFS = "templates/specledger/" + rest
	} else {
		relPathInFS = "templates/specledger/" + relPath
	}

	projectPath := filepath.Join(projectDir, relPath)
	isCustom, _ := IsFileCustomized(projectPath, relPathInFS, embedded.TemplatesFS)
	if isCustom {
		status.CustomizedFiles = append(status.CustomizedFiles, relPath)
	}
}
