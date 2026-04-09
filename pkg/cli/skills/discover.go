package skills

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// SkillMetadata is parsed from SKILL.md YAML frontmatter.
type SkillMetadata struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
	Slug        string `yaml:"-" json:"slug,omitempty"`
	Source      string `yaml:"-" json:"source,omitempty"`
	RepoPath    string `yaml:"-" json:"-"` // path to SKILL.md within the repo (set during discovery)
	Internal    bool   `yaml:"internal,omitempty" json:"-"`
}

// DiscoverSkills enumerates available skills in a repository.
// For GitHub sources, uses the Trees API (fast path).
// For git URLs or on GitHub 404, falls back to git clone.
func DiscoverSkills(client *Client, source *SkillSource) ([]SkillMetadata, error) {
	if source.Type == SourceTypeGitHub {
		skills, err := discoverViaGitHub(client, source)
		if err == nil {
			return skills, nil
		}
		// Auto-retry via git clone on GitHub API failure
		cloneURL := fmt.Sprintf("https://github.com/%s/%s.git", source.Owner, source.Repo)
		return discoverViaClone(cloneURL, source)
	}

	// Git type: always clone
	return discoverViaClone(source.URL, source)
}

func discoverViaGitHub(client *Client, source *SkillSource) ([]SkillMetadata, error) {
	tree, err := client.FetchRepoTree(source.Owner, source.Repo, source.Ref)
	if err != nil {
		return nil, err
	}

	// Find SKILL.md files in the tree
	var skillPaths []string
	for _, entry := range tree {
		if entry.Type != "blob" {
			continue
		}
		if !strings.HasSuffix(entry.Path, "/SKILL.md") {
			continue
		}
		skillPaths = append(skillPaths, entry.Path)
	}

	if len(skillPaths) == 0 {
		return nil, fmt.Errorf("no skills found in %s/%s\n→ The repository may not contain SKILL.md files", source.Owner, source.Repo)
	}

	var skills []SkillMetadata
	for _, path := range skillPaths {
		content, err := client.FetchSkillContent(source.Owner, source.Repo, source.Ref, path)
		if err != nil {
			continue // skip individual fetch failures
		}

		meta, err := ParseSkillFrontmatter(content)
		if err != nil {
			continue
		}

		meta.Source = source.SourceString()
		meta.Slug = source.SourceString() + "/" + meta.Name
		meta.RepoPath = path // preserve discovered path for install fetch

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
	defer os.RemoveAll(tmpDir)

	// git clone --depth 1
	cmd := exec.Command("git", "clone", "--depth", "1", cloneURL, tmpDir)
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("sl skill add failed: could not clone %q\n→ Verify the repository exists and is accessible", cloneURL)
	}

	// Scan for SKILL.md files
	var skills []SkillMetadata
	err = filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
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

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		meta, err := ParseSkillFrontmatter(content)
		if err != nil {
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

		skills = append(skills, *meta)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to scan cloned repository: %w", err)
	}

	if len(skills) == 0 && source.SkillFilter != "" {
		return nil, fmt.Errorf("skill %q not found in %s\n→ Use 'sl skill add %s' to see available skills",
			source.SkillFilter, source.SourceString(), source.SourceString())
	}
	if len(skills) == 0 {
		return nil, fmt.Errorf("no skills found in repository\n→ The repository may not contain SKILL.md files")
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
