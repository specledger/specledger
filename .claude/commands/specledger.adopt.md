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

1. Write spec to `SPEC_FILE` path from Step 2
2. Create checklist at `FEATURE_DIR/checklists/requirements.md`
3. Report: branch name, spec path, validation status, next steps

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
