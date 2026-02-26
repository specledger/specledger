# Quickstart: Issue Create Fields Enhancement

**Feature**: 597-issue-create-fields
**Date**: 2026-02-24

## Prerequisites

- Built `sl` binary: `make build`
- In a SpecLedger project directory

## Test Scenarios

### Scenario 1: Create Issue with All New Fields

```bash
# Create an issue with all 5 new fields
./bin/sl issue create \
  --title "Implement user authentication" \
  --type task \
  --priority 1 \
  --acceptance-criteria "User can log in with email/password. Invalid credentials show error." \
  --dod "Write unit tests" \
  --dod "Add integration test" \
  --dod "Update documentation" \
  --design "Use JWT tokens with 24h expiry. Store hashed passwords using bcrypt." \
  --notes "Consider adding OAuth support in future"

# Verify the issue was created
./bin/sl issue show <issue-id>
```

**Expected**: Issue displays with dedicated sections for Acceptance Criteria, Design, Definition of Done, and Notes.

### Scenario 2: Create Issue with Repeated --dod Flags

```bash
# Create issue with multiple DoD items
./bin/sl issue create \
  --title "Add API endpoint" \
  --type task \
  --dod "Create route handler" \
  --dod "Add request validation" \
  --dod "Write tests" \
  --dod "Update API docs"

# Show the issue to verify DoD items
./bin/sl issue show <issue-id>
```

**Expected**: Definition of Done shows 4 unchecked items.

### Scenario 3: Update DoD on Existing Issue

```bash
# Replace entire DoD
./bin/sl issue update <issue-id> \
  --dod "New requirement 1" \
  --dod "New requirement 2"

# Verify replacement
./bin/sl issue show <issue-id>
```

**Expected**: Previous DoD items replaced with 2 new unchecked items.

### Scenario 4: Check DoD Item

```bash
# Check a specific DoD item
./bin/sl issue update <issue-id> --check-dod "New requirement 1"

# Verify item is checked
./bin/sl issue show <issue-id>
```

**Expected**: "New requirement 1" shows `[x]` with a verified_at timestamp.

### Scenario 5: Check Non-existent DoD Item (Error Case)

```bash
# Try to check an item that doesn't exist
./bin/sl issue update <issue-id> --check-dod "Nonexistent item"
```

**Expected**: Error message: `DoD item not found: 'Nonexistent item'`

### Scenario 6: Uncheck DoD Item

```bash
# Uncheck a previously checked item
./bin/sl issue update <issue-id> --uncheck-dod "New requirement 1"

# Verify item is unchecked
./bin/sl issue show <issue-id>
```

**Expected**: "New requirement 1" shows `[ ]` and verified_at is cleared.

### Scenario 7: Exact Text Matching

```bash
# Create issue with DoD item
./bin/sl issue create --title "Test" --type task --dod "Write Tests"

# These should all FAIL (exact match required):
./bin/sl issue update <issue-id> --check-dod "write tests"    # Wrong case
./bin/sl issue update <issue-id> --check-dod "Write Tests "   # Trailing space
./bin/sl issue update <issue-id> --check-dod "Write  Tests"   # Double space

# This should SUCCEED:
./bin/sl issue update <issue-id> --check-dod "Write Tests"    # Exact match
```

**Expected**: Only exact match succeeds.

### Scenario 8: Issue Show Display

```bash
# Create a fully-specified issue
./bin/sl issue create \
  --title "Complex task" \
  --description "This is the description" \
  --type task \
  --acceptance-criteria "AC1: Success case. AC2: Error case." \
  --design "Use factory pattern with dependency injection" \
  --dod "Implement core logic" \
  --dod "Add error handling" \
  --notes "Reference: https://example.com/docs"

# Display the issue
./bin/sl issue show <issue-id>
```

**Expected Output Format**:
```
Issue: SL-xxxxxx
  Title: Complex task
  Type: task
  Status: open
  Priority: 2
  Spec: <spec-context>

Description:
  This is the description

Acceptance Criteria:
  AC1: Success case. AC2: Error case.

Design:
  Use factory pattern with dependency injection

Definition of Done:
  [ ] Implement core logic
  [ ] Add error handling

Notes:
  Reference: https://example.com/docs

Created: 2026-02-24 10:30:00
Updated: 2026-02-24 10:30:00
```

---

## Parent-Child Relationship Scenarios

### Scenario 9: Create Issue with Parent

```bash
# Create a parent epic
EPIC_ID=$(./bin/sl issue create --title "Epic: User Auth" --type epic --format json | jq -r '.id')

# Create a child feature
./bin/sl issue create \
  --title "Feature: Login" \
  --type feature \
  --parent $EPIC_ID

# Verify parent relationship
./bin/sl issue show <child-id>
```

**Expected**: Issue displays with `Parent: SL-xxxxx` field.

### Scenario 10: Set Parent on Existing Issue

