# Tasks Index: Silent Session Capture

Issue graph index for the 598-silent-session-capture feature implementation.
This index does **not contain tasks directly** - those are managed through `sl issue`.

## Feature Tracking

* **Epic ID**: `SL-613d70`
* **User Stories Source**: `specledger/598-silent-session-capture/spec.md`
* **Research Inputs**: `specledger/598-silent-session-capture/research.md`
* **Planning Details**: `specledger/598-silent-session-capture/plan.md`
* **Verification**: `specledger/598-silent-session-capture/quickstart.md`

## Query Hints

```bash
# All issues for this feature
sl issue list --label "spec:598-silent-session-capture"

# Open tasks only
sl issue list --label "spec:598-silent-session-capture" --status open

# Show full epic tree
sl issue show SL-613d70 --tree
```

## Tasks and Phases Structure

```
Epic: SL-613d70 (Silent Session Capture)
├── Feature: SL-4efce1 (Foundational: Reorder Capture flow)
│   └── Task: SL-04eba6 (Move LoadCredentials before GetProjectIDWithFallback)
├── Feature: SL-d3e926 (US1: Silent skip when no credentials)
│   └── Task: SL-c0be57 (Return silently when no credentials exist)
├── Feature: SL-395c8c (US2: Silent skip when no project ID)
│   └── Task: SL-2cd79f (Remove stderr warnings for missing project ID)
├── Feature: SL-ca36b5 (US3: Error logging only on real upload failures)
│   └── Task: SL-bc61e9 (Verify upload failure error paths unchanged)
└── Feature: SL-e62151 (US4: No regression in happy path)
    └── Task: SL-bdb05d (Add unit tests for silent skip behavior)
```

## Dependency Graph

```
SL-04eba6 (Reorder)
├── blocks → SL-c0be57 (Silent no-creds)
└── blocks → SL-2cd79f (Silent no-project)
                 │               │
                 └───────┬───────┘
                         ▼
                 SL-bc61e9 (Verify error paths)
                         │
                         ▼
                 SL-bdb05d (Add tests)
```

## Execution Order

1. **SL-04eba6**: Reorder Capture() - move credentials check first
2. **SL-c0be57 + SL-2cd79f**: (parallel) Silent returns for no-creds and no-project
3. **SL-bc61e9**: Verify real error paths unchanged
4. **SL-bdb05d**: Add unit tests for all silent skip scenarios

## Definition of Done Summary

| Issue ID   | DoD Items |
|------------|-----------|
| SL-04eba6  | - Code reordered in Capture()<br>- go vet passes<br>- go build succeeds |
| SL-c0be57  | - No error set for missing credentials<br>- No stderr output from capture function |
| SL-2cd79f  | - Three stderr warning lines removed<br>- No error set for missing project ID<br>- result.Error remains nil |
| SL-bc61e9  | - Storage upload failure still queues and reports<br>- Metadata creation failure still queues and reports<br>- Token refresh failure still queues and reports |
| SL-bdb05d  | - Test: silent skip when no credentials<br>- Test: silent skip when no project ID<br>- Test: silent skip for non-commit commands<br>- Test: silent skip for failed commits<br>- All tests pass with go test |

## Implementation Strategy

**MVP**: SL-04eba6 + SL-c0be57 + SL-2cd79f (Foundational + US1 + US2)
This alone solves the core bug - no more spam warnings for unauthenticated/unsynced users.

**Incremental delivery**:
1. MVP (above) - fixes the reported issue
2. SL-bc61e9 - verification that error paths still work
3. SL-bdb05d - test coverage for the new behavior
