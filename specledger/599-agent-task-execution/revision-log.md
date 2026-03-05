# Revision Log: 599-agent-task-execution/spec.md

## Revision Session 2026-03-05

### Cluster A — Task Hierarchy & Execution Model (Comments 1, 7, 8)
**Options Presented:**
1. Full hierarchy rewrite (Recommended) - Reprioritize cloud execution to P1, hierarchical task model with sub-minions
2. Incremental hierarchy addition - Keep priorities, add hierarchy notes
3. Swap P1/P3 priorities only - Minimal reordering

**Choice:** Option 1 — Full hierarchy rewrite

**Changes Applied:**
- User Story 1 rewritten as "Cloud-Triggered Agent Execution with Sub-Minions" (P1) — auto-scheduled on spec plan approval, main tasks break into sub-tasks, sub-minions execute in parallel, sub-tasks auto-merge, main tasks need human review
- User Story 3 (formerly Cloud Agent Execution) replaced by new User Story 1
- Clarification Q1 updated to reflect hierarchical branching model
- Clarification Q2 updated to reflect sub-task auto-merge and main-task-only human review
- Edge cases updated for sub-task/main-task review rules, sub-task completion, and timeout
- FR-015 rewritten for hierarchical branching
- FR-016, FR-017 updated for main-task-only human review
- FR-018 added for cloud-triggered scheduling
- Key Entities updated: added Main Task, Sub-Task, Sub-Minion
- SC-001 updated for cloud-triggered scheduling
- SC-007 added for sub-task auto-merge rate
- Assumptions updated for cloud-first workflow and hierarchical branching

### Cluster B — Local Execution Value (Comment 2)
**Options Presented:**
1. Demote to P3, clarify niche (Recommended)
2. Remove local execution entirely
3. Keep as P2, add differentiation

**Choice:** Option 1 — Demote to P3, clarify niche

**Changes Applied:**
- Former User Story 1 (Local Agent Task Pickup) moved to User Story 3 (P3)
- Description updated to clarify it's a development/testing convenience, not the primary execution path
- Added clarification Q&A about local vs /specledger.implement differentiation

### Cluster C — Observability & Configuration (Comments 3, 4)
**Options Presented:**
1. Logging platform + full creds config (Recommended)
2. Dual monitoring + basic creds
3. Observability-first, defer creds

**Choice:** Option 1 — Logging platform + full creds config

**Changes Applied:**
- User Story 2 rewritten as "Execution Observability and Metrics" — emit to Splunk/Sentry instead of CLI polling
- `sl agent status` CLI monitoring replaced with observability platform emission
- FR-007 rewritten for observability platform emission
- FR-021 added for git credentials configuration
- FR-022 added for specledger access token scoping (owner-scoped vs master credential)
- User Story 4 updated to include git credentials, access tokens, and timeout configuration
- SC-008 added for observability metric emission latency
- Assumptions updated to include observability platform availability

### Cluster D — Execution Resilience (Comments 5, 6)
**Options Presented:**
1. Timeout + mock/stub resolution (Recommended)
2. Timeout only, keep cycle skip
3. Both with phased rollout

**Choice:** Option 1 — Timeout + mock/stub resolution

**Changes Applied:**
- Edge case for circular dependencies rewritten: mock/stub approach replaces skip-both
- Edge case for crashes updated to include timeout threshold detection
- New edge case added for task timeout
- FR-019 added for configurable task timeout limits
- FR-020 added for circular dependency resolution via mock/stub approach
- User Story 1 acceptance scenario 5 added for timeout handling
- User Story 4 acceptance scenario 5 added for timeout configuration

### Comment 9 (rate-limiting acknowledgment)
No changes needed — "yes" confirms existing behavior is acceptable.
