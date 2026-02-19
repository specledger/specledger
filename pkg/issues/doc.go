// Package issues provides a built-in issue tracking system for SpecLedger.
// Issues are stored as JSONL files per spec at specledger/<spec>/issues.jsonl.
//
// Key features:
//   - Globally unique SHA-256 based issue IDs (SL-xxxxxx format)
//   - No daemon required - direct file I/O only
//   - File locking for concurrent access
//   - Migration support from Beads format
//   - Dependency tracking with cycle detection
//   - Definition of Done validation
package issues
