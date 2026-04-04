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

// SentinelMarkers holds the begin/end/comment strings for sentinel-based merge.
type SentinelMarkers struct {
	Begin   string
	End     string
	Comment string
}

// HashMarkers uses #-prefixed sentinels (for .gitattributes, TOML, etc.).
var HashMarkers = SentinelMarkers{
	Begin:   SentinelBegin,
	End:     SentinelEnd,
	Comment: SentinelComment,
}

// HTMLMarkers uses HTML comment sentinels (for Markdown files).
var HTMLMarkers = SentinelMarkers{
	Begin:   "<!-- >>> specledger-generated -->",
	End:     "<!-- <<< specledger-generated -->",
	Comment: "<!-- Auto-managed by specledger - do not edit this section -->",
}

// MergeSentinelSection merges managed content into an existing file using #-based sentinel markers.
// It is a convenience wrapper around MergeSentinelSectionWithMarkers using HashMarkers.
func MergeSentinelSection(existing, managed string) string {
	return MergeSentinelSectionWithMarkers(existing, managed, HashMarkers)
}

// MergeSentinelSectionWithMarkers merges managed content into an existing file using configurable sentinel markers.
// It handles four states:
//  1. Empty existing content: returns the sentinel block
//  2. No sentinels found: appends the sentinel block
//  3. Valid sentinels (begin < end): replaces the sentinel section
//  4. Malformed sentinel (begin without end): replaces from begin to EOF
//
// The function is idempotent — calling it twice with the same inputs produces the same output.
func MergeSentinelSectionWithMarkers(existing, managed string, markers SentinelMarkers) string {
	block := buildSentinelBlockWithMarkers(managed, markers)

	// State 1: empty existing content
	if strings.TrimSpace(existing) == "" {
		return block + "\n"
	}

	beginIdx := strings.Index(existing, markers.Begin)
	endIdx := strings.Index(existing, markers.End)

	// State 3: both markers found and valid
	if beginIdx != -1 && endIdx != -1 && beginIdx < endIdx {
		before := existing[:beginIdx]
		after := strings.TrimSpace(existing[endIdx+len(markers.End):])
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

// buildSentinelBlockWithMarkers constructs the full sentinel block with configurable markers.
func buildSentinelBlockWithMarkers(managed string, markers SentinelMarkers) string {
	return markers.Begin + "\n" + markers.Comment + "\n" + managed + "\n" + markers.End
}
