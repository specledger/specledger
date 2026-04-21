package skills

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// SkillMetadata is parsed from SKILL.md YAML frontmatter.
type SkillMetadata struct {
	Name        string   `yaml:"name" json:"name"`
	Description string   `yaml:"description" json:"description"`
	Slug        string   `yaml:"-" json:"slug,omitempty"`
	Source      string   `yaml:"-" json:"source,omitempty"`
	RepoPath    string   `yaml:"-" json:"-"` // path to SKILL.md within the repo (set during discovery)
	Files       []string `yaml:"-" json:"-"` // all repo-relative file paths under the skill directory
	Internal    bool     `yaml:"internal,omitempty" json:"-"`
}

// errTreeTruncated is returned by discoverViaGitHub when the Trees API response
// is truncated (>100k entries or >7 MB). DiscoverSkills uses this to auto-fallback
// to the clone path.
var errTreeTruncated = errors.New("GitHub tree response truncated")

// DiscoverSkills enumerates available skills in a repository.
// For GitHub sources, uses the Trees API (fast path).
// For git URLs or on GitHub 404, falls back to git clone.
func DiscoverSkills(client *Client, source *SkillSource) ([]SkillMetadata, error) {
	if source.Type == SourceTypeGitHub {
		skills, err := discoverViaGitHub(client, source)
		if err == nil {
			return skills, nil
		}
		// Auto-retry via git clone on GitHub API failure (including truncation)
		if errors.Is(err, errTreeTruncated) {
			fmt.Fprintf(os.Stderr, "→ tree response truncated, falling back to clone\n")
		}
		cloneURL := fmt.Sprintf("https://github.com/%s/%s.git", source.Owner, source.Repo)
		return discoverViaClone(cloneURL, source)
	}

	// Git type: always clone
	return discoverViaClone(source.URL, source)
}

func discoverViaGitHub(client *Client, source *SkillSource) ([]SkillMetadata, error) {
	tree, resolvedRef, truncated, err := client.FetchRepoTree(source.Owner, source.Repo, source.Ref)
	if err != nil {
		return nil, err
	}

	if truncated {
		return nil, errTreeTruncated
	}

	// Propagate the resolved ref so install fetches use the correct branch
	source.Ref = resolvedRef

	// Build set of skill directories (each containing a SKILL.md)
	skillDirs := make(map[string]string) // dir → full SKILL.md path
	for _, entry := range tree {
		if entry.Type != "blob" {
			continue
		}
		if !strings.HasSuffix(entry.Path, "/SKILL.md") {
			continue
		}
		dir := strings.TrimSuffix(entry.Path, "/SKILL.md")
		skillDirs[dir] = entry.Path
	}

	if len(skillDirs) == 0 {
		return nil, fmt.Errorf("no skills found in %s/%s\n→ The repository may not contain SKILL.md files", source.Owner, source.Repo)
	}

	// For each tree entry, find the deepest enclosing skill dir.
	// This avoids assigning a file to a parent skill when it belongs to a nested skill.
	findOwningSkillDir := func(filePath string) string {
		best := ""
		for dir := range skillDirs {
			prefix := dir + "/"
			if strings.HasPrefix(filePath, prefix) && len(dir) > len(best) {
				best = dir
			}
		}
		return best
	}

	// Build file lists per skill dir, skipping symlinks and submodules
	skillFiles := make(map[string][]string) // skill dir → list of repo-relative paths
	for _, entry := range tree {
		if entry.Type != "blob" {
			continue
		}
		// Skip symlinks (120000) and submodules (160000)
		if entry.Mode == "120000" || entry.Mode == "160000" {
			continue
		}
		owningDir := findOwningSkillDir(entry.Path)
		if owningDir == "" {
			continue
		}
		skillFiles[owningDir] = append(skillFiles[owningDir], entry.Path)
	}

	// Sort file lists for deterministic order (VCR cassettes depend on this)
	for dir := range skillFiles {
		sort.Strings(skillFiles[dir])
	}

	var skills []SkillMetadata
	for dir, skillMdPath := range skillDirs {
		content, err := client.FetchSkillContent(source.Owner, source.Repo, source.Ref, skillMdPath)
		if err != nil {
			continue // skip individual fetch failures
		}

		meta, err := ParseSkillFrontmatter(content)
		if err != nil {
			continue
		}

		meta.Source = source.SourceString()
		meta.Slug = source.SourceString() + "/" + meta.Name
		meta.RepoPath = skillMdPath // preserve discovered path for install fetch
		meta.Files = skillFiles[dir]

		// Apply skill filter if specified
		if source.SkillFilter != "" && meta.Name != source.SkillFilter {
			continue
		}

		skills = append(skills, *meta)
	}

	if len(skills) == 0 && source.SkillFilter != "" {
		return nil, fmt.Errorf("skill %q not found in %s\n→ Use 'sl skill add %s' to see available skills",
			source.SkillFilter, source.SourceString(), source.SourceString())
	}

	if len(skills) == 0 {
		return nil, fmt.Errorf("no valid skills found in %s/%s\n→ SKILL.md files may be missing name/description frontmatter", source.Owner, source.Repo)
	}

	return skills, nil
}

