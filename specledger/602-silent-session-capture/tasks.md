# Tasks Index: Silent Session Capture

SpecLedger Issue Graph Index for the 602-silent-session-capture feature implementation.
This index does **not contain tasks directly**—those are fully managed through SpecLedger CLI.

## Feature Tracking

* **Epic ID**: `SL-482b2c`
* **User Stories Source**: `specledger/602-silent-session-capture/spec.md`
* **Research Inputs**: `specledger/602-silent-session-capture/research.md`
* **Planning Details**: `specledger/602-silent-session-capture/plan.md`
* **Data Model**: `specledger/602-silent-session-capture/data-model.md`
* **Contract Definitions**: `specledger/602-silent-session-capture/contracts/`

## Query Hints (v2)

```bash
# All issues for this feature
sl issue list --label "spec:602-silent-session-capture"

# Open tasks only
sl issue list --label "spec:602-silent-session-capture" --status open

# By phase
sl issue list --label "spec:602-silent-session-capture" --label "phase:foundational"
sl issue list --label "spec:602-silent-session-capture" --label "phase:us1-us2"
sl issue list --label "spec:602-silent-session-capture" --label "phase:us4"
sl issue list --label "spec:602-silent-session-capture" --label "phase:us3"
sl issue list --label "spec:602-silent-session-capture" --label "phase:polish"

# By user story
sl issue list --label "spec:602-silent-session-capture" --label "story:US1"
sl issue list --label "spec:602-silent-session-capture" --label "story:US3"
sl issue list --label "spec:602-silent-session-capture" --label "story:US4"
```

## Phases and Structure

```
Epic: SL-482b2c (Silent Session Capture)
│
├── Foundational: SL-e1b274 (Error Logging Infrastructure) [P1]
│   ├── SL-b59b26: Create errorlog.go with LogCaptureError (local JSONL + Sentry)
│   └── SL-60611c: Add sentry-go dependency + init in CLI entrypoint
│
├── US1+US2: SL-781410 (Slash Command + Agent Integration) [P1]
│   └── SL-52a737: Create specledger.commit.md slash command
│
├── US4: SL-cf96c4 (Silent Skip in PostToolUse Hook) [P2]
│   ├── SL-2a86fe: Reorder auth check in Capture() for silent skip
│   └── SL-afe557: Update capture_test.go for silent skip behavior
│
├── US3: SL-63afd5 (Error Logging Integration) [P2]
│   │   ⚠️ Blocked by: Foundational (SL-e1b274) + US4 (SL-cf96c4)
│   ├── SL-95d0ac: Integrate LogCaptureError into capture.go
│   └── SL-b04e86: Integrate LogCaptureError into queue.go
│
└── Polish: SL-9260d3 (Validation & Cross-Cutting) [P3]
        ⚠️ Blocked by: US1+US2, US3, US4
    └── SL-a323f4: Run quickstart.md validation tests
```

## Dependency Graph

```
                    ┌─────────────────┐
                    │   Foundational   │
                    │    SL-e1b274     │
                    │ errorlog.go +    │
                    │ Sentry setup     │
                    └────────┬────────┘
                             │ blocks
    ┌────────────────┐       │       ┌────────────────┐
    │   US1+US2 (P1) │       │       │    US4 (P2)    │
    │   SL-781410    │       │       │   SL-cf96c4    │
    │ slash command  │       │       │  silent skip   │
    └───────┬────────┘       │       └───────┬────────┘
            │                │ blocks        │ blocks
            │                ▼               │
            │       ┌────────────────┐       │
            │       │    US3 (P2)    │◄──────┘
            │       │   SL-63afd5    │
            │       │ error logging  │
            │       │  integration   │
            │       └───────┬────────┘
            │               │
            │ blocks        │ blocks
            ▼               ▼
    ┌───────────────────────────────┐
    │        Polish (P3)            │
    │         SL-9260d3             │
    │   quickstart validation       │
    └───────────────────────────────┘
```

### Parallel Execution Opportunities

| Wave | Issues | Can Run In Parallel |
|------|--------|---------------------|
| 1 | Foundational (SL-e1b274), US1+US2 (SL-781410), US4 (SL-cf96c4) | Yes - all independent |
| 2 | US3 (SL-63afd5) | After wave 1 completes |
| 3 | Polish (SL-9260d3) | After all implementation |

Within phases:
- **Foundational**: SL-b59b26 (errorlog.go) and SL-60611c (Sentry setup) are parallel
- **US4**: SL-2a86fe blocks SL-afe557 (reorder before tests)
- **US3**: SL-95d0ac (capture.go) and SL-b04e86 (queue.go) are parallel (after errorlog.go)

## Definition of Done Summary

| Issue ID | DoD Items |
|----------|-----------|
| SL-b59b26 | - CaptureErrorEntry struct defined<br>- LogCaptureError writes JSONL to local file<br>- LogCaptureError sends to Sentry with context tags<br>- Local write before Sentry<br>- Never panics or blocks |
| SL-60611c | - `sentry-go` added to go.mod<br>- Sentry initialized in CLI entrypoint<br>- DSN configurable via env var or build-time embed<br>- `sentry.Flush()` on exit |
| SL-52a737 | - YAML frontmatter with description<br>- Staged check workflow<br>- Commit message from $ARGUMENTS or generated<br>- Auth check with silent skip<br>- Push always proceeds<br>- Summary shows capture status |
| SL-2a86fe | - LoadCredentials moved before project ID<br>- Silent return when no credentials<br>- Silent return when no project ID<br>- Stderr warnings removed<br>- capture_test.go updated |
| SL-afe557 | - Test no credentials silent skip<br>- Test no project ID silent skip<br>- Test invalid credentials JSON<br>- All existing tests passing |
| SL-95d0ac | - LogCaptureError on upload failure<br>- LogCaptureError on metadata failure<br>- All error fields populated<br>- Non-blocking |
| SL-b04e86 | - LogCaptureError on retry failure<br>- Retry count included<br>- Queue processing not blocked |
| SL-a323f4 | - 6 quickstart tests verified<br>- make test passes |

## Implementation Strategy

### MVP Scope (Suggested)

**US1+US2 (P1)**: The slash command alone delivers immediate value - users get a controlled commit workflow. Can be deployed independently since it's just a markdown file.

### Incremental Delivery

1. **Wave 1** (parallel): Slash command + Silent skip + Error logging module
2. **Wave 2**: Error logging integration (requires wave 1)
3. **Wave 3**: End-to-end validation

### Key Files Modified/Created

| File | Action | Phase |
|------|--------|-------|
| `pkg/embedded/skills/commands/specledger.commit.md` | Create | US1+US2 |
| `pkg/cli/session/errorlog.go` | Create | Foundational |
| `pkg/cli/session/capture.go` | Modify | US4 + US3 |
| `pkg/cli/session/capture_test.go` | Modify | US4 |
| `pkg/cli/session/queue.go` | Modify | US3 |

---

> This file is an index only. Implementation data lives in SpecLedger issues. Use the query hints above to navigate.
