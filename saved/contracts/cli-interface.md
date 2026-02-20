# CLI Interface Contract: sl new (Extended)

**Feature**: 592-init-project-templates
**Command**: `sl new`
**Version**: 2.0.0 (extends v1.0.0 with template/agent selection)

This document defines the command-line interface contract for the extended `sl new` command with template and agent selection.

---

## Command Signature

```bash
sl new [flags]
```

---

## Flags

### Required Flags (Non-Interactive Mode)

| Flag | Type | Description | Example |
|------|------|-------------|---------|
| `--project-name <name>` | string | Project name | `--project-name my-app` |
| `--project-dir <path>` | string | Project directory path | `--project-dir /tmp/my-app` |
| `--short-code <code>` | string | Short code (≤4 chars) | `--short-code ma` |
| `--template <id>` | string | Template ID | `--template full-stack` |
| `--agent <id>` | string | Agent ID | `--agent claude-code` |

**Note**: All required flags MUST be provided when running in non-interactive mode (no TTY).

### Optional Flags

| Flag | Type | Description | Default |
|------|------|-------------|---------|
| `--list-templates` | boolean | List available templates and exit | false |
| `--preview-template <id>` | string | Preview template structure and exit (P2) | N/A |
| `--recommend <description>` | string | Get template recommendations (P3) | N/A |
| `--force` | boolean | Overwrite existing directory | false |
| `--ci` | boolean | CI mode (skip interactive prompts) | false |

### Legacy Flags (Deprecated)

| Flag | Type | Status | Replacement |
|------|------|--------|-------------|
| `--framework <name>` | string | Deprecated | Use `--template` |
| `--playbook <name>` | string | Deprecated | Use `--template` |

---

## Interactive Mode (TTY Detected)

When stdin is a TTY, `sl new` runs in interactive mode with the following prompts:

### Step 1: Project Name
```
┌──────────────────────────────────────────────────┐
│ Project Name                                     │
│ ________________________________________________ │
│                                                  │
│ Enter the name of your project                  │
└──────────────────────────────────────────────────┘
```

**Input**: Text input, validated on Enter
**Validation**: Non-empty, ≤100 characters
**Error**: "Project name cannot be empty"

### Step 2: Project Directory
```
┌──────────────────────────────────────────────────┐
│ Project Directory                                │
│ ________________________________________________ │
│                                                  │
│ Enter the directory path (default: ./<name>)    │
└──────────────────────────────────────────────────┘
```

**Input**: Text input with default value
**Validation**: Valid directory path
**Error**: "Invalid directory path"

### Step 3: Short Code
```
┌──────────────────────────────────────────────────┐
│ Short Code                                       │
│ ________________________________________________ │
│                                                  │
│ Enter a short code (≤4 chars) for issue tracking│
└──────────────────────────────────────────────────┘
```

**Input**: Text input, validated on Enter
**Validation**: Non-empty, ≤4 characters
**Error**: "Short code must be 4 characters or less"

### Step 4: Select Template (NEW)
```
┌──────────────────────────────────────────────────┐
│ Select Project Template                          │
│                                                  │
│  ◉ General Purpose                               │
│    Default template for any project type        │
│    Tech: Go, CLI, Single Binary                 │
│                                                  │
│  ○ Full-Stack Application                        │
│    Go backend with TypeScript/React frontend    │
│    Tech: Go, TypeScript, React, REST API        │
│                                                  │
│  ○ Batch Data Processing                         │
│    Scheduled data pipelines with orchestration  │
│    Tech: Go, Airflow, DAGs, Cron                │
│                                                  │
│  ○ Real-Time Workflow                            │
│    Durable long-running workflow orchestration  │
│    Tech: Go, Temporal, Workflows, Activities    │
│                                                  │
│  ○ ML Image Processing                           │
│    Machine learning pipeline for images         │
│    Tech: Python, TensorFlow, Training, Inference│
│                                                  │
│  ○ Real-Time Data Pipeline                       │
│    Streaming data ingestion and processing      │
│    Tech: Go, Kafka, Streams, Feature Store      │
│                                                  │
│ [↑/↓ to navigate, Enter to select]              │
└──────────────────────────────────────────────────┘
```

**Input**: Arrow keys (↑/↓) + Enter
**Navigation**: Wraps around (top ↑ → bottom, bottom ↓ → top)
**Selection**: Highlighted with `◉` (selected) vs `○` (unselected)

