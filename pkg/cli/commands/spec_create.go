package commands

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/specledger/specledger/pkg/cli/spec"
	"github.com/specledger/specledger/pkg/embedded"
	"github.com/spf13/cobra"
)

type SpecCreateOutput struct {
	BranchName string `json:"BRANCH_NAME"`
	FeatureDir string `json:"FEATURE_DIR"`
	SpecFile   string `json:"SPEC_FILE"`
	FeatureNum string `json:"FEATURE_NUM"`
}

var specCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new feature branch and spec directory",
	Long: `Create a new feature branch and spec directory with template files.

This command creates a new feature by:
1. Generating a branch name from the description (with stop-word filtering)
2. Checking for feature number collisions
3. Creating the feature branch
4. Creating the spec directory
5. Copying the spec template

Examples:
  sl spec create --number 600 --short-name "test-feature"
  sl spec create --number 600 --short-name "add OAuth2 authentication" --json
  sl spec create --number 600 --short-name "very long description that will be truncated automatically"`,
	RunE: runSpecCreate,
}

func init() {
	VarSpecCmd.AddCommand(specCreateCmd)

	specCreateCmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	specCreateCmd.Flags().String("number", "", "Feature number (e.g., 600)")
	specCreateCmd.Flags().String("short-name", "", "Short name or description for the feature")
}

func runSpecCreate(cmd *cobra.Command, args []string) error {
	jsonOutput, _ := cmd.Flags().GetBool("json")
	numberStr, _ := cmd.Flags().GetString("number")
	shortName, _ := cmd.Flags().GetString("short-name")

	if numberStr == "" {
		return fmt.Errorf("--number flag is required")
	}

	if shortName == "" {
		return fmt.Errorf("--short-name flag is required")
	}

	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	repo, err := gogit.PlainOpenWithOptions(workDir, &gogit.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return fmt.Errorf("failed to open git repository: %w", err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	repoRoot := wt.Filesystem.Root()

	if err := spec.CheckFeatureCollision(repoRoot, numberStr); err != nil {
		return fmt.Errorf("collision detected: %w", err)
	}

	branchName := spec.GenerateBranchName(shortName, parseNumber(numberStr))

	if len(branchName) > spec.MaxBranchLength {
		branchName = spec.TruncateToLimit(branchName, spec.MaxBranchLength)
		if !jsonOutput {
			fmt.Fprintf(os.Stderr, "Warning: branch name truncated to %d bytes\n", spec.MaxBranchLength)
		}
	}

	refName := plumbing.ReferenceName("refs/heads/" + branchName)

	headRef, err := repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD: %w", err)
	}

	err = wt.Checkout(&gogit.CheckoutOptions{
		Branch: refName,
		Create: true,
		Hash:   headRef.Hash(),
		Keep:   true,
	})
	if err != nil {
		return fmt.Errorf("failed to create branch %q: %w", branchName, err)
	}

	featureDir := spec.GetFeatureDir(repoRoot, branchName)

	if err := spec.EnsureDir(featureDir); err != nil {
		return fmt.Errorf("failed to create feature directory: %w", err)
	}

	specFile := spec.GetSpecFile(featureDir)

	templateContent, err := readSpecTemplate()
	if err != nil {
		return fmt.Errorf("failed to read spec template: %w", err)
	}

	if err := os.WriteFile(specFile, templateContent, 0600); err != nil {
		return fmt.Errorf("failed to write spec file: %w", err)
	}

	output := SpecCreateOutput{
		BranchName: branchName,
		FeatureDir: featureDir,
		SpecFile:   specFile,
		FeatureNum: numberStr,
	}

	if jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetEscapeHTML(false)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(output); err != nil {
			return fmt.Errorf("failed to encode JSON output: %w", err)
		}
	} else {
		fmt.Printf("Created feature branch: %s\n", output.BranchName)
		fmt.Printf("Feature directory: %s\n", output.FeatureDir)
		fmt.Printf("Spec file: %s\n", output.SpecFile)
		fmt.Printf("Feature number: %s\n", output.FeatureNum)
	}

	return nil
}

func parseNumber(s string) int {
	var num int
	_, _ = fmt.Sscanf(s, "%d", &num)
	return num
}

func readSpecTemplate() ([]byte, error) {
	templatePath := filepath.Join("specledger", ".specledger", "templates", "spec-template.md")

	content, err := embedded.TemplatesFS.ReadFile(templatePath)
	if err != nil {
		altPath := filepath.Join("templates", "specledger", ".specledger", "templates", "spec-template.md")
		content, err = embedded.TemplatesFS.ReadFile(altPath)
		if err != nil {
			return tryAlternativeTemplatePaths()
		}
	}

	return content, nil
}

func tryAlternativeTemplatePaths() ([]byte, error) {
	paths := []string{
		"spec-template.md",
		filepath.Join("templates", "spec-template.md"),
		filepath.Join(".specledger", "templates", "spec-template.md"),
	}

	for _, p := range paths {
		content, err := fs.ReadFile(embedded.TemplatesFS, p)
		if err == nil {
			return content, nil
		}
	}

	return []byte("# Feature Specification\n\n**Feature Branch**: `###-feature-name`\n**Created**: \n**Status**: Draft\n\n## Overview\n\n[Describe the feature]\n\n## Requirements\n\n### Functional Requirements\n\n- [ ] FR-001: [Requirement]\n\n## Success Criteria\n\n- [ ] SC-001: [Success criterion]\n"), nil
}

func NewSpecCreateCmd() *cobra.Command {
	return specCreateCmd
}
