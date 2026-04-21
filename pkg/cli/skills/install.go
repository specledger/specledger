package skills

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
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

// validateRelPath checks that a file-relative path within a skill directory
// is safe (no traversal, not absolute, clean).
func validateRelPath(relPath string) error {
	if relPath == "" {
		return fmt.Errorf("empty file path")
	}
	if filepath.IsAbs(relPath) {
		return fmt.Errorf("absolute file path %q", relPath)
	}
	if strings.Contains(relPath, "..") {
		return fmt.Errorf("path traversal in %q", relPath)
	}
	// Normalize to OS separators for the clean check
	osPath := filepath.FromSlash(relPath)
	cleaned := filepath.Clean(osPath)
	if cleaned != osPath {
		return fmt.Errorf("file path %q is not clean (resolves to %q)", relPath, cleaned)
	}
	return nil
}

// maxRetryAfter is the maximum time we'll wait for a Retry-After header before giving up.
const maxRetryAfter = 30 * time.Second

// FetchSkillFiles returns every file under the skill's directory keyed by the
// skill-directory-relative path (e.g. "SKILL.md", "references/core.md").
// Uses the cloned tmp dir when available, otherwise raw.githubusercontent.com.
// All-or-nothing: returns an error (no partial map) on any fetch failure.
func FetchSkillFiles(client *Client, source *SkillSource, skill *SkillMetadata) (map[string][]byte, error) {
	if len(skill.Files) == 0 {
		return nil, fmt.Errorf("skill %q has no files listed from discovery", skill.Name)
	}

	skillDir := strings.TrimSuffix(skill.RepoPath, "/SKILL.md")

	files := make(map[string][]byte, len(skill.Files))
	for _, repoPath := range skill.Files {
		// Derive the skill-relative key
		relKey := strings.TrimPrefix(repoPath, skillDir+"/")
		if err := validateRelPath(relKey); err != nil {
			return nil, fmt.Errorf("unsafe file in skill %q: %w", skill.Name, err)
		}

		var content []byte
		var err error
		if source.cloneDir != "" {
			// Clone path: read from filesystem
			absPath := filepath.Join(source.cloneDir, filepath.FromSlash(repoPath))
			content, err = os.ReadFile(absPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read %s from clone: %w", repoPath, err)
			}
		} else {
			// GitHub path: fetch via raw URL with 429-aware retry
			content, err = fetchWithRetry(client, source, repoPath)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch %s: %w", repoPath, err)
			}
		}

		files[relKey] = content
	}

	if _, ok := files["SKILL.md"]; !ok {
		return nil, fmt.Errorf("skill %q missing SKILL.md in file list", skill.Name)
	}

	return files, nil
}

// fetchWithRetry wraps FetchSkillContent with a single retry on 429/403+Retry-After.
func fetchWithRetry(client *Client, source *SkillSource, repoPath string) ([]byte, error) {
	content, err := client.FetchSkillContent(source.Owner, source.Repo, source.Ref, repoPath)
	if err == nil {
		return content, nil
	}

	// Check if the error is a rate-limit (429 or 403 with Retry-After).
	// FetchSkillContent wraps HTTP errors with the status code in the message.
	// We need a direct HTTP check, so we do a second attempt after waiting.
	// Since FetchSkillContent already consumed the response, we do a lightweight
	// probe to decide whether to retry.
	wait := parseRetryWait(client, source, repoPath)
	if wait <= 0 {
		return nil, err // Not a rate-limit error; return original error
	}

	if wait > maxRetryAfter {
		return nil, fmt.Errorf("rate-limited; Retry-After %v exceeds %v cap\n→ Try again later. If rate-limited, set GITHUB_TOKEN.", wait, maxRetryAfter)
	}

	time.Sleep(wait)
	return client.FetchSkillContent(source.Owner, source.Repo, source.Ref, repoPath)
}

// parseRetryWait does a HEAD request to check if the URL is rate-limited
// and returns the Retry-After duration, or 0 if not rate-limited.
func parseRetryWait(client *Client, source *SkillSource, repoPath string) time.Duration {
	reqURL := fmt.Sprintf("%s/%s/%s/%s/%s",
		client.RawGHURL, source.Owner, source.Repo, source.Ref, repoPath)

	req, err := http.NewRequest(http.MethodHead, reqURL, nil)
	if err != nil {
		return 0
	}

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusTooManyRequests && resp.StatusCode != http.StatusForbidden {
		return 0
	}

	retryAfter := resp.Header.Get("Retry-After")
	if retryAfter == "" {
		// Default fallback for rate-limits without Retry-After
		return 2 * time.Second
	}

	seconds, parseErr := strconv.Atoi(retryAfter)
	if parseErr != nil {
		return 2 * time.Second // unparseable → use default
	}

	return time.Duration(seconds) * time.Second
}

// InstallSkill writes all skill files to each agent's skills directory
// and updates the lock file.
func InstallSkill(name string, files map[string][]byte, agentPaths []string, lockPath string, source *SkillSource) error {
	if err := ValidateSkillName(name); err != nil {
		return fmt.Errorf("unsafe skill name: %w", err)
	}
	if len(files) == 0 {
		return fmt.Errorf("skill %q has no files", name)
	}
	if _, ok := files["SKILL.md"]; !ok {
		return fmt.Errorf("skill %q missing SKILL.md", name)
	}
	if len(agentPaths) == 0 {
		return fmt.Errorf("no agent paths configured\n→ Run 'sl init' to configure agents")
	}

	for _, basePath := range agentPaths {
		skillDir := filepath.Join(basePath, name)
		if err := os.MkdirAll(skillDir, 0755); err != nil {
			return fmt.Errorf("failed to create skill directory %s: %w", skillDir, err)
		}

		for relPath, content := range files {
			// Defense-in-depth: re-validate each relative path
			if err := validateRelPath(relPath); err != nil {
				return fmt.Errorf("unsafe file path in skill %q: %w", name, err)
			}

			destPath := filepath.Join(skillDir, filepath.FromSlash(relPath))

			// Final traversal defense: verify dest is under skillDir
			rel, relErr := filepath.Rel(skillDir, destPath)
			if relErr != nil || strings.HasPrefix(rel, "..") {
				return fmt.Errorf("file path %q escapes skill directory", relPath)
			}

			// Ensure parent directories exist (e.g. references/)
			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				return fmt.Errorf("failed to create directory for %s: %w", destPath, err)
			}

			// #nosec G306 -- skill files need to be readable, 0644 is appropriate
			if err := os.WriteFile(destPath, content, 0644); err != nil {
				return fmt.Errorf("failed to write %s: %w", destPath, err)
			}
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