```bash
# Create two issues
PARENT_ID=$(./bin/sl issue create --title "Parent Task" --type task --format json | jq -r '.id')
CHILD_ID=$(./bin/sl issue create --title "Child Task" --type task --format json | jq -r '.id')

# Set parent
./bin/sl issue update $CHILD_ID --parent $PARENT_ID

# Verify
./bin/sl issue show $CHILD_ID
```

**Expected**: Issue displays with `Parent: SL-xxxxx` field.

### Scenario 11: Single Parent Constraint (Error Case)

```bash
# Create three issues
PARENT1_ID=$(./bin/sl issue create --title "Parent 1" --type task --format json | jq -r '.id')
PARENT2_ID=$(./bin/sl issue create --title "Parent 2" --type task --format json | jq -r '.id')
CHILD_ID=$(./bin/sl issue create --title "Child" --type task --parent $PARENT1_ID --format json | jq -r '.id')

# Try to set second parent (should fail)
./bin/sl issue update $CHILD_ID --parent $PARENT2_ID
```

**Expected**: Error message: `issue already has a parent, remove existing parent first`

### Scenario 12: Remove Parent

```bash
# Create issue with parent
PARENT_ID=$(./bin/sl issue create --title "Parent" --type task --format json | jq -r '.id')
CHILD_ID=$(./bin/sl issue create --title "Child" --type task --parent $PARENT_ID --format json | jq -r '.id')

# Remove parent
./bin/sl issue update $CHILD_ID --parent ""

# Verify parent is cleared
./bin/sl issue show $CHILD_ID
```

**Expected**: Issue no longer shows Parent field.

### Scenario 13: Self as Parent (Error Case)

```bash
# Create an issue
ISSUE_ID=$(./bin/sl issue create --title "Test" --type task --format json | jq -r '.id')

# Try to set self as parent
./bin/sl issue update $ISSUE_ID --parent $ISSUE_ID
```

**Expected**: Error message: `cannot set self as parent`

### Scenario 14: Circular Parent (Error Case)

```bash
# Create three issues in a chain
A_ID=$(./bin/sl issue create --title "A" --type task --format json | jq -r '.id')
B_ID=$(./bin/sl issue create --title "B" --type task --parent $A_ID --format json | jq -r '.id')
C_ID=$(./bin/sl issue create --title "C" --type task --parent $B_ID --format json | jq -r '.id')

# Try to create cycle: A -> B -> C -> A
./bin/sl issue update $A_ID --parent $C_ID
```

**Expected**: Error message: `circular parent-child relationship detected`

### Scenario 15: Non-existent Parent (Error Case)

```bash
# Create an issue
ISSUE_ID=$(./bin/sl issue create --title "Test" --type task --format json | jq -r '.id')

# Try to set non-existent parent
./bin/sl issue update $ISSUE_ID --parent SL-nonexistent
```

**Expected**: Error message: `parent issue not found: SL-nonexistent`

### Scenario 16: Tree View with Children

```bash
# Create hierarchy
EPIC_ID=$(./bin/sl issue create --title "Epic" --type epic --format json | jq -r '.id')
FEATURE_ID=$(./bin/sl issue create --title "Feature" --type feature --parent $EPIC_ID --format json | jq -r '.id')
TASK1_ID=$(./bin/sl issue create --title "Task 1" --type task --parent $FEATURE_ID --priority 1 --format json | jq -r '.id')
TASK2_ID=$(./bin/sl issue create --title "Task 2" --type task --parent $FEATURE_ID --priority 2 --format json | jq -r '.id')

# View tree
./bin/sl issue show $EPIC_ID --tree
```

**Expected Output**:
```
Issue: SL-xxxxx (Epic)
└── Issue: SL-yyyyy (Feature)
    ├── Issue: SL-zzzz1 (Task 1) [P1]
    └── Issue: SL-zzzz2 (Task 2) [P2]
```

Note: Children are ordered by priority (P1 before P2).

---

## Verification Checklist

### Create Flags
- [ ] `--acceptance-criteria` flag works on create
- [ ] `--dod` flag works with repeated values on create
- [ ] `--design` flag works on create
- [ ] `--notes` flag works on create
- [ ] `--parent` flag works on create

### Update Flags
- [ ] `--dod` flag replaces entire DoD on update
- [ ] `--check-dod` marks item as checked with timestamp
- [ ] `--uncheck-dod` marks item as unchecked
- [ ] `--parent` flag sets parent on update
- [ ] `--parent ""` clears parent

### Error Handling
- [ ] Error returned when checking non-existent DoD item
- [ ] Exact text matching (case-sensitive, no normalization)
- [ ] Error returned when setting second parent
- [ ] Error returned when setting self as parent
- [ ] Error returned when creating circular parent chain
- [ ] Error returned when setting non-existent parent

### Display
- [ ] `sl issue show` displays all 5 fields in dedicated sections
- [ ] `sl issue show` displays parent relationship
- [ ] `sl issue show --tree` displays children in tree format
- [ ] Children are ordered by priority then ID
