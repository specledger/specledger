# Tasks Index: Advanced Agent Model Configuration

Issue Graph Index into the tasks and phases for this feature implementation.
This index does **not contain tasks directly**—those are fully managed through `sl issue` CLI.

## Feature Tracking

* **Epic ID**: `SL-aaf16e`
* **User Stories Source**: `specledger/597-agent-model-config/spec.md`
* **Research Inputs**: `specledger/597-agent-model-config/research.md`
* **Planning Details**: `specledger/597-agent-model-config/plan.md`
* **Data Model**: `specledger/597-agent-model-config/data-model.md`
* **Quickstart Guide**: `specledger/597-agent-model-config/quickstart.md`

## Query Hints

Use the `sl issue` CLI to query and manipulate the issue graph:

```bash
# Find all open tasks for this feature
sl issue list --label spec:597-agent-model-config --status open

# Find ready tasks (unblocked, open)
sl issue ready --label spec:597-agent-model-config

# View issues by component
sl issue list --label "spec:597-agent-model-config,component:config"
sl issue list --label "spec:597-agent-model-config,component:commands"
sl issue list --label "spec:597-agent-model-config,component:launcher"

# Show all phases (feature-type issues)
sl issue list --type feature --label spec:597-agent-model-config

# View task tree for the epic
sl issue show SL-aaf16e
```

## Tasks and Phases Structure

This feature follows a 3-level hierarchy:

* **Epic**: `SL-aaf16e` → Advanced Agent Model Configuration
* **Phases**: Issues of type `feature`, children of the epic
  * Phase 1: Setup (SL-a4323b)
  * Phase 2: Foundational (SL-27dcdd)
  * Phase 3: US1 - Configure Agent Model Overrides via CLI (SL-c85367)
  * Phase 4: US2 - Local vs Global Configuration Hierarchy (SL-209624)
  * Phase 5: US3 - Custom Agent Profiles (SL-b3a289)
  * Phase 6: Polish & Cross-Cutting (SL-01f1c6)
* **Tasks**: Issues of type `task`, children of each feature issue

## Phase Overview

| Phase | Issue ID | Purpose | Tasks |
|-------|----------|---------|-------|
| Setup | SL-a4323b | Config key registry, AgentConfig struct | 3 |
| Foundational | SL-27dcdd | Merge logic, launcher env injection, metadata extension | 3 |
| US1 | SL-c85367 | sl config command, scope flags, masking, launcher integration | 5 |
| US2 | SL-209624 | Personal override file, scope indicators, sensitive warnings | 3 |
| US3 | SL-b3a289 | Profile CRUD, profile commands, agent.env support | 4 |
| Polish | SL-01f1c6 | Tests, validation, documentation | 4 |

## Phase Dependencies

```
Setup (SL-a4323b)
    ↓
Foundational (SL-27dcdd)
    ↓
    ├── US1 (SL-c85367) ──┐
    ├── US2 (SL-209624) ──┼── Polish (SL-01f1c6)
    └── US3 (SL-b3a289) ──┘
```

**US1, US2, US3 can run in parallel** after Foundational completes.

## Task Summary by Phase

### Phase 1: Setup (SL-a4323b)

| ID | Title | Parallel |
|----|-------|----------|
| SL-d420b1 | Define ConfigKeyDef struct and schema registry | ✓ |
| SL-83aa12 | Define AgentConfig struct with all agent fields | ✓ |
| SL-0bb0d3 | Extend Config struct with Agent and Profiles fields | (depends on SL-83aa12) |

### Phase 2: Foundational (SL-27dcdd)

| ID | Title | Parallel |
|----|-------|----------|
| SL-6ba412 | Implement config merge logic | ✓ |
| SL-cc22d9 | Add BuildEnv method to AgentLauncher | ✓ |
| SL-e1df13 | Extend ProjectMetadata with AgentConfig | ✓ |

### Phase 3: US1 - Configure Agent Model Overrides (SL-c85367)

| ID | Title | Parallel |
|----|-------|----------|
| SL-e10d8a | Implement sl config command with set/get/show/unset | — |
| SL-342cd4 | Add --global and --personal scope flags | (depends on SL-e10d8a) |
| SL-24a33c | Mask sensitive values in sl config show | (depends on SL-e10d8a) |
| SL-eb3d1b | Store sensitive values with restricted file permissions | ✓ |
| SL-0a932f | Integrate resolved config with agent launcher | (depends on SL-6ba412, SL-cc22d9) |

### Phase 4: US2 - Local vs Global Hierarchy (SL-209624)

