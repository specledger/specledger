# Quickstart: Skills Registry Integration

**Branch**: `610-skills-registry` | **Date**: 2026-04-05

These scenarios map 1:1 to E2E test cases per Constitution Principle VIII.

## Prerequisites

- `sl` binary built and installed (`make build && make install`)
- Project initialized with `sl init` (specledger.yaml exists with agent selection)
- Internet connectivity (skills.sh API access)

## Scenario 1: Search for Skills

```bash
# Search by keyword
$ sl skill search "web design"
web-design-guidelines   vercel-labs/agent-skills   210.6K installs
web-design-reviewer     github/awesome-copilot       8.7K installs
web-component-design    wshobson/agents              4.7K installs
→ sl skill add <owner/repo@skill> to install

# Search with limit
$ sl skill search "testing" --limit 3

# Search with JSON output
$ sl skill search "deploy" --json
[{"id":"owner/repo/skill","name":"skill","source":"owner/repo","installs":1234}]

# Search with no results
$ sl skill search "xyznonexistent"
No skills found for "xyznonexistent"
```

**E2E test**: `TestSkillsSearch`, `TestSkillsSearchJSON`, `TestSkillsSearchNoResults`, `TestSkillsSearchLimit`

## Scenario 2: Install a Skill

```bash
# Install a specific skill from a repo
$ sl skill add vercel-labs/agent-skills@creating-pr

Security Risk Assessments
                  Gen          Socket       Snyk
creating-pr       Safe         0 alerts     Safe

Install creating-pr from vercel-labs/agent-skills? [Y/n] y
✓ Installed creating-pr to .claude/skills/creating-pr/SKILL.md
✓ Updated skills-lock.json

# Install with JSON output (non-interactive)
$ sl skill add vercel-labs/agent-skills@creating-pr --json -y

# Install all skills from a repo
$ sl skill add vercel-labs/agent-skills -y

# Overwrite existing skill
$ sl skill add vercel-labs/agent-skills@creating-pr
creating-pr is already installed. Overwrite? [y/N]
```

**E2E test**: `TestSkillsAdd`, `TestSkillsAddJSON`, `TestSkillsAddOverwrite`, `TestSkillsAddInvalidSource`

## Scenario 3: View Skill Info

```bash
# View skill details with audit
$ sl skill info vercel-labs/agent-skills@creating-pr
creating-pr (vercel-labs/agent-skills)
Description: Orchestrate the full PR creation workflow

Security Risk Assessments
  Gen:    Safe (score: 100, analyzed: 2026-03-01)
  Socket: 0 alerts (score: 95, analyzed: 2026-03-01)
  Snyk:   Safe (score: 100, analyzed: 2026-03-01)

# JSON output
$ sl skill info vercel-labs/agent-skills@creating-pr --json
```

**E2E test**: `TestSkillsInfo`, `TestSkillsInfoJSON`, `TestSkillsInfoNoAudit`

## Scenario 4: List Installed Skills

```bash
# List installed skills
$ sl skill list
creating-pr   vercel-labs/agent-skills
commit        vercel-labs/agent-skills
→ 2 skill(s) installed. Use 'sl skill audit' to check security.

# JSON output
$ sl skill list --json
[{"name":"creating-pr","source":"vercel-labs/agent-skills","sourceType":"github","computedHash":"abc123"}]

# No skills installed
$ sl skill list
No skills installed.
→ Use 'sl skill search' to discover skills.
```

**E2E test**: `TestSkillsList`, `TestSkillsListJSON`, `TestSkillsListEmpty`

## Scenario 5: Remove a Skill

```bash
# Remove by name
$ sl skill remove creating-pr
✓ Removed creating-pr from .claude/skills/
✓ Updated skills-lock.json

# Remove non-existent skill
$ sl skill remove nonexistent
Error: skill "nonexistent" is not installed.
→ Use 'sl skill list' to see installed skills.

# JSON output
$ sl skill remove creating-pr --json
```

**E2E test**: `TestSkillsRemove`, `TestSkillsRemoveJSON`, `TestSkillsRemoveNotInstalled`

## Scenario 6: Audit Installed Skills

```bash
# Audit all installed skills
$ sl skill audit
Security Risk Assessments for 2 installed skill(s)

                  Gen          Socket       Snyk
creating-pr       Safe         0 alerts     Safe
commit            Safe         0 alerts     Low Risk

✓ No high or critical risks detected.

# Audit specific skill
$ sl skill audit creating-pr

# Audit with high-risk warning
$ sl skill audit
⚠ 1 skill has HIGH or CRITICAL risk. Review before using.

# JSON output
$ sl skill audit --json
```

**E2E test**: `TestSkillsAudit`, `TestSkillsAuditJSON`, `TestSkillsAuditSingle`, `TestSkillsAuditWarning`

## Scenario 7: Error Handling

```bash
# Network error (3-part: what failed, raw error, suggested fix)
$ sl skill search "test"  # (with no connectivity)
Error: sl skill search failed: skills.sh API unreachable
→ Check your internet connection and try again.
→ skills.sh status: https://skills.sh

# Invalid source format
$ sl skill add invalid-source
Error: sl skill add failed: invalid source "invalid-source"
→ Use format: owner/repo or owner/repo@skill-name
→ Example: sl skill add vercel-labs/agent-skills@creating-pr

# Repository not found
$ sl skill add nonexistent/repo@skill
Error: sl skill add failed (404): repository "nonexistent/repo" not found
→ Verify the repository exists and is public.
→ For non-GitHub repos, use the full git URL.

# Skill not found in repo
$ sl skill add vercel-labs/agent-skills@nonexistent
Error: sl skill add failed: skill "nonexistent" not found in vercel-labs/agent-skills
→ Use 'sl skill add vercel-labs/agent-skills' to see available skills.

# Corrupted lock file
$ sl skill list  # (with malformed skills-lock.json)
Error: skills-lock.json is invalid.
→ Fix the JSON syntax or delete skills-lock.json to start fresh.
```

**E2E test**: `TestSkillsErrorNetwork`, `TestSkillsErrorInvalidSource`, `TestSkillsErrorCorruptLock`

## Scenario 8: Telemetry

```bash
# Normal install (telemetry sent)
$ sl skill add vercel-labs/agent-skills@creating-pr -y
# → GET https://add-skill.vercel.sh/t?v=specledger-1.2.0&event=install&source=vercel-labs/agent-skills&skills=creating-pr&agents=claude-code

# Telemetry disabled
$ DISABLE_TELEMETRY=1 sl skill add vercel-labs/agent-skills@creating-pr -y
# → No telemetry sent
```

**E2E test**: `TestSkillsTelemetrySent`, `TestSkillsTelemetryDisabled`
