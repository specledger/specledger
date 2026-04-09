package skills

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

// skipDirs lists directory names to skip during hash computation.
var skipDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
}

// ComputeFolderHash computes a deterministic SHA-256 hash of all files in dir.
// Files are sorted by relative path; each file contributes its relative path
// and contents to the running hash. Directories .git and node_modules are skipped.
func ComputeFolderHash(dir string) (string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if skipDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		// Normalize to forward slashes for cross-platform determinism
		files = append(files, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to walk directory %s: %w", dir, err)
	}

	sort.Strings(files)

	h := sha256.New()
	for _, rel := range files {
		// Include the relative path in the hash so renames are detected
		if _, err := io.WriteString(h, rel); err != nil {
			return "", err
		}

		absPath := filepath.Join(dir, filepath.FromSlash(rel))
		f, err := os.Open(absPath)
		if err != nil {
			return "", fmt.Errorf("failed to open %s: %w", rel, err)
		}
		if _, err := io.Copy(h, f); err != nil {
			f.Close()
			return "", fmt.Errorf("failed to read %s: %w", rel, err)
		}
		f.Close()
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
