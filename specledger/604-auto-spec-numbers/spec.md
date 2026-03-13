# Feature Specification: Hash-Based Spec IDs

**Feature Branch**: `604-auto-spec-numbers`
**Created**: 2026-03-10
**Status**: In Progress
**Input**: GitHub Issue #66 - Auto-generate spec numbers in sl spec create (remove --number flag)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Create Feature With Auto-Generated Hash (Priority: P1)

A developer wants to start a new feature. They run `sl spec create --short-name "user-auth"`. The system generates a unique 6-character hex hash, creates the branch (e.g., `a3f2b1-user-auth`), and sets up the spec directory.

**Why this priority**: This is the core value proposition. Hash-based IDs eliminate collision risk when multiple people work on the same repo concurrently.

**Independent Test**: Can be fully tested by running `sl spec create --short-name "test"` multiple times and verifying each gets a unique hash with no collisions.

**Acceptance Scenarios**:

1. **Given** any repo state, **When** user runs `sl spec create --short-name "new-feature"`, **Then** the system generates a 6-char hex hash and creates branch `<hash>-new-feature`
2. **Given** a repo with no existing features, **When** user runs `sl spec create --short-name "first-feature"`, **Then** a unique hash is generated
3. **Given** two developers create specs concurrently, **When** both run `sl spec create`, **Then** each gets a unique hash (no coordination needed)

---

### User Story 2 - AI Agent Creates Specs Without Manual Lookup (Priority: P2)

The `specledger.specify` AI skill creates new features automatically. It runs `sl spec create --json --short-name "feature-name"` and the CLI generates a unique hash.

**Why this priority**: Eliminates a common failure mode where AI agents default to "001" or miscalculate the next number.

**Independent Test**: Can be tested by invoking `/specledger.specify` and verifying the JSON output contains a valid, unique FEATURE_HASH.

**Acceptance Scenarios**:

1. **Given** the AI skill invokes `sl spec create --json --short-name "analytics"`, **When** the command completes, **Then** JSON output includes BRANCH_NAME, FEATURE_DIR, SPEC_FILE, FEATURE_HASH, and FEATURE_ID
2. **Given** multiple AI agents run concurrently, **When** each creates a spec, **Then** no collisions occur

---

### User Story 3 - Backward Compatibility With Legacy Numeric Specs (Priority: P2)

Existing specs with numeric prefixes (e.g., `604-auto-spec-numbers`) continue to work. The system recognizes both formats for branch detection, issue context, and spec directory scanning.

**Why this priority**: Must not break existing workflows.

**Acceptance Scenarios**:

1. **Given** a repo with legacy spec `604-auto-spec-numbers`, **When** running `sl issue list`, **Then** it correctly detects the spec context
2. **Given** a mix of legacy numeric and hash-based specs, **When** listing features, **Then** both formats are recognized

---

### Edge Cases

- What happens if a generated hash collides? System retries up to 10 times (probability ~1 in 16 million per attempt).
- What happens when remote is unreachable? System falls back to local-only collision checks (best-effort remote check).
- What if the specledger/ directory doesn't exist yet? System generates a hash and creates the directory.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST generate a random 6-character hex hash for each new feature
- **FR-002**: System MUST verify the hash has no collisions against local directories, local branches, and remote branches
- **FR-003**: The `--number` flag is REMOVED; all new specs use hash-based IDs
- **FR-004**: System MUST include a `FEATURE_HASH` field in JSON output
- **FR-005**: System MUST include a `FEATURE_ID` field in JSON output (matching the branch name)
- **FR-006**: System MUST support legacy numeric specs (existing NNN-name format remains unchanged for detection and context)
- **FR-007**: The AI skill documentation MUST be updated to reflect hash-based IDs

### Key Entities

- **Feature Hash**: A 6-character lowercase hex string (e.g., "a3f2b1") generated from `crypto/rand`
- **Feature ID**: The full branch name combining hash and short name (e.g., "a3f2b1-user-auth")
- **Collision**: When a feature hash already exists as a local directory, local branch, or remote branch

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of feature creation commands succeed without any manual ID input
- **SC-002**: Zero feature hash collisions occur during generation
- **SC-003**: AI skill (`/specledger.specify`) creates features successfully without hash-related failures
- **SC-004**: All existing legacy numeric specs continue to function correctly
- **SC-005**: Feature creation completes in under 5 seconds including remote branch checks

### Previous work

- **598-sdd-workflow-streamline**: Established the SDD 4-layer model and CLI-first approach
- **600-bash-cli-migration**: Migrated bash scripts to Go CLI, including `sl spec create`
- **601-cli-skills**: Defined AI skill structure including `specledger.specify`

### Assumptions

- Hash-based IDs (6 hex chars = 16.7 million possibilities) are sufficient for any practical project scale
- Remote branch checking is best-effort; network failures should not block feature creation
- Legacy numeric format (e.g., "604") is recognized but no longer generated for new specs
