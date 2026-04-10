package skills

import (
	"fmt"
	"strings"

	cligit "github.com/specledger/specledger/pkg/cli/git"
)

// SourceType indicates how to fetch the skill content.
type SourceType string

const (
	SourceTypeGitHub SourceType = "github"
	SourceTypeGit    SourceType = "git"
)

// SkillSource is a parsed representation of a user-provided skill identifier.
type SkillSource struct {
	Owner       string
	Repo        string
	SkillFilter string // optional: specific skill name from @skill syntax
	Ref         string // git ref; empty means auto-resolve (tries HEAD, main, master)
	Type        SourceType
	URL         string // original URL for git type
}

// SourceString returns the canonical "owner/repo" string.
func (s *SkillSource) SourceString() string {
	return s.Owner + "/" + s.Repo
}

// ParseSource parses a skill source identifier into a structured SkillSource.
//
// Supported formats:
//   - owner/repo           → GitHub shorthand
//   - owner/repo@skill     → GitHub shorthand with skill filter
//   - https://github.com/owner/repo.git → full HTTPS URL
//   - git@github.com:owner/repo.git     → SSH URL
func ParseSource(input string) (*SkillSource, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("empty source identifier\n→ Use format: owner/repo or owner/repo@skill-name\n→ Example: sl skill add vercel-labs/agent-skills@creating-pr")
	}

	// Check for path traversal
	if strings.Contains(input, "..") {
		return nil, fmt.Errorf("invalid source %q: path traversal not allowed\n→ Use format: owner/repo or owner/repo@skill-name", input)
	}

	// Extract skill filter from @ suffix (before URL parsing)
	var skillFilter string
	base := input

	// For full URLs, @ is part of the URL syntax (git@...), so only split on @
	// for shorthand format (no :// and no git@)
	if !isURL(input) {
		if idx := strings.LastIndex(input, "@"); idx > 0 {
			base = input[:idx]
			skillFilter = input[idx+1:]
			if skillFilter == "" {
				return nil, fmt.Errorf("invalid source %q: empty skill name after @\n→ Use format: owner/repo@skill-name", input)
			}
			// Validate skill filter: no slashes, no path traversal
			if strings.Contains(skillFilter, "/") {
				return nil, fmt.Errorf("invalid source %q: skill name must not contain '/'\n→ Use format: owner/repo@skill-name", input)
			}
		}
	}

	// Try shorthand owner/repo first
	if !isURL(base) {
		owner, repo, err := cligit.ParseRepoFlag(base)
		if err != nil {
			return nil, fmt.Errorf("invalid source %q\n→ Use format: owner/repo or owner/repo@skill-name\n→ Example: sl skill add vercel-labs/agent-skills@creating-pr", input)
		}
		return &SkillSource{
			Owner:       owner,
			Repo:        repo,
			SkillFilter: skillFilter,
			Ref:         "", // empty = auto-resolve via HEAD/main/master fallback
			Type:        SourceTypeGitHub,
		}, nil
	}

	// Full URL (HTTPS, SSH)
	owner, repo, err := cligit.ParseRepoURL(base)
	if err != nil {
		return nil, fmt.Errorf("invalid source URL %q\n→ Use format: owner/repo or a full git URL\n→ Example: sl skill add https://github.com/owner/repo.git", input)
	}
	return &SkillSource{
		Owner:       owner,
		Repo:        repo,
		SkillFilter: skillFilter,
		Ref:         "", // empty = auto-resolve; clone uses repo default, GitHub tries HEAD/main/master
		Type:        SourceTypeGit,
		URL:         base,
	}, nil
}

// isURL returns true if the input looks like a full URL (not shorthand).
func isURL(s string) bool {
	return strings.Contains(s, "://") || strings.HasPrefix(s, "git@")
}
