package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidateSkillName rejects names that could cause path traversal or
// escape the skills directory. Called before any filesystem operation.
func ValidateSkillName(name string) error {
	if name == "" {
		return fmt.Errorf("skill name is empty")
	}
	if strings.ContainsAny(name, "/\\") {
		return fmt.Errorf("skill name %q contains path separators", name)
	}
	if strings.Contains(name, "..") {
		return fmt.Errorf("skill name %q contains path traversal", name)
	}
	if filepath.IsAbs(name) {
		return fmt.Errorf("skill name %q is an absolute path", name)
	}
	cleaned := filepath.Clean(name)
	if cleaned != name {
		return fmt.Errorf("skill name %q is not clean (resolves to %q)", name, cleaned)
	}
	return nil
}

// InstallSkill writes SKILL.md content to each agent's skills directory
// and updates the lock file.
func InstallSkill(name string, content []byte, agentPaths []string, lockPath string, source *SkillSource) error {
	if err := ValidateSkillName(name); err != nil {
		return fmt.Errorf("unsafe skill name: %w", err)
	}
	if len(agentPaths) == 0 {
		return fmt.Errorf("no agent paths configured\n→ Run 'sl init' to configure agents")
	}

	for _, basePath := range agentPaths {
		skillDir := filepath.Join(basePath, name)
		if err := os.MkdirAll(skillDir, 0755); err != nil {
			return fmt.Errorf("failed to create skill directory %s: %w", skillDir, err)
		}

		skillFile := filepath.Join(skillDir, "SKILL.md")
		// #nosec G306 -- skill files need to be readable, 0644 is appropriate
		if err := os.WriteFile(skillFile, content, 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", skillFile, err)
		}
	}

	// Compute hash from the first agent path (all paths have identical content)
	hashDir := filepath.Join(agentPaths[0], name)
	hash, err := ComputeFolderHash(hashDir)
	if err != nil {
		return fmt.Errorf("failed to compute skill hash: %w", err)
	}

	// Update lock file
	lock, err := ReadLocalLock(lockPath)
	if err != nil {
		return fmt.Errorf("failed to read lock file: %w", err)
	}

	entry := LocalSkillLockEntry{
		Source:       source.SourceString(),
		SourceType:   string(source.Type),
		ComputedHash: hash,
	}
	if source.Ref != "" {
		entry.Ref = source.Ref
	}

	AddSkill(lock, name, entry)

	if err := WriteLocalLock(lockPath, lock); err != nil {
		return fmt.Errorf("failed to update lock file: %w", err)
	}

	return nil
}

// UninstallSkill removes a skill from all agent directories and the lock file.
func UninstallSkill(name string, agentPaths []string, lockPath string) error {
	if err := ValidateSkillName(name); err != nil {
		return fmt.Errorf("unsafe skill name: %w", err)
	}
	for _, basePath := range agentPaths {
		skillDir := filepath.Join(basePath, name)
		if err := os.RemoveAll(skillDir); err != nil {
			return fmt.Errorf("failed to remove %s: %w", skillDir, err)
		}
	}

	lock, err := ReadLocalLock(lockPath)
	if err != nil {
		return fmt.Errorf("failed to read lock file: %w", err)
	}

	RemoveSkill(lock, name)

	if err := WriteLocalLock(lockPath, lock); err != nil {
		return fmt.Errorf("failed to update lock file: %w", err)
	}

	return nil
}

// IsSkillInstalled checks if a skill is present in the lock file.
func IsSkillInstalled(name, lockPath string) bool {
	lock, err := ReadLocalLock(lockPath)
	if err != nil {
		return false
	}
	_, ok := lock.Skills[name]
	return ok
}
