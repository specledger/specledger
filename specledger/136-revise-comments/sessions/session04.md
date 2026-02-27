# Session 04 — Thread Support & Comment Grouping
**Date**: 2026-02-21
**Branch**: `597-agent-model-config` (hot-fix, applies to 136-revise-comments feature)
**Agent**: claude-opus-4-6
**Scope**: Add thread reply support to `sl revise` + comment grouping strategy (GitHub issue #34)

---

## Context

The `review_comments` table already supports threading via a self-referential `parent_comment_id` FK, but the revise command explicitly filtered these out with `parent_comment_id=is.null`. Thread replies containing important clarifications (e.g., "this also affects user story 2") were invisible to the revision agent.

Additionally, GitHub issue [#34](https://github.com/specledger/specledger/issues/34) requested that the revision prompt instruct the agent to collate related comments by theme rather than processing them one-by-one.

## What Was Built

### Modified Files

| File | Change |
|------|--------|
| `pkg/cli/revise/types.go` | Added `ThreadReply` struct; added `Replies []ThreadReply` to `PromptComment` |
| `pkg/cli/revise/client.go` | Added `FetchReplies()`, `BuildReplyMap()`, `ResolveCommentWithReplies()` |
| `pkg/cli/revise/prompt.go` | Updated `BuildRevisionContext` to accept replies and populate thread data |
| `pkg/cli/revise/prompt.tmpl` | Added inline thread display; rewrote instructions with Revision Strategy for thematic clustering |
| `pkg/cli/commands/revise.go` | Wired threads into `runRevise`, `processComments` (TUI display), `runSummary` (reply counts), `runAuto`, `commentResolutionFlow` (cascade resolve) |
| `pkg/cli/revise/prompt_test.go` | Added 4 new tests: `WithReplies`, `WithThreads`, `BuildReplyMap`, `BuildReplyMap_Nil` |
| `pkg/cli/revise/automation_test.go` | Updated `BuildRevisionContext` call signature |
| `pkg/cli/revise/testdata/snapshot_prompt.golden` | Regenerated for new template |

### Design Decisions

1. **Separate fetch for replies** — `FetchReplies()` queries `parent_comment_id=not.is.null` in a separate call rather than modifying the existing `FetchComments()` query. This keeps the existing API stable and makes thread fetching non-fatal (graceful degradation).

2. **Client-side thread assembly** — `BuildReplyMap()` creates a `map[parentID][]Reply` that is passed through the pipeline. This avoids PostgREST join complexity and keeps the data model flat.

3. **Inline thread display in prompt** — Replies appear as blockquotes under their parent comment with author attribution. The agent sees full discussion context.

4. **Thematic clustering (issue #34)** — The prompt now instructs the agent to: (a) read ALL comments first, (b) identify thematic clusters, (c) present a single AskUserQuestion per cluster with 2-3 approaches, (d) apply coordinated edits across all impacted artifacts.

5. **Cascade resolution** — Resolving a parent comment auto-resolves its thread replies via a single batch PATCH using PostgREST's `id=in.(...)` filter.

6. **TUI thread display** — Thread replies appear in the interactive `processComments` loop with `└─ author: content` formatting so users have full context when deciding to Process/Skip.

### Verification

- `make build` — compiles cleanly
- `make test` — all tests pass (18 in revise package, 60+ total)
- `gofmt -l .` — no formatting issues
- `golangci-lint` — no new warnings (all pre-existing)

### Addresses

- GitHub issue [#34](https://github.com/specledger/specledger/issues/34): revise prompt should tell the agent to collate review comments
- Thread support for `review_comments.parent_comment_id`