func discoverViaClone(cloneURL string, source *SkillSource) ([]SkillMetadata, error) {
	tmpDir, err := os.MkdirTemp("", "sl-skill-clone-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	// Store on source so FetchSkillFiles can read from the clone.
	// Caller is responsible for calling source.Cleanup().
	source.cloneDir = tmpDir

	// git clone --depth 1; omit --branch when ref is empty to use repo's default
	args := []string{"clone", "--depth", "1"}
	if source.Ref != "" {
		args = append(args, "--branch", source.Ref)
	}
	args = append(args, cloneURL, tmpDir)
	cmd := exec.Command("git", args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("sl skill add failed: could not clone %q\n→ Verify the repository exists and is accessible", cloneURL)
	}

	// Scan for SKILL.md files and collect all files per skill directory
	type skillCandidate struct {
		meta     *SkillMetadata
		skillDir string // absolute path to the skill directory in the clone
	}
	var candidates []skillCandidate

	err = filepath.Walk(tmpDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return nil // skip errors
		}
		if info.IsDir() {
			name := info.Name()
			if name == ".git" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if info.Name() != "SKILL.md" {
			return nil
		}

		// Skip symlinked SKILL.md that resolves outside the clone
		if info.Mode()&os.ModeSymlink != 0 {
			resolved, resolveErr := filepath.EvalSymlinks(path)
			if resolveErr != nil {
				return nil
			}
			if !strings.HasPrefix(resolved, tmpDir+string(filepath.Separator)) {
				return nil
			}
		}

		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}

		meta, parseErr := ParseSkillFrontmatter(content)
		if parseErr != nil {
			return nil
		}

		meta.Source = source.SourceString()
		meta.Slug = source.SourceString() + "/" + meta.Name
		// Preserve relative path within the cloned repo for install fetch
		relPath, _ := filepath.Rel(tmpDir, path)
		if relPath != "" {
			meta.RepoPath = filepath.ToSlash(relPath)
		}

		if source.SkillFilter != "" && meta.Name != source.SkillFilter {
			return nil
		}

		candidates = append(candidates, skillCandidate{
			meta:     meta,
			skillDir: filepath.Dir(path),
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to scan cloned repository: %w", err)
	}

	// Build set of all skill dirs for nested-skill isolation
	allSkillDirs := make(map[string]bool, len(candidates))
	for _, c := range candidates {
		allSkillDirs[c.skillDir] = true
	}

	// Walk each skill directory to collect files, skipping nested skill subtrees
	for i := range candidates {
		c := &candidates[i]
		skillDirAbs := c.skillDir
		var files []string
		_ = filepath.Walk(skillDirAbs, func(p string, fi os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return nil
			}
			if fi.IsDir() {
				// Skip nested skill directories (they belong to their own candidate)
				if p != skillDirAbs && allSkillDirs[p] {
					return filepath.SkipDir
				}
				return nil
			}

			// Symlink escape check: resolve symlink and verify it stays under the skill dir
			if fi.Mode()&os.ModeSymlink != 0 {
				resolved, resolveErr := filepath.EvalSymlinks(p)
				if resolveErr != nil {
					return nil // skip broken symlinks
				}
				if !strings.HasPrefix(resolved, skillDirAbs+string(filepath.Separator)) {
					return nil // skip symlinks that escape the skill directory
				}
			}

			rel, _ := filepath.Rel(tmpDir, p)
			files = append(files, filepath.ToSlash(rel))
			return nil
		})
		sort.Strings(files)
		c.meta.Files = files
	}

	var skills []SkillMetadata
	for _, c := range candidates {
		skills = append(skills, *c.meta)
	}

	if len(skills) == 0 && source.SkillFilter != "" {
		return nil, fmt.Errorf("skill %q not found in %s\n→ Use 'sl skill add %s' to see available skills",
			source.SkillFilter, source.SourceString(), source.SourceString())
	}
	if len(skills) == 0 {
		return nil, fmt.Errorf("no skills found in repository\n→ The repository may not contain SKILL.md files")
	}

	// Set ref to HEAD if unset so install fetches use the repo's default branch
	if source.Ref == "" {
		source.Ref = "HEAD"
	}

	return skills, nil
}

// ParseSkillFrontmatter parses YAML frontmatter from SKILL.md content.
// Expects --- delimited frontmatter with at least name and description fields.
func ParseSkillFrontmatter(content []byte) (*SkillMetadata, error) {
	text := string(content)

	if !strings.HasPrefix(text, "---") {
		return nil, fmt.Errorf("SKILL.md missing frontmatter")
	}

	end := strings.Index(text[3:], "---")
	if end < 0 {
		return nil, fmt.Errorf("SKILL.md frontmatter not terminated")
	}

	frontmatter := text[3 : 3+end]

	var meta SkillMetadata
	if err := yaml.Unmarshal([]byte(frontmatter), &meta); err != nil {
		return nil, fmt.Errorf("invalid SKILL.md frontmatter: %w", err)
	}

	if meta.Name == "" {
		return nil, fmt.Errorf("SKILL.md missing required 'name' field")
	}
	if meta.Description == "" {
		return nil, fmt.Errorf("SKILL.md missing required 'description' field")
	}

	if err := ValidateSkillName(meta.Name); err != nil {
		return nil, fmt.Errorf("SKILL.md has unsafe name: %w", err)
	}

	return &meta, nil
}
