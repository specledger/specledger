# SpecLedger Constitution

## Core Principles

### I. YAGNI (You Aren't Gonna Need It)
Don't build abstractions, features, or configuration for hypothetical future needs. Build what's needed now. If a requirement isn't on the current task list, it doesn't exist yet. Three similar lines of code are better than a premature abstraction.

### II. DRY (Don't Repeat Yourself)
Extract shared logic only when duplication is proven — apply the rule of three. Premature abstraction is worse than duplication. When you do extract, the abstraction must earn its keep by being clearer than the inlined version.

### III. Shortest Path to MVP
For every feature, find the minimum viable implementation. Ship, get feedback, iterate. No gold-plating, no "while we're here" scope creep. The right amount of complexity is the minimum needed for the current task.

### IV. Short-Lived Feature Branches
Branches should be small, focused, and merged quickly. Long-lived branches accumulate merge debt and stale context. Use Supabase data branches to pair DB schema changes with code branches, keeping both in sync and short-lived.

### V. Simplicity Over Cleverness
Prefer boring, obvious code over clever solutions. If a reviewer needs to pause to understand it, simplify it. Optimize for readability and maintainability, not elegance.

### VI. Contract-First Testing
Supabase/pgREST API contracts are the source of truth for integration boundaries. Contracts are snapshotted on disk and validated in tests. Any migration or schema change must update the contract snapshot. Three testing tiers enforce this — see [`tests/README.md`](../../tests/README.md) for the full testing strategy and current implementation status.

**Testing tiers:**
- **Unit:** Mock HTTP (`httptest`) + go-vcr cassettes for fast, deterministic tests
- **Integration:** Real `supabase start` local stack in `tests/integration/` — validates actual pgREST behavior, RLS policies, and migrations
- **E2E:** Full CLI binary invocation against local Supabase stack, orchestrated entirely via Go's `testing` package — no GUI, pure CLI validation

### VII. Supabase Local Stack as Test Infrastructure
Every feature branch maintains a working local Supabase environment (`supabase start`). Schema migrations, RLS policies, and edge functions are validated against the local stack in CI. Supabase data branching is used to isolate feature work. The local stack must always be updated as part of any feature or refactoring work — if you change a migration, RLS policy, or edge function, the local stack and its tests must reflect that change.

### VIII. Quickstart-Driven Validation
The `quickstart.md` generated during the planning phase defines user scenarios that map 1:1 to E2E test cases. Plans and task lists must include a phase that translates:

  **Spec user stories/FRs → quickstart.md scenarios → Go e2e test functions**

If a quickstart scenario isn't covered by a test, it's a gap. If a test doesn't trace back to a quickstart scenario, question whether it's needed.

### IX. Fail Fast, Fix Forward
Surface errors early with clear, actionable messages. Don't silently swallow failures. Fix issues with new commits, not by rewriting history. Errors should tell the user what went wrong and suggest what to do next.

## Architecture & CLI Design

The detailed CLI architecture, command patterns, and layer design rules live in [`docs/design/README.md`](../../docs/design/README.md). The design docs are authoritative for:

- **4-layer tooling model** (L0 Hooks, L1 CLI, L2 Commands, L3 Skills)
- **CLI design principles** (Progressive Discovery, Error Messages as Navigation, Two-Level Output)
- **Command pattern classification** (Data CRUD, Launcher, Hook Trigger, Environment, Template Management)
- **Cross-layer interaction rules**

**This separation is intentional:** high-level development principles live here in the constitution; architecture and CLI-specific rules live in the design docs.

## Testing Strategy

The detailed, up-to-date testing implementation lives in [`tests/README.md`](../../tests/README.md). The principles above (VI, VII, VIII) govern the strategy; the README documents the current state, tooling, and conventions.

**This separation is intentional:** principles are stable and live here; implementation details evolve and live with the test code.

## Development Workflow

- Feature branches are short-lived and focused on a single spec
- Every feature goes through: Specify → Clarify → Plan → Tasks → Review → Implement
- Quickstart scenarios are written during planning, before implementation begins
- E2E tests covering quickstart scenarios are part of the task list, not an afterthought
- New CLI commands must be classified into one of the 5 command patterns (see [`docs/design/cli.md` — Pattern Classification](../../docs/design/cli.md)) and task lists must include a verification task for pattern compliance
- Local Supabase stack is validated before pushing (CI enforces this)
- Use `/specledger.commit` for all git operations (auth-aware session capture)

## Agent Preferences

- **Preferred Agent**: Claude Code

## Governance

This constitution supersedes all other practices. Amendments require:
1. Documentation of the change and rationale
2. Review and approval
3. Update to this file and any affected artifacts (including [`tests/README.md`](../../tests/README.md))

All PRs and reviews must verify compliance with these principles. Complexity must be justified against Principles I, III, and V.

**Version**: 1.0.0 | **Ratified**: 2026-03-16 | **Last Amended**: 2026-03-16
