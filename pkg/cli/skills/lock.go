package skills

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
)

// LocalSkillLockEntry is a single entry in skills-lock.json.
// Schema matches official Vercel format.
type LocalSkillLockEntry struct {
	Source       string `json:"source"`
	Ref          string `json:"ref,omitempty"`
	SourceType   string `json:"sourceType"`
	ComputedHash string `json:"computedHash"`
}

// LocalSkillLockFile is the skills-lock.json file. Matches official Vercel schema v1.
type LocalSkillLockFile struct {
	Version int                            `json:"version"`
	Skills  map[string]LocalSkillLockEntry `json:"skills"`
}

// ReadLocalLock reads and parses skills-lock.json.
// Returns an empty lock file struct if the file does not exist.
// Returns an error on invalid JSON.
func ReadLocalLock(path string) (*LocalSkillLockFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &LocalSkillLockFile{
				Version: 1,
				Skills:  make(map[string]LocalSkillLockEntry),
			}, nil
		}
		return nil, fmt.Errorf("failed to read skills-lock.json: %w", err)
	}

	if len(data) == 0 {
		return &LocalSkillLockFile{
			Version: 1,
			Skills:  make(map[string]LocalSkillLockEntry),
		}, nil
	}

	var lock LocalSkillLockFile
	if err := json.Unmarshal(data, &lock); err != nil {
		return nil, fmt.Errorf("skills-lock.json is invalid.\n→ Fix the JSON syntax or delete skills-lock.json to start fresh")
	}

	if lock.Skills == nil {
		lock.Skills = make(map[string]LocalSkillLockEntry)
	}

	return &lock, nil
}

// WriteLocalLock writes the lock file with skills sorted alphabetically,
// 2-space indent, and trailing newline.
func WriteLocalLock(path string, lock *LocalSkillLockFile) error {
	if lock.Version == 0 {
		lock.Version = 1
	}

	// Sort skills alphabetically by creating an ordered structure
	ordered := &orderedLockFile{
		Version: lock.Version,
		Skills:  sortedSkillMap(lock.Skills),
	}

	data, err := json.MarshalIndent(ordered, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal skills-lock.json: %w", err)
	}

	// Append trailing newline
	data = append(data, '\n')

	// #nosec G306 -- lock file needs to be readable, 0644 is appropriate
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write skills-lock.json: %w", err)
	}

	return nil
}

// AddSkill adds or updates a skill entry in the lock file.
func AddSkill(lock *LocalSkillLockFile, name string, entry LocalSkillLockEntry) {
	if lock.Skills == nil {
		lock.Skills = make(map[string]LocalSkillLockEntry)
	}
	lock.Skills[name] = entry
}

// RemoveSkill removes a skill entry from the lock file.
func RemoveSkill(lock *LocalSkillLockFile, name string) {
	delete(lock.Skills, name)
}

// orderedLockFile ensures JSON output has skills in alphabetical order.
type orderedLockFile struct {
	Version int            `json:"version"`
	Skills  sortedSkillMap `json:"skills"`
}

// sortedSkillMap implements json.Marshaler to produce sorted keys.
type sortedSkillMap map[string]LocalSkillLockEntry

func (m sortedSkillMap) MarshalJSON() ([]byte, error) {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build ordered JSON manually
	buf := []byte("{")
	for i, k := range keys {
		if i > 0 {
			buf = append(buf, ',')
		}
		keyJSON, err := json.Marshal(k)
		if err != nil {
			return nil, err
		}
		valJSON, err := json.Marshal(m[k])
		if err != nil {
			return nil, err
		}
		buf = append(buf, keyJSON...)
		buf = append(buf, ':')
		buf = append(buf, valJSON...)
	}
	buf = append(buf, '}')
	return buf, nil
}