### Step 5: Select Agent (NEW)
```
┌──────────────────────────────────────────────────┐
│ Select Coding Agent                              │
│                                                  │
│  ◉ Claude Code                                   │
│    Anthropic's official CLI with commands/skills│
│                                                  │
│  ○ OpenCode                                      │
│    Open-source AI agent with LLM flexibility    │
│                                                  │
│  ○ None                                          │
│    Agent-agnostic setup (no agent files)        │
│                                                  │
│ [↑/↓ to navigate, Enter to select]              │
└──────────────────────────────────────────────────┘
```

**Input**: Arrow keys (↑/↓) + Enter
**Navigation**: Wraps around
**Selection**: Highlighted with `◉` (selected) vs `○` (unselected)

### Step 6: Confirmation
```
┌──────────────────────────────────────────────────┐
│ Confirm Project Setup                            │
│                                                  │
│  Project Name:    my-app                         │
│  Directory:       /tmp/my-app                    │
│  Short Code:      ma                             │
│  Template:        Full-Stack Application         │
│  Agent:           Claude Code                    │
│                                                  │
│ Press Enter to create project, Ctrl+C to cancel │
└──────────────────────────────────────────────────┘
```

**Input**: Enter (confirm) or Ctrl+C (cancel)

---

## Non-Interactive Mode (No TTY)

When stdin is not a TTY (e.g., in CI/CD), `sl new` requires all flags:

```bash
sl new \
  --project-name my-app \
  --project-dir /tmp/my-app \
  --short-code ma \
  --template full-stack \
  --agent claude-code
```

**Exit Codes**:
- `0`: Success
- `1`: Missing required flags
- `2`: Invalid flag values
- `3`: Directory already exists (use `--force`)
- `4`: Template not found
- `5`: Agent not found
- `6`: Project creation failed

**Error Message Format**:
```
Error: missing required flag: --template

Usage: sl new [flags]

Required flags (non-interactive mode):
  --project-name   Project name
  --project-dir    Project directory
  --short-code     Short code (≤4 chars)
  --template       Template ID (use --list-templates to see options)
  --agent          Agent ID (claude-code, opencode, none)

Run 'sl new --help' for more information.
```

---

## Flag Commands

### List Templates

```bash
sl new --list-templates
```

**Output**:
```
Available Project Templates:

  general-purpose (default)
    Default template for any project type
    Tech: Go, CLI, Single Binary

  full-stack
    Go backend with TypeScript/React frontend
    Tech: Go, TypeScript, React, REST API

  batch-data
    Scheduled data pipelines with workflow orchestration
    Tech: Go, Airflow, DAGs, Cron

  realtime-workflow
    Durable long-running workflow orchestration
    Tech: Go, Temporal, Workflows, Activities

  ml-image
    Machine learning pipeline for image classification
    Tech: Python, TensorFlow, Training, Inference

  realtime-data
    Streaming data ingestion and processing
    Tech: Go, Kafka, Streams, Feature Store

Use: sl new --template <id>
```

**Exit Code**: 0

### Preview Template (P2 - Future)

```bash
sl new --preview-template full-stack
```

**Output**:
```
Template: Full-Stack Application (full-stack)
Description: Go backend with TypeScript/React frontend

Directory Structure:

backend/
├── cmd/
│   └── server/
│       └── main.go          # Backend entry point
├── internal/
│   ├── api/                 # REST API handlers
│   ├── models/              # Data models
│   └── services/            # Business logic
├── go.mod
└── go.sum

frontend/
├── src/
│   ├── components/          # React components
│   ├── pages/               # Page components
│   ├── services/            # API clients
│   └── App.tsx             # Main app component
├── package.json
└── tsconfig.json

tests/
├── backend/                 # Backend tests
└── frontend/                # Frontend tests

README.md                    # Project documentation
specledger.yaml              # Project metadata

Key Files:
- backend/cmd/server/main.go: HTTP server with REST API routes
- frontend/src/App.tsx: Main React application component
- README.md: Getting started guide with build instructions
```

**Exit Code**: 0

**Notes**: No files are created, only preview is displayed.

### Get Recommendations (P3 - Future)

```bash
sl new --recommend "REST API for data ingestion with batch processing"
```

**Output**:
```
Based on your description, we recommend:

1. Batch Data Processing (Match: 85%)
   Reason: Keywords "data ingestion" and "batch processing" match this
   template's focus on scheduled pipelines and data transformation.

2. Real-Time Data Pipeline (Match: 60%)
   Reason: "Data ingestion" suggests streaming, but "batch" indicates
   scheduled processing is primary concern.

3. General Purpose (Match: 40%)
   Reason: Safe default if other templates don't fit your needs.

Use: sl new --template batch-data
```

