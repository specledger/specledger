---
description: Create feature specification from audit data or branch analysis
---

## User Input

```text
$ARGUMENTS
```

**Input Modes:**
1. **From Audit**: `--module-id [ID] --from-audit` (reads from `scripts/audit-cache.json`)
2. **Manual**: Feature description text (analyzes current branch)

## Purpose

Create or update a feature specification from either:
- **Audit mode**: Use pre-analyzed module data from `/specledger.audit`
- **Manual mode**: Analyze branch commits to understand existing implementation

## Execution Flow

### Step 1: Determine Mode

**If `--from-audit` flag present:**
1. Read `scripts/audit-cache.json`
2. Extract module by `module-id`: `jq '.modules[] | select(.id == "[MODULE_ID]")'`
3. Validate paths exist and code matches audit data
4. If stale: ERROR "Run /specledger.audit --force"
5. Skip to Step 3 (Spec Generation)

**If manual mode:**
1. Spawn Explore Agent to research branch history:
   ```bash
   git rev-parse --abbrev-ref HEAD  # Get branch name
   git fetch origin main
   git merge-base HEAD origin/main  # Find branch-off point (BASE)
   git log --oneline --graph BASE..HEAD  # List commits
   git diff --stat BASE..HEAD  # File changes summary
   ```
2. Summarize changes and prompt user to confirm feature description
3. Continue to Step 2

### Step 2: Create Branch and Spec File

1. Generate short name (2-4 words) from feature description
2. Find highest spec number: check `specledger/[0-9]+-*` directories
3. Run branch creation script:
   ```bash
   .specledger/scripts/bash/adopt-feature-branch.sh --json --number N+1 --short-name "name" "description"
   ```
4. Load `.specledger/templates/spec-template.md` for required sections

### Step 3: Generate Specification

**For Audit Mode:**
- User Scenarios: Infer from `key_functions` names and purposes
- Functional Requirements: Based on `api_contracts`
- Key Entities: Copy from `data_models`
- Success Criteria: Derive from function signatures

**For Manual Mode:**
1. Parse feature description, extract: actors, actions, data, constraints
2. Fill User Scenarios from branch research
3. Generate testable Functional Requirements
4. Define measurable Success Criteria
5. Mark unclear aspects with `[NEEDS CLARIFICATION: question]` (max 3)

### Step 4: Quality Validation

Apply validation from `.specledger/templates/partials/spec-quality-validation.md`:

1. **Content Quality**: No implementation details, focused on user value
2. **Requirement Completeness**: Testable, measurable, bounded scope
3. **Feature Readiness**: Clear acceptance criteria, primary flows covered

**If [NEEDS CLARIFICATION] markers exist (max 3):**
- Present each with suggested options (A, B, C, Custom)
- Wait for user response
- Update spec with answers
- Re-validate

### Step 5: Save and Report

### Step 5: Save and Report

1. Parse user description from Input
   If empty: ERROR "No feature description provided"
2. Extract key concepts from description and existing branch research
   Identify: actors, actions, data, constraints
3. For unclear aspects:
   - Make informed guesses based on context and industry standards
   - Only mark with [NEEDS CLARIFICATION: specific question] if:
     - The choice significantly impacts feature scope or user experience
     - Multiple reasonable interpretations exist with different implications
     - No reasonable default exists
   - **LIMIT: Maximum 3 [NEEDS CLARIFICATION] markers total**
   - Prioritize clarifications by impact: scope > security/privacy > user experience > technical details
4. Fill User Scenarios & Testing section (both included from completed work and missing flows)
   If no clear user flow: ERROR "Cannot determine user scenarios"
5. Generate Functional Requirements
   Each requirement must be testable
   Use reasonable defaults for unspecified details (document assumptions in Assumptions section)
6. Define Success Criteria
   Create measurable, technology-agnostic outcomes
   Include both quantitative metrics (time, performance, volume) and qualitative measures (user satisfaction, task completion)
   Each criterion must be verifiable without implementation details
7. Identify Key Entities (if data involved)
8. Query Beads for related features/tasks
   Include references in Previous work section
9. Check for External References (if feature references external specifications):
   - Does this feature reference external specifications, APIs, or standards for reading/reference?
   - If yes: Note in Dependencies & Assumptions section that external specs should be added
   - Remind user: "If this feature references external specifications for reading/reference, use 'sl deps add' to add them"
   - Example: Reading API contracts from other teams, referencing industry standards, shared design documents
10. Write spec to `SPEC_FILE` path from Step 2
11. Create checklist at `FEATURE_DIR/checklists/requirements.md`
12. Return: SUCCESS (spec ready for planning)

## Guidelines

- Focus on **WHAT** users need, not HOW to implement
- Written for business stakeholders, not developers
- No implementation details (languages, frameworks, APIs)
- Maximum 3 [NEEDS CLARIFICATION] markers

**Reasonable defaults** (don't ask):
- Performance: Standard web/mobile expectations
- Authentication: Session-based or OAuth2
- Integration: RESTful APIs
- Error handling: User-friendly messages

**Success criteria must be:**
- Measurable (time, percentage, count)
- Technology-agnostic
- User-focused outcomes
- Verifiable without implementation details

## Error Handling

- **Empty input**: "No feature description provided"
- **No commits**: "Branch has no commits to analyze"
- **Stale audit**: "Audit cache stale - run /specledger.audit --force"
- **Invalid module-id**: "Module [ID] not found in audit cache"

## Examples

```bash
# From audit data
/specledger.adopt --module-id user-auth --from-audit

# Manual mode with description
/specledger.adopt "Add user authentication with OAuth support"
```
