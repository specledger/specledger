# Data Model: 011-streamline-onboarding

**Date**: 2026-02-18
**Branch**: `011-streamline-onboarding`

## Entities

### 1. ProjectMetadata (existing — extended)

**File**: `specledger/specledger.yaml`
**Package**: `pkg/cli/metadata/`

No changes to existing schema. Agent preference is stored in the constitution, not here. `specledger.yaml` continues to handle: project name, short code, version, playbook, artifact path, dependencies, task tracker choice.

### 2. Constitution (existing template — extended)

**File**: `.specledger/memory/constitution.md`
**Format**: Markdown with structured sections

**New section added**:

```markdown
## Agent Preferences

- **Preferred Agent**: [agent_name]
```

**Valid values for `Preferred Agent`**:
- `Claude Code` — launches `claude` CLI
- `None` — no agent launched; manual workflow

**Detection rules**:
- **Populated**: File exists AND does not contain `[ALL_CAPS_IDENTIFIER]` placeholder patterns (regex: `\[[A-Z_]{3,}\]`)
- **Unfilled/Template**: File exists AND contains placeholder patterns → treat as "no constitution"
- **Missing**: File does not exist → treat as "no constitution"

**State transitions**:

```
Missing/Template → sl new TUI (default principles) → Populated
Missing/Template → sl init + AI agent (audit → propose) → Populated
Populated → sl init --force → Updated (re-run with existing as defaults)
Populated → /specledger.constitution → Updated (manual revision)
```

### 3. TUI Model (existing — extended)

**File**: `pkg/cli/tui/sl_new.go`
**Current steps**: 5 (projectName → directory → shortCode → playbook → confirm)

**New steps**:

| Step | Name | Type | Applies to |
| ---- | ---- | ---- | ---------- |
| 5 | stepConstitution | Multi-select list | `sl new` only |
| 6 | stepAgentPreference | Single-select list | `sl new` and `sl init` |
| 7 | stepConfirm (moved) | Confirmation | Both |

**stepConstitution data**:

```go
type ConstitutionPrinciple struct {
    Name        string // e.g., "Specification-First"
    Description string // e.g., "Every feature starts with a spec before code"
    Selected    bool   // Default: true (all suggested principles pre-selected)
}
```

**Default principles** (presented during `sl new`):
1. Specification-First — Every feature starts with a spec before code
2. Test-First — Tests written before implementation; TDD enforced
3. Code Quality — Consistent formatting, linting, and review standards
4. Simplicity — Start simple; avoid premature abstraction (YAGNI)
5. Observability — Structured logging and metrics for debuggability

**stepAgentPreference data**:

```go
type AgentOption struct {
    Name        string // e.g., "Claude Code"
    Command     string // e.g., "claude"
    Description string // e.g., "AI coding assistant with SpecLedger integration"
}
```

**Available agents**:
1. Claude Code — `claude` — AI coding assistant with deep SpecLedger integration
2. None — (no command) — Skip agent launch; manual workflow

### 4. Init TUI Model (new)

**File**: `pkg/cli/tui/sl_init.go` (new file)
**Purpose**: Interactive prompts for `sl init`, presenting only missing configuration

**Dynamic step list** — determined at model creation:

```go
type InitModel struct {
    steps        []InitStep    // Only steps for missing config
    currentStep  int
    answers      map[string]string
    // ... standard Bubble Tea fields
}

type InitStep struct {
    Key       string // "short_code", "playbook", "agent_preference"
    Component interface{} // textinput.Model or list selection
}
```

**Step inclusion logic**:
- `short_code`: Include if `--short-code` flag not provided
- `playbook`: Include if `--playbook` flag not provided
- `agent_preference`: Always include (unless constitution has it already)
- `constitution`: Never included (delegated to AI agent for `sl init`)

### 5. Onboarding Command (new embedded template)

**File**: `pkg/embedded/templates/specledger/.claude/commands/specledger.onboard.md`
**Format**: Claude Code command markdown

**Key data passed to onboarding**:
- `has_constitution`: boolean — whether populated constitution exists
- `source_command`: string — "sl new" or "sl init" (determines workflow path)

**Workflow state machine**:

```
START
  ├─ has_constitution=false → AUDIT → CONSTITUTION → WELCOME
  └─ has_constitution=true  → WELCOME
WELCOME → ASK_FEATURE_DESCRIPTION
ASK_FEATURE_DESCRIPTION → /specledger.specify
→ /specledger.clarify
→ /specledger.plan
→ /specledger.tasks
→ REVIEW_PAUSE (wait for explicit user approval)
→ /specledger.implement (only after approval)
```

### 6. Agent Launcher (new)

**Package**: `pkg/cli/launcher/` (new package)

```go
type AgentLauncher struct {
    Name    string // "claude"
    Command string // "claude"
    Dir     string // Project directory
}

// Methods:
// IsAvailable() bool — exec.LookPath check
// Launch() error — exec.Command with stdio passthrough
// InstallInstructions() string — help text for installation
```

## Relationships

```
specledger.yaml (metadata)
    ├── project.short_code ← TUI step 3
    ├── playbook.name ← TUI step 4
    └── (unchanged)

constitution.md (governance)
    ├── Core Principles ← TUI step 5 (sl new) or AI agent (sl init)
    └── Agent Preferences.preferred_agent ← TUI step 6

onboard command
    ├── reads → constitution.md (check populated)
    ├── triggers → /specledger.audit (if no constitution)
    ├── triggers → /specledger.constitution (if no constitution)
    └── orchestrates → specify → clarify → plan → tasks → implement

launcher
    ├── reads → constitution.md (agent preference)
    └── launches → claude (or other agent)
```

## Validation Rules

- **Short code**: 1-4 alphanumeric characters, lowercase (existing)
- **Agent preference**: Must be one of the known agent options
- **Constitution populated check**: No `\[[A-Z_]{3,}\]` patterns remaining in file
- **Agent availability**: `exec.LookPath(command)` must succeed before launch attempt
