package playbooks

import (
	"strings"
)

const (
	// SentinelBegin marks the start of the specledger-managed section.
	SentinelBegin = "# >>> specledger-generated"
	// SentinelEnd marks the end of the specledger-managed section.
	SentinelEnd = "# <<< specledger-generated"
	// SentinelComment is the warning comment inside the sentinel block.
	SentinelComment = "# Auto-managed by specledger - do not edit this section"
)

// MergeSentinelSection merges managed content into an existing file using sentinel markers.
// It handles four states:
//  1. Empty existing content: returns the sentinel block
//  2. No sentinels found: appends the sentinel block
//  3. Valid sentinels (begin < end): replaces the sentinel section
//  4. Malformed sentinel (begin without end): replaces from begin to EOF
//
// The function is idempotent — calling it twice with the same inputs produces the same output.
func MergeSentinelSection(existing, managed string) string {
	block := buildSentinelBlock(managed)

	// State 1: empty existing content
	if strings.TrimSpace(existing) == "" {
		return block + "\n"
	}

	beginIdx := strings.Index(existing, SentinelBegin)
	endIdx := strings.Index(existing, SentinelEnd)

	// State 3: both markers found and valid
	if beginIdx != -1 && endIdx != -1 && beginIdx < endIdx {
		before := existing[:beginIdx]
		after := strings.TrimSpace(existing[endIdx+len(SentinelEnd):])
		if after != "" {
			return before + block + "\n\n" + after + "\n"
		}
		return before + block + "\n"
	}

	// State 4: malformed — begin found but no valid end
	if beginIdx != -1 {
		before := existing[:beginIdx]
		return before + block + "\n"
	}

	// State 2: no sentinels — append
	content := strings.TrimRight(existing, "\n")
	return content + "\n\n" + block + "\n"
}

// buildSentinelBlock constructs the full sentinel block with markers and managed content.
func buildSentinelBlock(managed string) string {
	return SentinelBegin + "\n" + SentinelComment + "\n" + managed + "\n" + SentinelEnd
}
