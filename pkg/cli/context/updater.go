package context

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	ManualAdditionsStart = "<!-- MANUAL ADDITIONS START -->"
	ManualAdditionsEnd   = "<!-- MANUAL ADDITIONS END -->"
)

type AgentUpdater struct {
	AgentType string
	FilePath  string
}

func NewAgentUpdater(agentType, repoRoot string) *AgentUpdater {
	agentFileMap := map[string]string{
		"claude":    "CLAUDE.md",
		"gemini":    "GEMINI.md",
		"copilot":   ".github/agents/copilot-instructions.md",
		"cursor":    ".cursor/rules/specify-rules.mdc",
		"qwen":      "QWEN.md",
		"windsurf":  ".windsurf/rules/specify-rules.md",
		"kilocode":  ".kilocode/rules/specify-rules.md",
		"auggie":    ".augment/rules/specify-rules.md",
		"roo":       ".roo/rules/specify-rules.md",
		"codebuddy": "CODEBUDDY.md",
		"qoder":     "QODER.md",
		"shai":      "SHAI.md",
		"amazonq":   "AGENTS.md",
		"ibmbob":    "AGENTS.md",
		"opencode":  "AGENTS.md",
		"codex":     "AGENTS.md",
	}

	fileName := agentFileMap[agentType]
	if fileName == "" {
		fileName = "AGENTS.md"
	}

	return &AgentUpdater{
		AgentType: agentType,
		FilePath:  filepath.Join(repoRoot, fileName),
	}
}

func (u *AgentUpdater) Update(ctx *TechnicalContext) error {
	content, err := u.readFile()
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to read agent file: %w", err)
		}
		content = ""
	}

	manualAdditions := u.extractManualAdditions(content)

	activeTech := u.generateActiveTechnologies(ctx)

	newContent := u.buildContent(activeTech, manualAdditions)

	if err := u.writeFile(newContent); err != nil {
		return fmt.Errorf("failed to write agent file: %w", err)
	}

	return nil
}

func (u *AgentUpdater) readFile() (string, error) {
	data, err := os.ReadFile(u.FilePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (u *AgentUpdater) writeFile(content string) error {
	dir := filepath.Dir(u.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	var buf bytes.Buffer
	buf.WriteString(content)

	tmpFile := u.FilePath + ".tmp"
	if err := os.WriteFile(tmpFile, buf.Bytes(), 0600); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := os.Rename(tmpFile, u.FilePath); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

func (u *AgentUpdater) extractManualAdditions(content string) string {
	startIdx := strings.Index(content, ManualAdditionsStart)
	endIdx := strings.Index(content, ManualAdditionsEnd)

	if startIdx == -1 || endIdx == -1 || startIdx >= endIdx {
		return ""
	}

	return content[startIdx : endIdx+len(ManualAdditionsEnd)]
}

func (u *AgentUpdater) generateActiveTechnologies(ctx *TechnicalContext) string {
	var entries []string

	if ctx.Language != "" {
		entries = append(entries, ctx.Language)
	}
	if ctx.PrimaryDeps != "" {
		deps := strings.Split(ctx.PrimaryDeps, ",")
		for _, dep := range deps {
			dep = strings.TrimSpace(dep)
			if dep != "" {
				entries = append(entries, dep)
			}
		}
	}
	if ctx.Storage != "" {
		entries = append(entries, ctx.Storage)
	}
	if ctx.Testing != "" {
		entries = append(entries, ctx.Testing)
	}

	entries = u.deduplicateEntries(entries)

	var lines []string
	lines = append(lines, "## Active Technologies")
	lines = append(lines, "")
	for _, entry := range entries {
		lines = append(lines, fmt.Sprintf("- %s", entry))
	}

	return strings.Join(lines, "\n")
}

func (u *AgentUpdater) deduplicateEntries(entries []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, entry := range entries {
		lower := strings.ToLower(strings.TrimSpace(entry))
		if !seen[lower] && lower != "" {
			seen[lower] = true
			result = append(result, strings.TrimSpace(entry))
		}
	}

	sort.Strings(result)

	return result
}

func (u *AgentUpdater) buildContent(activeTech, manualAdditions string) string {
	var lines []string

	lines = append(lines, "# Active Technologies")
	lines = append(lines, "")
	lines = append(lines, "This file is auto-generated from plan.md. Manual additions are preserved below.")
	lines = append(lines, "")
	lines = append(lines, activeTech)
	lines = append(lines, "")

	if manualAdditions != "" {
		lines = append(lines, manualAdditions)
	} else {
		lines = append(lines, ManualAdditionsStart)
		lines = append(lines, "")
		lines = append(lines, ManualAdditionsEnd)
	}

	lines = append(lines, "")

	return strings.Join(lines, "\n")
}

func (u *AgentUpdater) DiscoverAgentFiles(repoRoot string) ([]string, error) {
	var files []string

	agentPatterns := []string{
		"CLAUDE.md",
		"GEMINI.md",
		"AGENTS.md",
		"QWEN.md",
		"CODEBUDDY.md",
		"QODER.md",
		"SHAI.md",
		".github/agents/copilot-instructions.md",
		".cursor/rules/specify-rules.mdc",
		".windsurf/rules/specify-rules.md",
		".kilocode/rules/specify-rules.md",
		".augment/rules/specify-rules.md",
		".roo/rules/specify-rules.md",
	}

	for _, pattern := range agentPatterns {
		path := filepath.Join(repoRoot, pattern)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			files = append(files, path)
		}
	}

	return files, nil
}

func WalkAgentFiles(root string) ([]string, error) {
	var files []string

	err := fs.WalkDir(os.DirFS(root), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		name := strings.ToLower(d.Name())
		if name == "claude.md" || name == "gemini.md" || name == "agents.md" ||
			name == "qwen.md" || name == "codebuddy.md" || name == "qoder.md" || name == "shai.md" {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

func (u *AgentUpdater) PreserveManualAdditions(content string) string {
	return u.extractManualAdditions(content)
}

func (u *AgentUpdater) DeduplicateEntries(entries []string) []string {
	return u.deduplicateEntries(entries)
}
