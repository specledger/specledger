# Feature Specification: Fix Embedded Skill Templates

**Feature Branch**: `607-fix-skill-templates`
**Created**: 2026-03-12
**Status**: Draft
**Input**: User description: "fix https://github.com/specledger/specledger/issues/82"
**GitHub Issue**: [#82](https://github.com/specledger/specledger/issues/82)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Claude Code Triggers Correct Skill (Priority: P1)

As a developer using SpecLedger, when I mention "dependencies" or "sl deps" in a conversation, Claude Code should load the `sl-deps` skill with correct content about managing specification dependencies (not issue tracking content).

**Why this priority**: Without correct content, users receive wrong guidance and wasted tokens. This is a blocking issue for anyone using `sl deps` functionality.

**Independent Test**: Can be tested by examining the `sl-deps/skill.md` file content and verifying it contains `sl deps` commands (add/remove/list/resolve/link/unlink) rather than `sl issue` commands.

**Acceptance Scenarios**:

1. **Given** the `sl-deps/skill.md` file, **When** examining its content, **Then** it describes `sl deps add`, `sl deps remove`, `sl deps list`, `sl deps resolve`, `sl deps link`, and `sl deps unlink` commands
2. **Given** the `sl-deps/skill.md` file, **When** examining its content, **Then** it does NOT contain `sl issue` commands or issue tracking guidance
3. **Given** the `sl-deps/skill.md` file, **When** examining its content, **Then** it clearly distinguishes `sl deps` (cross-repo spec dependencies) from `sl issue link` (inter-issue dependencies)

---

### User Story 2 - Token-Efficient Skill Content (Priority: P1)

As a developer using SpecLedger, when Claude Code loads skill files, they should not contain duplicate sections that waste tokens.

**Why this priority**: Every duplicate section wastes ~700 tokens per skill load, directly impacting context window efficiency.

**Independent Test**: Can be tested by scanning `sl-audit/skill.md` for duplicate content sections.

**Acceptance Scenarios**:

1. **Given** the `sl-audit/skill.md` file, **When** scanning for duplicate content, **Then** no section appears more than once
2. **Given** the `sl-audit/skill.md` file, **When** counting lines, **Then** the file does not exceed 240 lines (after removing 65 lines of duplicate content)

---

### User Story 3 - Rich Skill Descriptions for Reliable Triggering (Priority: P2)

As a developer using SpecLedger, when I look at available skills in Claude Code's system prompt, I should see descriptive skill names with trigger keywords to help Claude Code reliably select the correct skill.

**Why this priority**: Generic descriptions like "sl-comment Skill" don't help Claude Code understand when to trigger. Better descriptions improve skill activation accuracy.

**Independent Test**: Can be tested by examining the manifest.yaml skill descriptions and verifying they contain trigger keywords.

**Acceptance Scenarios**:

1. **Given** the `manifest.yaml` file, **When** examining skill descriptions, **Then** each description includes relevant trigger keywords (command names, use cases)
2. **Given** the `manifest.yaml` file, **When** examining the `sl-deps` description, **Then** it mentions "cross-repo", "multi-repo", or "spec imports" to distinguish from `sl issue link`
3. **Given** the `manifest.yaml` file, **When** examining the `sl-issue-tracking` description, **Then** it mentions "multi-session", "task tracking", or "inter-issue dependencies"

---

### User Story 4 - Remove Aspirational Content (Priority: P2)

As a developer using SpecLedger, when I load a skill, it should only contain guidance for implemented features, not references to non-existent commands.

**Why this priority**: Aspirational content confuses the agent about what's actually available, leading to failed command attempts.

**Independent Test**: Can be tested by verifying `sl-audit/skill.md` does not reference `--force` flag or `scripts/audit-cache.json` for a non-existent `sl audit` command.

**Acceptance Scenarios**:

1. **Given** the `sl-audit/skill.md` file, **When** examining content, **Then** it does not reference `--force` flag for an `sl audit` command
2. **Given** the `sl-audit/skill.md` file, **When** examining content, **Then** cache strategy sections reference manual file paths only, not automated cache management

---

### Edge Cases

- What happens when deployed skills have uppercase filenames (SKILL.md) while embedded templates use lowercase (skill.md)? → `sl init` and `sl doctor --template` should handle filename case consistently
- What happens if users have customized their deployed skill files? → Template sync should preserve user modifications or warn before overwriting

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The `sl-deps/skill.md` file MUST contain correct content about `sl deps` commands (add, remove, list, resolve, link, unlink)
- **FR-002**: The `sl-deps/skill.md` file MUST NOT contain content about `sl issue` commands
- **FR-003**: The `sl-deps/skill.md` file MUST distinguish between `sl deps` (cross-repo spec dependencies) and `sl issue link` (inter-issue dependencies)
- **FR-004**: The `sl-audit/skill.md` file MUST NOT contain duplicate sections
- **FR-005**: All skill files MUST use lowercase `skill.md` filename convention
- **FR-006**: Skill descriptions in `manifest.yaml` MUST include trigger keywords for reliable Claude Code activation
- **FR-007**: Skill content MUST NOT reference non-existent CLI commands or flags
- **FR-008**: Skills MUST focus on decision patterns and workflow orchestration, deferring command syntax to `--help` output

### Key Entities

- **Skill File**: Markdown file providing guidance for Claude Code on when and how to use CLI commands
- **Manifest Entry**: YAML configuration defining skill name, path, and description for Claude Code system prompt
- **Token Efficiency**: Principle that skill content should minimize context window usage while maximizing guidance value

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `sl-deps/skill.md` content matches the purpose described in manifest.yaml ("Manage specification dependencies with sl deps commands")
- **SC-002**: `sl-audit/skill.md` contains no duplicate sections (verified by line comparison)
- **SC-003**: Total token count of all skill files reduced by at least 700 tokens (from removing duplicates in sl-audit)
- **SC-004**: All four skill descriptions in manifest.yaml contain at least 3 trigger keywords each
- **SC-005**: No skill file references CLI commands or flags that don't exist in the current codebase

### Previous work

- **008-fix-sl-deps**: Previous attempt to fix sl-deps skill (may have been incomplete or overwritten)
- **598-sdd-workflow-streamline**: Established D21 token efficiency principle that skills should focus on decision patterns
- **601-cli-skills**: CLI skills architecture and progressive disclosure design patterns

### Epic: 607 - Fix Embedded Skill Templates

- **Fix sl-deps content**: Replace duplicate issue tracking content with correct deps guidance
- **Remove sl-audit duplicates**: Delete lines 239-271 containing duplicated CLI Reference and Troubleshooting
- **Enhance skill descriptions**: Add trigger keywords to manifest.yaml skill entries
- **Remove aspirational content**: Clean up references to non-existent commands

## Dependencies & Assumptions

### Dependencies

- Existing `sl deps` CLI commands must be functional (add, remove, list, resolve, link, unlink)
- The manifest.yaml parsing logic must support updated descriptions

### Assumptions

- Users have not heavily customized their deployed skill files (or are willing to reconcile during `sl doctor --template`)
- The `sl deps` commands are sufficiently documented via `--help` to supplement skill guidance
- Progressive disclosure structure (decision patterns first, command reference second) is the desired format

## Out of Scope

- Adding new `sl audit` CLI command (skill provides patterns for manual reconnaissance)
- Changing the skill loading mechanism in Claude Code
- Creating `sl playbook skills` CLI subcommand for progressive skill discovery (P1 recommendation from issue)
- Adding `sl doctor --template` skill sync functionality (P2 recommendation from issue)