**Exit Code**: 0

---

## Output

### Success (Interactive Mode)

```
✓ Project created successfully!

  Name:      my-app
  Location:  /tmp/my-app
  Template:  Full-Stack Application
  Agent:     Claude Code

Next steps:
  cd /tmp/my-app
  # Read README.md for project-specific setup instructions
```

**Exit Code**: 0

### Success (Non-Interactive Mode)

```
Project 'my-app' created at /tmp/my-app (template: full-stack, agent: claude-code)
```

**Exit Code**: 0

### Error Examples

**Invalid template ID**:
```
Error: unknown template: invalid-template

Available templates:
  general-purpose, full-stack, batch-data, realtime-workflow, ml-image, realtime-data

Use: sl new --list-templates for details
```

**Exit Code**: 4

**Invalid agent ID**:
```
Error: unknown agent: invalid-agent

Available agents:
  claude-code, opencode, none

Use: sl new --help for details
```

**Exit Code**: 5

**Directory exists**:
```
Error: directory already exists: /tmp/my-app

Use --force to overwrite
```

**Exit Code**: 3

**Non-interactive without flags**:
```
Error: interactive terminal required or provide all flags

Required flags for non-interactive mode:
  --project-name, --project-dir, --short-code, --template, --agent

Run 'sl new --help' for more information.
```

**Exit Code**: 1

---

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `OPENCODE_CONFIG_DIR` | Override OpenCode config location | `~/.config/opencode/` |
| `SL_TEMPLATE_DIR` | Override template source (for testing) | Embedded templates |

---

## File System Side Effects

### General Structure (All Templates)

```
<project-dir>/
├── specledger.yaml          # Project metadata (v1.1.0)
├── AGENTS.md                # Agent context (if agent != none)
├── .<agent-config>/         # Agent config directory (if agent != none)
│   ├── commands/
│   └── skills/
└── <template-specific>/     # Template-specific structure
```

### Metadata File (specledger.yaml)

```yaml
version: 1.1.0
project:
    id: 550e8400-e29b-41d4-a716-446655440000
    name: my-app
    short_code: ma
    template: full-stack
    agent: claude-code
    created: 2026-02-18T17:05:59+07:00
    modified: 2026-02-18T17:05:59+07:00
    version: 0.1.0
playbook:
    name: specledger
    version: 1.0.0
    applied_at: 2026-02-18T17:05:59+07:00
task_tracker:
    choice: beads
    enabled_at: 2026-02-18T17:05:59+07:00
```

**Fields**:
- `project.id`: UUID v4, generated at creation
- `project.template`: Template ID selected by user
- `project.agent`: Agent ID selected by user
- `version`: Metadata schema version (1.1.0)

---

## Backward Compatibility

### Legacy Command (Current Behavior)

```bash
# Old: sl new (prompts for name, dir, short code, playbook)
sl new
```

### New Command (Equivalent Behavior)

```bash
# New: sl new (prompts for name, dir, short code, template, agent)
# Default selections: template=general-purpose, agent=claude-code
sl new
```

**Guarantee**: Accepting all defaults produces identical project structure to old `sl new`.

### Migration Path

**Old projects** (v1.0.0 metadata without `id`, `template`, `agent`):
- `sl` commands auto-generate `project.id` on first load
- `project.template` remains empty (backward compatible)
- `project.agent` inferred from `.claude/` presence or remains empty
- No breaking changes, all old projects continue to work

---

## Contract Tests

### Test Case 1: Interactive Happy Path

```bash
# Given: User runs sl new interactively
$ sl new

# When: User enters inputs and accepts defaults
Project Name: [test-app]
Project Directory: [./test-app]
Short Code: [ta]
Select Template: [General Purpose] (default)
Select Agent: [Claude Code] (default)
Confirm: [Enter]

# Then: Project created with correct structure
$ ls test-app/
specledger.yaml  AGENTS.md  .claude/  cmd/  internal/  pkg/  tests/

# And: Metadata contains UUID, template, agent
$ grep -A5 "^project:" test-app/specledger.yaml
project:
    id: <UUID>
    name: test-app
    short_code: ta
    template: general-purpose
    agent: claude-code
```

### Test Case 2: Non-Interactive with All Flags

