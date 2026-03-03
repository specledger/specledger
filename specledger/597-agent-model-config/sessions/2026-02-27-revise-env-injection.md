# Session: Revise Command Env Injection Fix

**Date**: 2026-02-27
**Feature**: 597-agent-model-config - Advanced Agent Model Configuration
**Session Type**: Bug Fix + Analysis

## Summary

Fixed a gap where `sl revise` command was not injecting configuration environment variables when launching the agent. Also performed specification analysis to verify feature completeness.

## Problem Identified

The `sl revise` command in `pkg/cli/commands/revise.go` was creating an `AgentLauncher` but not calling `SetEnv()` with the resolved configuration before launching the agent. This meant:

- `sl new` ✓ Injected env vars (via `bootstrap_helpers.go`)
- `sl init` ✓ Injected env vars (via `bootstrap_helpers.go`)
- `sl revise` ✗ Did NOT inject env vars (missing call)

Users running `sl revise` would not have their configured GLM/custom provider settings applied to the agent.

## Solution

Added the missing integration in `pkg/cli/commands/revise.go`:

1. Added import for `github.com/specledger/specledger/pkg/cli/config`
2. Added `al.SetEnv(config.ResolveAgentEnv())` before `al.LaunchWithPrompt()`

### Code Changes

**File**: `pkg/cli/commands/revise.go`

```go
// Before (line 182-188)
al := launcher.NewAgentLauncher(agentOpt, cwd)
if !al.IsAvailable() {
    fmt.Printf("No AI agent found. Install with: %s\n", al.InstallInstructions())
    return writePromptToFile(finalPrompt)
}

fmt.Printf("Launching %s...\n", al.Name)

// After (line 182-191)
al := launcher.NewAgentLauncher(agentOpt, cwd)
if !al.IsAvailable() {
    fmt.Printf("No AI agent found. Install with: %s\n", al.InstallInstructions())
    return writePromptToFile(finalPrompt)
}

// Inject config environment variables (base-url, auth-token, model overrides, etc.)
al.SetEnv(config.ResolveAgentEnv())

fmt.Printf("Launching %s...\n", al.Name)
```

## Verification

### Config Resolution Test

```
$ go run /tmp/testenv.go
ANTHROPIC_DEFAULT_OPUS_MODEL=glm-5
ANTHROPIC_BASE_URL=https://api.z.ai/...
ANTHROPIC_AUTH_TOKEN=fee15305fc4a48718...
ANTHROPIC_DEFAULT_HAIKU_MODEL=glm-4.5-air
ANTHROPIC_DEFAULT_SONNET_MODEL=glm-4.7-flash
```

### Build & Tests

```
$ go build ./...
✓ Go build: Success

$ go test ./pkg/cli/... -v -count=1
✓ Go test: 205 passed in 16 packages

$ make build
go build -o bin/sl cmd/sl/main.go
```

## Specification Analysis

Ran `/specledger.analyze` to verify feature completeness:

### Metrics

| Metric | Value |
|--------|-------|
| Total Functional Requirements | 11 |
| Total User Stories | 4 |
| Total Tasks | 24 |
| Requirement Coverage | 100% |
| Task Completion | 100% |
| Critical Issues | 0 |

### Findings

| ID | Severity | Summary |
|----|----------|---------|
| T1 | LOW | Precedence order doc inconsistency (profile vs global) |
| A1 | LOW | Subjective UX criterion acceptable |
| I1 | LOW | Terminology drift "team-local" vs "local" |

**No blocking issues. Feature is complete.**

## Files Modified

| File | Changes |
|------|---------|
| pkg/cli/commands/revise.go | Added config import, SetEnv call |

## Related Issues

This fix completes the implementation of:
- **FR-007**: System MUST inject configured agent environment variables into the agent subprocess when launching

The `sl revise` command was an uncovered path for this requirement.

## Commands Verified

```bash
# Set GLM config globally
sl config set --global agent.base-url https://api.z.ai/api/anthropic
sl config set --global agent.auth-token fee15305fc4a4871874d70d0184f03ec.yJmq1oYyPI2vjDn1
sl config set --global agent.model.sonnet glm-4.7-flash
sl config set --global agent.model.opus glm-5
sl config set --global agent.model.haiku glm-4.5-air

# Verify config
sl config show

# Now sl revise will use these settings when launching the agent
```

## Conclusion

The gap in `sl revise` env injection has been fixed. All agent launch paths now properly inject configuration environment variables:

| Command | Status |
|---------|--------|
| `sl new` | ✓ Injects config env vars |
| `sl init` | ✓ Injects config env vars |
| `sl revise` | ✓ Injects config env vars (fixed this session) |

Feature 597-agent-model-config is fully complete and ready for production use.
