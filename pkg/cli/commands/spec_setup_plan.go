package commands

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/specledger/specledger/pkg/cli/spec"
	"github.com/specledger/specledger/pkg/embedded"
	"github.com/spf13/cobra"
)

type SpecSetupPlanOutput struct {
	PlanFile string `json:"PLAN_FILE"`
}

var specSetupPlanCmd = &cobra.Command{
	Use:   "setup-plan",
	Short: "Copy plan template to feature directory",
	Long: `Copy the plan template to the feature directory.

This command creates a plan.md file in the feature directory from the
embedded template. It will error if plan.md already exists to prevent
overwriting existing work.

The template includes:
  - Implementation plan structure
  - Technical context section
  - Phase breakdown
  - Success criteria tracking

Examples:
  sl spec setup-plan              # Create plan.md in current feature
  sl spec setup-plan --json       # Output as JSON`,
	RunE: runSpecSetupPlan,
}

func init() {
	VarSpecCmd.AddCommand(specSetupPlanCmd)

	specSetupPlanCmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	specSetupPlanCmd.Flags().String("spec", "", "Override feature spec name (bypasses detection)")
}

func runSpecSetupPlan(cmd *cobra.Command, args []string) error {
	jsonOutput, _ := cmd.Flags().GetBool("json")
	specOverride, _ := cmd.Flags().GetString("spec")

	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	opts := spec.DetectionOptions{
		SpecOverride: specOverride,
	}

	ctx, err := spec.DetectFeatureContextWithOptions(workDir, opts)
	if err != nil {
		return fmt.Errorf("failed to detect feature context: %w", err)
	}

	if spec.FileExists(ctx.PlanFile) {
		return fmt.Errorf("plan.md already exists at: %s", ctx.PlanFile)
	}

	templateContent, err := readPlanTemplate()
	if err != nil {
		return fmt.Errorf("failed to read plan template: %w", err)
	}

	if err := spec.EnsureDir(ctx.FeatureDir); err != nil {
		return fmt.Errorf("failed to create feature directory: %w", err)
	}

	if err := os.WriteFile(ctx.PlanFile, templateContent, 0600); err != nil {
		return fmt.Errorf("failed to write plan.md: %w", err)
	}

	output := SpecSetupPlanOutput{
		PlanFile: ctx.PlanFile,
	}

	if jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetEscapeHTML(false)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(output); err != nil {
			return fmt.Errorf("failed to encode JSON output: %w", err)
		}
	} else {
		fmt.Printf("Created plan file: %s\n", output.PlanFile)
		fmt.Printf("Feature: %s\n", ctx.Branch)
	}

	return nil
}

func readPlanTemplate() ([]byte, error) {
	templatePath := filepath.Join("specledger", ".specledger", "templates", "plan-template.md")

	content, err := embedded.TemplatesFS.ReadFile(templatePath)
	if err != nil {
		altPath := filepath.Join("templates", "specledger", ".specledger", "templates", "plan-template.md")
		content, err = embedded.TemplatesFS.ReadFile(altPath)
		if err != nil {
			return tryAlternativePlanTemplatePaths()
		}
	}

	return content, nil
}

func tryAlternativePlanTemplatePaths() ([]byte, error) {
	paths := []string{
		"plan-template.md",
		filepath.Join("templates", "plan-template.md"),
		filepath.Join(".specledger", "templates", "plan-template.md"),
	}

	for _, p := range paths {
		content, err := fs.ReadFile(embedded.TemplatesFS, p)
		if err == nil {
			return content, nil
		}
	}

	return []byte(`# Implementation Plan: [Feature Name]

**Branch**: ` + "`###-feature-name`" + ` | **Date**: | **Spec**: [spec.md](./spec.md)

## Summary

[Brief description of what this feature implements]

## Technical Context

**Language/Version**: 
**Primary Dependencies**: 
**Storage**: 
**Project Type**: 
**Target Platform**: 

## Phase 0: Research Summary

[Research findings and decisions]

## Phase 1: Design

[Design decisions and architecture]

## Phase 2: Work Breakdown

[Task breakdown by phase]

## Success Criteria

- [ ] SC-001: [Success criterion]
`), nil
}

func NewSpecSetupPlanCmd() *cobra.Command {
	return specSetupPlanCmd
}