```bash
# Given: User runs sl new in CI
$ sl new --project-name ci-app --project-dir /tmp/ci-app \
         --short-code ci --template full-stack --agent opencode

# Then: Project created without prompts
$ echo $?
0

# And: Correct template applied
$ ls /tmp/ci-app/
backend/  frontend/  tests/  specledger.yaml  AGENTS.md  .opencode/

# And: Correct agent config
$ ls /tmp/ci-app/.opencode/
commands/  skills/

# And: No .claude/ directory
$ ls /tmp/ci-app/.claude/
ls: .claude/: No such file or directory
```

### Test Case 3: List Templates

```bash
# Given: User wants to see available templates
$ sl new --list-templates

# Then: All templates listed
# And: Exit code is 0
$ echo $?
0

# And: Output contains template IDs and descriptions
$ sl new --list-templates | grep "full-stack"
  full-stack
```

### Test Case 4: Invalid Template ID

```bash
# Given: User provides invalid template
$ sl new --project-name test --project-dir /tmp/test \
         --short-code t --template invalid --agent claude-code

# Then: Error message displayed
Error: unknown template: invalid

# And: Available templates listed
Available templates:
  general-purpose, full-stack, ...

# And: Exit code is 4
$ echo $?
4
```

### Test Case 5: Non-Interactive Missing Flags

```bash
# Given: User runs sl new in CI without all flags
$ sl new --project-name test

# Then: Error message displayed
Error: missing required flag: --project-dir

# And: Usage help shown
Usage: sl new [flags]

# And: Exit code is 1
$ echo $?
1
```

### Test Case 6: Backward Compatibility

```bash
# Given: User runs new sl new with defaults
$ sl new --project-name compat --project-dir /tmp/compat \
         --short-code co --template general-purpose --agent claude-code

# And: Old sl new would have created reference structure
$ sl-old new --project-name compat-old --project-dir /tmp/compat-old \
             --short-code co --framework none

# Then: New structure matches old structure (except metadata fields)
$ diff -r /tmp/compat /tmp/compat-old --exclude specledger.yaml
<no differences>

# And: Both have Claude Code config
$ ls /tmp/compat/.claude/
commands/  skills/
$ ls /tmp/compat-old/.claude/
commands/  skills/
```

---

## Performance Requirements

| Operation | Target | Measured |
|-----------|--------|----------|
| List templates | <100ms | TBD |
| Preview template | <200ms | TBD (P2) |
| Interactive flow | <60s | TBD |
| Project creation | <5s | TBD |
| UUID generation | <1ms | TBD |

---

## Security Considerations

### Input Validation

- **Project name**: Sanitized to prevent shell injection (no special chars in commands)
- **Directory path**: Validated to prevent path traversal (no `../` sequences allowed outside project)
- **Template ID**: Validated against hardcoded list (no arbitrary template loading)
- **Agent ID**: Validated against hardcoded list (no arbitrary agent loading)

### File Creation

- **Permissions**: All directories created with `0755`, files with `0644`
- **Overwrite protection**: Requires `--force` flag to overwrite existing directory
- **Symlink attacks**: Resolved paths validated before writing

### UUID Security

- **Generation**: Uses `crypto/rand` (cryptographically secure)
- **Collision resistance**: UUID v4 has 122 bits of randomness
- **No PII**: UUID contains no personally identifiable information

---

## Versioning

**CLI Version**: 2.0.0 (semver)
- Major: 2 (breaking change: new required fields in non-interactive mode)
- Minor: 0
- Patch: 0

**Metadata Version**: 1.1.0
- Major: 1 (backward compatible extension)
- Minor: 1 (new fields: `id`, `template`, `agent`)
- Patch: 0

**Breaking Changes**:
- Non-interactive mode now requires `--template` and `--agent` flags
- Metadata schema v1.1.0 adds required `project.id` field

**Deprecations**:
- `--framework` flag (use `--template`)
- `--playbook` flag (use `--template`)

---

## Summary

This CLI contract extends `sl new` with:

1. **New interactive steps**: Template selection (6 options), Agent selection (3 options)
2. **New flags**: `--template`, `--agent`, `--list-templates`
3. **New metadata fields**: `project.id` (UUID), `project.template`, `project.agent`
4. **Backward compatibility**: Defaults preserve old behavior exactly
5. **Non-interactive mode**: Requires all flags, suitable for CI/CD
6. **Error handling**: Clear messages, distinct exit codes, helpful guidance

All contracts are testable with integration tests and maintain backward compatibility with existing projects.
