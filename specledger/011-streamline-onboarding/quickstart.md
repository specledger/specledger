# Quickstart: 011-streamline-onboarding

**Date**: 2026-02-18
**Branch**: `011-streamline-onboarding`

## Development Setup

```bash
# Ensure Go 1.24+ is installed
go version

# Clone and build
git checkout 011-streamline-onboarding
go build -o sl ./cmd/sl

# Run tests
go test ./pkg/cli/tui/... ./pkg/cli/commands/... ./pkg/cli/launcher/...

# Run integration tests
go test ./tests/integration/ -run TestBootstrap
```

## Key Files to Modify

| File | Change |
| ---- | ------ |
| `pkg/cli/tui/sl_new.go` | Add constitution + agent preference steps |
| `pkg/cli/tui/sl_init.go` (new) | Interactive TUI for `sl init` |
| `pkg/cli/commands/bootstrap.go` | Wire agent launch after setup |
| `pkg/cli/commands/bootstrap_helpers.go` | Add `launchAgent()` helper |
| `pkg/cli/launcher/launcher.go` (new) | Agent availability check + launch |
| `pkg/embedded/templates/specledger/.claude/commands/specledger.onboard.md` (new) | Onboarding workflow command |
| `pkg/embedded/templates/specledger/.specledger/memory/constitution.md` | Add Agent Preferences section |

## Testing Strategy

1. **Unit tests**: TUI step logic, constitution detection, agent availability check
2. **Integration tests**: End-to-end `sl new --ci` and `sl init` with agent launch verification
3. **Manual tests**: Interactive TUI flow, Claude Code launch, onboarding command execution

## Verification Checklist

- [ ] `sl new` presents constitution principles step
- [ ] `sl new` presents agent preference step
- [ ] `sl new` creates populated `.specledger/memory/constitution.md`
- [ ] `sl new` launches Claude Code after setup (when selected)
- [ ] `sl init` presents interactive prompts for missing config
- [ ] `sl init` preserves existing populated constitution
- [ ] `sl init` treats template constitution as "no constitution"
- [ ] `/specledger.onboard` command guides through full workflow
- [ ] Onboarding pauses for task review before implementation
- [ ] `--ci` mode skips agent launch and uses defaults
