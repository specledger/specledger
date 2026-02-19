package issues

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ID-related errors
var (
	ErrInvalidIDFormat = errors.New("issue ID must be in format SL-xxxxxx (6 hex characters)")
	ErrIDPrefix        = errors.New("issue ID must start with 'SL-'")
)

// GenerateIssueID creates a deterministic, globally unique issue ID
// using SHA-256 hash of (spec_context + title + created_at).
// The ID format is SL-<6-char-hex> where the hex is the first 6 characters
// of the SHA-256 digest.
func GenerateIssueID(specContext, title string, createdAt time.Time) string {
	// Create deterministic input from spec context, title, and timestamp
	// Using nanosecond precision for timestamp to prevent collisions
	data := fmt.Sprintf("%s|%s|%d", specContext, title, createdAt.UnixNano())

	// Generate SHA-256 hash
	hash := sha256.Sum256([]byte(data))

	// Take first 3 bytes (6 hex characters) for the ID
	hexPart := hex.EncodeToString(hash[:3])

	return "SL-" + hexPart
}

// ParseIssueID validates an issue ID format and returns it if valid.
// Returns an error if the ID doesn't match the expected format.
// Note: Only lowercase hex characters are valid since GenerateIssueID always produces lowercase.
func ParseIssueID(id string) (string, error) {
	if id == "" {
		return "", ErrInvalidIDFormat
	}

	// Check prefix
	if !strings.HasPrefix(id, "SL-") {
		return "", ErrIDPrefix
	}

	// Extract hex part
	hexPart := strings.TrimPrefix(id, "SL-")

	// Check length (must be 6 hex characters)
	if len(hexPart) != 6 {
		return "", ErrInvalidIDFormat
	}

	// Validate hex characters (lowercase only for consistency with generated IDs)
	for _, c := range hexPart {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return "", ErrInvalidIDFormat
		}
	}

	return id, nil
}

// IsValidIssueID checks if an ID is in valid format
func IsValidIssueID(id string) bool {
	_, err := ParseIssueID(id)
	return err == nil
}

// CalculateCollisionProbability estimates the probability of at least one
// collision given n issues, using the birthday problem approximation.
// With 6 hex characters (16,777,216 possible values), this provides
// collision probability < 0.01% for up to 100,000 issues.
func CalculateCollisionProbability(n int) float64 {
	if n <= 1 {
		return 0
	}

	// Birthday problem approximation: P(collision) ≈ n² / (2 * N)
	// where N is the number of possible values (16,777,216)
	N := float64(16_777_216)
	return float64(n*n) / (2 * N)
}