| ID | Title | Parallel |
|----|-------|----------|
| SL-1dd93a | Create specledger.local.yaml personal override file support | ✓ |
| SL-148bce | Display scope indicators in sl config show | ✓ |
| SL-5faa9a | Add warning for sensitive values in git-tracked scope | ✓ |

### Phase 5: US3 - Custom Agent Profiles (SL-b3a289)

| ID | Title | Parallel |
|----|-------|----------|
| SL-3c69b4 | Implement profile CRUD operations | — |
| SL-f0d996 | Implement sl config profile subcommands | (depends on SL-3c69b4) |
| SL-b27eef | Integrate profile values into config merge | (depends on SL-3c69b4) |
| SL-56531a | Implement agent.env arbitrary environment variable support | ✓ |

### Phase 6: Polish (SL-01f1c6)

| ID | Title | Parallel |
|----|-------|----------|
| SL-fe787f | Add integration tests for config CLI commands | ✓ |
| SL-e910b0 | Add unit tests for config merge logic | ✓ |
| SL-0b50c9 | Add config key validation with helpful errors | ✓ |
| SL-880f52 | Update README with config command documentation | ✓ |

## Definition of Done Summary

| Issue ID | DoD Items |
|----------|-----------|
| SL-d420b1 | ConfigKeyDef struct compiles; ConfigKeyType enum has all 5 types; SchemaRegistry.Lookup returns key definition; All agent keys registered |
| SL-83aa12 | AgentConfig struct compiles; YAML serialization produces correct keys; sensitive:"true" tags on AuthToken and APIKey |
| SL-0bb0d3 | Config struct extended; Load() handles missing Agent field; Save() writes new fields |
| SL-6ba412 | MergeConfigs function implemented; Precedence order verified; ResolvedConfig/ResolvedValue structs defined |
| SL-cc22d9 | SetEnv method added; BuildEnv returns correct env slice; Launch() uses cmd.Env |
| SL-e1df13 | ProjectMetadata struct extended; Load() handles missing Agent; Save() writes new fields |
| SL-e10d8a | set/get/show/unset commands work; Values persist; Invalid keys rejected |
| SL-342cd4 | --global flag works; --personal flag works; Default targets team-local |
| SL-24a33c | auth-token displays masked; api-key displays masked; Non-sensitive values display in full |
| SL-eb3d1b | Global config file has 0600 permissions; Personal-local config has 0600 |
| SL-0a932f | sl new uses resolved config; sl init uses resolved config; Env vars injected |
| SL-1dd93a | LoadPersonal() reads specledger.local.yaml; SavePersonal() writes; File added to .gitignore |
| SL-148bce | Scope indicators display correctly; All 5 scope types supported |
| SL-5faa9a | Warning displays for auth-token/api-key in team-local; Warning recommends --personal |
| SL-3c69b4 | CreateProfile stores profile; DeleteProfile removes; ListProfiles returns names |
| SL-f0d996 | profile create/use/list/delete work; profile use --none deactivates |
| SL-b27eef | Profile values merged; Precedence correct; Scope indicator shows profile name |
| SL-56531a | agent.env.KEY set works; unset works; Env vars shown in show; Injected at launch |

## MVP Scope

**Recommended MVP**: Complete Setup → Foundational → US1

This delivers the core value proposition: persistent agent configuration via CLI commands that replaces brittle shell aliases.

**MVP Commands Available**:
- `sl config set agent.base-url https://api.example.com`
- `sl config set agent.model.sonnet gpt-4-turbo`
- `sl config set --personal agent.auth-token sk-xxx`
- `sl config show`
- `sl config unset agent.base-url`

**Post-MVP**: US2 (hierarchy), US3 (profiles), Polish can follow incrementally.

## Implementation Strategy

1. **Start with Setup** (SL-a4323b): Complete all 3 tasks. T001 and T002 are parallel.
2. **Then Foundational** (SL-27dcdd): All 3 tasks are parallel.
3. **US1, US2, US3 in parallel**: After Foundational, different agents can work on different user stories simultaneously.
4. **Polish last**: After all stories complete, run tests and documentation in parallel.

## Parallel Execution Opportunities

**Maximum parallelization** (with 3+ agents):
```
Agent A: Setup → Foundational → US1 → Polish tests
Agent B: (wait for Foundational) → US2 → Polish docs
Agent C: (wait for Foundational) → US3
```

**Minimum sequential** (single agent):
```
Setup → Foundational → US1 → US2 → US3 → Polish
```
