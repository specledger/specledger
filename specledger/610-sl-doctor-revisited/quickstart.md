# Quickstart: sl doctor revisited

These scenarios map 1:1 to integration/E2E test cases per Constitution Principle VIII.

---

## Scenario 1: Stale file detection (US-1)

```bash
# Setup: create a project with an extra specledger command
mkdir -p /tmp/test-project/.claude/commands
sl init /tmp/test-project
echo "# stale" > /tmp/test-project/.claude/commands/specledger.commit.md

# Run doctor --template
cd /tmp/test-project
sl doctor --template
# Expected: WARNING listing specledger.commit.md as stale
# Expected: suggestion to re-run with --force to delete

# Run with --force to delete stale files
sl doctor --template --force
# Expected: "Deleted stale file: specledger.commit.md"
# Expected: specledger.commit.md is removed from disk
```

## Scenario 2: Custom commands are never touched (US-1 edge case)

```bash
# Setup: place a custom (non-specledger) command
echo "# custom" > /tmp/test-project/.claude/commands/my-deploy.md

sl doctor --template --force
# Expected: my-deploy.md is NOT flagged, NOT deleted
```

## Scenario 3: Doctor from subdirectory (US-3)

```bash
cd /tmp/test-project/pkg/cli/commands
sl doctor --template
# Expected: succeeds, finds project root at /tmp/test-project

cd /tmp/no-project
sl doctor --template
# Expected: error "not in a SpecLedger project (no specledger.yaml found)"
```

## Scenario 4: --check flag dry run (US-3)

```bash
cd /tmp/test-project
sl doctor --check
# Expected: human-readable status showing CLI version + template freshness
# Expected: exit 0 if everything current, exit 1 if updates needed
# Expected: no prompts, no changes made
```

## Scenario 5: specledger.commit removed from binary (US-2)

```bash
sl doctor --template
# Expected: specledger.commit.md is NOT deployed to .claude/commands/
# Expected: if old copy exists, it's reported as stale

# Verify binary doesn't contain the template
sl doctor --json | jq '.template_files'
# Expected: no entry for specledger.commit
```

## Scenario 6: Protected files shown in output (FR-023)

```bash
sl doctor --template
# Expected: output includes "Skipped N protected files: constitution.md, AGENTS.md"
```

## Scenario 7: Scaffold command JSON output (US-4)

```bash
sl spec create --json
# Expected JSON includes:
# {
#   "BRANCH_NAME": "...",
#   "SPEC_FILE": "...",
#   "NEXT_STEPS": ["Read .specledger/templates/spec-template.md before writing the spec"]
# }

sl spec setup-plan --json
# Expected JSON includes:
# {
#   "PLAN_FILE": "...",
#   "NEXT_STEPS": ["Read plan template", "Read checklist template", "Read constitution"]
# }
```

## Scenario 8: Hook opt-out (US-9)

```bash
# Remove hook and persist opt-out
sl auth hook --remove
# Expected: hook removed from ~/.claude/settings.json
# Expected: session_capture: false added to specledger.yaml

# Login should NOT re-install
sl auth login
# Expected: hook NOT re-installed

# Explicit install clears opt-out
sl auth hook --install
# Expected: hook installed
# Expected: session_capture: false removed from specledger.yaml
```

## Scenario 9: Comment UUID prefix matching (US-8)

```bash
# Given a comment with ID fda6ac86-cab4-4850-...
sl comment resolve fda6ac86 --reason "addressed in latest commit"
# Expected: resolves to full UUID, succeeds

sl comment resolve fda
# Expected (if ambiguous): error "ambiguous comment ID, matches: fda6ac86-..., fda7bb92-..."

sl comment resolve nonexistent
# Expected: error "comment not found" with suggestion to run sl comment list
```

## Scenario 10: Onboarding constitution quality (US-5)

```bash
# Run onboarding on a new project
sl init /tmp/new-project
cd /tmp/new-project
# Trigger /specledger.onboard

# Expected: constitution prompt asks for software design principles
# Expected: example categories: testing philosophy, code standards, deployment strategy
# Expected: NOT "Use Go 1.24", "Use Cobra CLI" style tech inventory
```

## Scenario 11: Checkpoint decision log (US-6)

```bash
# During implementation, run checkpoint
# /specledger.checkpoint

# Expected: session log includes ### Decision Log section
# Expected: prompts for divergences from plan/spec
# Expected: structured fields: What, Why, Impact, Artifacts affected
```

## Scenario 12: Implement uses sl issue ready (US-7)

```bash
# During /specledger.implement workflow
sl issue ready
# Expected: returns tasks whose blockers are all satisfied
# Expected: blocked tasks are not offered

# The implement command should call this, not sl issue list --status in_progress
```

## Scenario 13: CI template drift guard (FR-019)

```bash
make build
./bin/sl doctor --template
git diff --exit-code .claude/commands/ .claude/skills/ .specledger/templates/
# Expected: exit 0 if no drift
# Expected: exit 1 if embedded ≠ runtime (CI fails)
```
