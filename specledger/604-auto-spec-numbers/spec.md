# Feature Specification: Auto-Generate Spec Numbers

**Feature Branch**: `604-auto-spec-numbers`
**Created**: 2026-03-10
**Status**: Draft
**Input**: GitHub Issue #66 - Auto-generate spec numbers in sl spec create (remove --number flag)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Create Feature Without Specifying a Number (Priority: P1)

A developer wants to start a new feature. They run `sl spec create --short-name "user-auth"` without needing to know what feature numbers already exist. The system automatically assigns the next available number, creates the branch, and sets up the spec directory.

**Why this priority**: This is the core value proposition. Manual number selection is the primary friction point causing collisions and workflow interruption.

**Independent Test**: Can be fully tested by running `sl spec create --short-name "test"` in a repo with existing features and verifying the assigned number doesn't collide with any existing feature.

**Acceptance Scenarios**:

1. **Given** a repo with features 001 through 005, **When** user runs `sl spec create --short-name "new-feature"`, **Then** the system assigns number 006 and creates branch `006-new-feature`
2. **Given** a repo with no existing features, **When** user runs `sl spec create --short-name "first-feature"`, **Then** the system assigns number 001
3. **Given** a repo where number 006 exists as a remote branch but not locally, **When** user runs `sl spec create --short-name "new-feature"`, **Then** the system skips 006 and assigns 007

---

### User Story 2 - Manual Number Override (Priority: P2)

A developer has a specific number in mind (e.g., matching an external issue tracker) and wants to use it. They run `sl spec create --number 42 --short-name "my-feature"`. The system validates the number is available and uses it.

**Why this priority**: Supports backward compatibility and workflows where numbers are coordinated externally.

**Independent Test**: Can be tested by running `sl spec create --number 42 --short-name "test"` and verifying the number is used as-is.

**Acceptance Scenarios**:

1. **Given** number 42 is not in use, **When** user runs `sl spec create --number 42 --short-name "my-feature"`, **Then** branch `042-my-feature` is created
2. **Given** number 42 already has a local directory, **When** user runs `sl spec create --number 42 --short-name "other-feature"`, **Then** an error is returned with the collision details and a suggested available number

---

### User Story 3 - AI Agent Creates Specs Without Manual Number Lookup (Priority: P2)

The `specledger.specify` AI skill creates new features automatically. Previously, the AI agent had to scan directories to determine the next number. Now it simply runs `sl spec create --json --short-name "feature-name"` and the CLI handles numbering.

**Why this priority**: Eliminates a common failure mode where AI agents default to "001" or miscalculate the next number.

**Independent Test**: Can be tested by invoking `/specledger.specify` and verifying the JSON output contains a valid, collision-free FEATURE_NUM.

**Acceptance Scenarios**:

1. **Given** the AI skill invokes `sl spec create --json --short-name "analytics"`, **When** the command completes, **Then** JSON output includes BRANCH_NAME, FEATURE_DIR, SPEC_FILE, FEATURE_NUM, and FEATURE_ID
2. **Given** the AI skill does not pass `--number`, **When** the command runs, **Then** it auto-assigns a collision-free number without AI intervention

---

### Edge Cases

- What happens when 100 consecutive numbers are all taken? System returns a clear error message.
- How does the system handle concurrent feature creation by two users? Each user gets a unique number based on local + remote checks at creation time.
- What happens when remote is unreachable? System falls back to local-only checks and proceeds (best-effort remote check).
- What if the specledger/ directory doesn't exist yet? System starts numbering from 001.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST create features without requiring the `--number` flag
- **FR-002**: System MUST auto-assign the next available number by scanning existing feature directories, local branches, and remote branches
- **FR-003**: System MUST still accept an optional `--number` flag for manual override
- **FR-004**: When a manually specified number collides, the system MUST report the collision and suggest the next available number
- **FR-005**: System MUST include a `FEATURE_ID` field in JSON output (matching the branch name)
- **FR-006**: System MUST support legacy numeric specs (existing NNN-name format remains unchanged)
- **FR-007**: The AI skill documentation MUST be updated to reflect that `--number` is no longer required

### Key Entities

- **Feature Number**: A zero-padded 3+ digit identifier (e.g., "001", "042", "604") assigned to each feature
- **Feature ID**: The full branch name combining number and short name (e.g., "604-auto-spec-numbers")
- **Collision**: When a feature number already exists as a local directory, local branch, or remote branch

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of feature creation commands succeed without manual number input
- **SC-002**: Zero feature number collisions occur during auto-assignment
- **SC-003**: AI skill (`/specledger.specify`) creates features successfully without number-related failures
- **SC-004**: Manual `--number` override continues to work for users who need it
- **SC-005**: Feature creation completes in under 5 seconds including remote branch checks

### Previous work

- **598-sdd-workflow-streamline**: Established the SDD 4-layer model and CLI-first approach
- **600-bash-cli-migration**: Migrated bash scripts to Go CLI, including `sl spec create`
- **601-cli-skills**: Defined AI skill structure including `specledger.specify`

### Assumptions

- Sequential auto-increment (Option A from issue #66) is sufficient; hash-based IDs are not needed for the current project scale
- Remote branch checking is best-effort; network failures should not block feature creation
- The zero-padded 3-digit format (e.g., "001") is maintained for readability and backward compatibility
