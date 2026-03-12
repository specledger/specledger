# Research: Gitattributes Merge

## Prior Work

- **GitHub Issue #74**: PRs cluttered with auto-generated files — originating bug
- **PR #70**: Example of cluttered PR (review comment r2921999907)
- **135-fix-missing-chmod-x**: Previous template handling work (executable permissions in copy.go)
- **pkg/cli/context/updater.go**: Existing sentinel marker pattern for CLAUDE.md updates

## Key Finding: Existing Sentinel Pattern in Codebase

The codebase already implements sentinel-based content preservation in `pkg/cli/context/updater.go`:

- **Markers**: `<!-- MANUAL ADDITIONS START -->` / `<!-- MANUAL ADDITIONS END -->`
- **Logic**: `extractManualAdditions()` uses `strings.Index()` to find markers, extracts content between them
- **Handling of missing markers**: Returns empty string (lines 106-115)
- **Handling of malformed markers (start >= end)**: Returns empty string
- **Atomic writes**: Uses temp file + rename pattern (`writeFile()` lines 84-103)

**Key difference from our use case**: The updater.go pattern preserves the *user's* manual additions while regenerating everything else. Our `.gitattributes` feature is the inverse — we preserve the *user's* content while managing a *specledger* section. But the string manipulation pattern is identical.

### Decision: No External Libraries Needed

- The codebase uses only Go standard library for all file manipulation (`strings`, `os`, `bytes`)
- No third-party merge/patch libraries are imported
- The sentinel pattern in `updater.go` is ~10 lines of code — simple enough to adapt
- Standard library `strings.Index()` + string slicing is sufficient

### Rationale

- Keeps dependency footprint minimal (Go standard library only)
- Follows existing codebase patterns (consistency)
- The merge logic is simple enough that a library would be over-engineering
- Table-driven tests provide sufficient coverage for edge cases

### Alternatives Considered

| Alternative | Rejected Because |
|-------------|-----------------|
| External merge library (e.g., go-diff) | Overkill for sentinel block replacement; adds dependency |
| Regex-based matching | `strings.Index()` is simpler, faster, and already proven in codebase |
| Line-by-line scanner (bufio) | Not needed — sentinel blocks are found by index, not line-by-line |
| init.sh-based merging | Would require shell scripting, less testable, cross-platform concerns |

## Playbook Copy System Analysis

**Current flow** (`pkg/cli/playbooks/copy.go`):
1. `CopyPlaybooks()` → orchestrates structure items, commands, skills
2. `copyStructureItem()` → checks protected, then delegates to file/directory copy
3. `copySingleFileFromContent()` → checks `Overwrite` flag, writes with permissions

**Integration point for merge**: `copyStructureItem()` (line 90) — add a check for mergeable files before the protected file check. Mergeable files should bypass both protected and overwrite logic entirely, since merge always reads + writes.

**`sl doctor --template` path**: `pkg/templates/updater.go` → `UpdateTemplates()` calls `source.Copy()` with `Overwrite: true`, which flows through the same `CopyPlaybooks` → `copyStructureItem` pipeline. This means our merge logic will automatically work for both `sl init` and `sl doctor --template` — no separate implementation needed.

**Manifest types** (`pkg/cli/playbooks/template.go`):
- `Playbook` struct already has `Protected []string` — add `Mergeable []string` in the same pattern
- `CopyResult` struct has `FilesCopied` and `FilesSkipped` — add `FilesMerged`

## Sentinel Block Design

Based on the existing `updater.go` pattern, adapted for `.gitattributes`:

```
# >>> specledger-generated
# Auto-managed by specledger - do not edit this section
specledger/*/issues.jsonl linguist-generated=true
specledger/*/tasks.md linguist-generated=true
# <<< specledger-generated
```

**Merge algorithm** (adapted from `extractManualAdditions()`):
1. Find `SentinelBegin` index in existing content
2. Find `SentinelEnd` index in existing content
3. **Both found (begin < end)**: Replace from begin to end+len(SentinelEnd) with new block
4. **Begin found, no end (malformed)**: Replace from begin to EOF with new block (FR-011)
5. **Neither found**: Append new block to existing content
6. Ensure single trailing newline

**Idempotency**: Same input always produces same output because:
- The sentinel block is deterministic (same template content)
- User content outside sentinels is untouched
- No timestamp or variable content in the managed section

## Spike: Sentinel Block Libraries & Industry Patterns

### Industry Precedent: Conda's `conda init`

Conda uses the exact same sentinel pattern we're implementing. When you run `conda init`, it adds a managed block to `~/.bashrc`:

```bash
# >>> conda initialize >>>
# !! Contents within this block are managed by 'conda init' !!
__conda_setup="$('/path/to/conda' 'shell.bash' 'hook' 2> /dev/null)"
...
# <<< conda initialize <<<
```

Key design choices from conda:
- **Marker format**: `# >>> conda initialize >>>` / `# <<< conda initialize <<<` (we use `# >>> specledger-generated` / `# <<< specledger-generated`)
- **Warning comment**: Includes `!! Contents within this block are managed by 'conda init' !!` — we include `# Auto-managed by specledger - do not edit this section`
- **Idempotent**: Re-running `conda init` replaces the block without duplicating
- **User content preserved**: Everything outside the markers is untouched

Our design directly follows this proven pattern.

Sources: [conda init deep-dive](https://docs.conda.io/projects/conda/en/stable/dev-guide/deep-dives/activation.html), [Baeldung: conda in bashrc](https://www.baeldung.com/linux/bashrc-activate-conda-environment)

### Go Libraries Evaluated

No Go-specific libraries exist for sentinel-based file section management. The search for "golang sentinel block merge" returns only:
- **alibaba/sentinel-golang**: Circuit breaker/rate limiter (unrelated)
- **hashicorp/consul/sentinel**: Policy engine (unrelated)
- **Go sentinel errors**: Error handling pattern (unrelated)

**Conclusion**: This is a ~20-line function using Go standard library (`strings.Index`, string slicing). No external dependency warranted.

### Existing Codebase Merge Logic Summary

| Location | Pattern | Reusable? |
|----------|---------|-----------|
| `pkg/cli/context/updater.go` | Sentinel markers (`<!-- MANUAL ADDITIONS -->`) for CLAUDE.md | **Yes** — same `strings.Index()` approach, adapt markers |
| `pkg/cli/config/merge.go` | Layered config merge (last-write-wins) | No — config merging, not file content |
| `pkg/cli/playbooks/copy.go` | Protected files map + overwrite/skip flags | **Yes** — add `mergeableMap` in same pattern |
| `init.sh` | Post-init script (bash) | No — runs after copy, doesn't merge files |

### Recommendation

Implement `MergeSentinelSection()` as a pure function in `pkg/cli/playbooks/merge.go` using the same `strings.Index()` approach from `updater.go`. No external libraries needed. The conda pattern validates our design choice.
