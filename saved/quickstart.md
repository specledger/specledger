# Quick Start: Project Templates & Agent Selection

**Feature**: 592-init-project-templates
**Target**: Developers implementing template/agent selection for `sl new`

This guide provides a fast path to understanding and implementing the template/agent selection feature.

---

## 5-Minute Overview

### What's Changing?

**Before** (Current `sl new`):
```
sl new → Prompts → Creates general-purpose project with Claude Code
```

**After** (New `sl new`):
```
sl new → Prompts → Select Template (6 options) → Select Agent (3 options) → Creates project
```

### Key Additions

1. **6 Project Templates**: General Purpose, Full-Stack, Batch Data, Real-Time Workflow, ML Image, Real-Time Data Pipeline
2. **3 Coding Agents**: Claude Code (default), OpenCode, None
3. **Project ID**: UUID generated for each project (Supabase session storage)
4. **2 New TUI Steps**: Template selection, Agent selection
5. **2 New CLI Flags**: `--template`, `--agent`

---

## Core Files to Modify

| File | Purpose | Changes |
|------|---------|---------|
| `pkg/models/template.go` | Template definitions | **NEW** - TemplateDefinition struct |
| `pkg/models/agent.go` | Agent configurations | **NEW** - AgentConfig struct |
| `pkg/cli/metadata/schema.go` | Project metadata | **EXTEND** - Add ID, Template, Agent fields |
| `pkg/cli/tui/sl_new.go` | TUI flow | **EXTEND** - Add stepTemplate, stepAgent |
| `pkg/cli/commands/bootstrap.go` | Project creation | **EXTEND** - Read template/agent, create project |
| `pkg/embedded/templates/manifest.yaml` | Template catalog | **EXTEND** - Add template definitions |

---

## Implementation Checklist

### Phase 1: Data Structures (1-2 hours)

**1.1 Create Template Model** (`pkg/models/template.go`):
```go
type TemplateDefinition struct {
    ID              string   `yaml:"id"`
    Name            string   `yaml:"name"`
    Description     string   `yaml:"description"`
    Characteristics []string `yaml:"characteristics,omitempty"`
    Path            string   `yaml:"path"`
    IsDefault       bool     `yaml:"is_default"`
}

func (t *TemplateDefinition) Validate() error { /* ... */ }
```

**1.2 Create Agent Model** (`pkg/models/agent.go`):
```go
type AgentConfig struct {
    ID          string `yaml:"id"`
    Name        string `yaml:"name"`
    Description string `yaml:"description"`
    ConfigDir   string `yaml:"config_dir,omitempty"`
}

func SupportedAgents() []AgentConfig { /* hardcoded list */ }
func GetAgentByID(id string) (*AgentConfig, error) { /* ... */ }
```

**1.3 Extend Metadata Schema** (`pkg/cli/metadata/schema.go`):
```go
import "github.com/google/uuid"

type ProjectInfo struct {
    ID        uuid.UUID `yaml:"id"`                 // NEW
    Name      string    `yaml:"name"`
    ShortCode string    `yaml:"short_code"`
    Template  string    `yaml:"template,omitempty"` // NEW
    Agent     string    `yaml:"agent,omitempty"`    // NEW
    Created   time.Time `yaml:"created"`
    Modified  time.Time `yaml:"modified"`
    Version   string    `yaml:"version"`
}
```

**1.4 Install Dependencies**:
```bash
go get github.com/google/uuid
```

---

### Phase 2: TUI Extension (2-3 hours)

**2.1 Extend TUI Model** (`pkg/cli/tui/sl_new.go`):
```go
const (
    stepProjectName = iota
    stepDirectory
    stepShortCode
    stepTemplate  // NEW
    stepAgent     // NEW
    stepPlaybook  // May deprecate
    stepConfirm
    stepDone
)

type Model struct {
    // Existing fields...
    step                  int
    textInput             textinput.Model
    answers               map[string]string

    // NEW fields
    templates             []models.TemplateDefinition
    selectedTemplateIndex int
    agents                []models.AgentConfig
    selectedAgentIndex    int
}
```

**2.2 Add Template Selection Handler**:
```go
// In Update() method
case stepTemplate:
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.Type {
        case tea.KeyUp:
            m.selectedTemplateIndex--
            if m.selectedTemplateIndex < 0 {
                m.selectedTemplateIndex = len(m.templates) - 1
            }
        case tea.KeyDown:
            m.selectedTemplateIndex++
            if m.selectedTemplateIndex >= len(m.templates) {
                m.selectedTemplateIndex = 0
            }
        case tea.KeyEnter:
            selectedTemplate := m.templates[m.selectedTemplateIndex]
            m.answers["template"] = selectedTemplate.ID
            m.step = stepAgent
        }
    }
```

**2.3 Add Template Selection View**:
```go
// In View() method
case stepTemplate:
    return m.renderTemplateSelection()

func (m Model) renderTemplateSelection() string {
    var b strings.Builder
    b.WriteString(titleStyle.Render("Select Project Template"))
    b.WriteString("\n\n")

    for i, tmpl := range m.templates {
        cursor := "  "
        radio := "○"
        if i == m.selectedTemplateIndex {
            cursor = "› "
            radio = "◉"
        }

        b.WriteString(fmt.Sprintf("%s%s %s\n", cursor, radio, tmpl.Name))
        b.WriteString(fmt.Sprintf("   %s\n", subtleStyle.Render(tmpl.Description)))

        if len(tmpl.Characteristics) > 0 {
            tech := strings.Join(tmpl.Characteristics, ", ")
            b.WriteString(fmt.Sprintf("   %s\n", subtleStyle.Render("Tech: "+tech)))
        }
        b.WriteString("\n")
    }

    b.WriteString(helpStyle.Render("[↑/↓ to navigate, Enter to select]"))
    return b.String()
}
```

**2.4 Repeat for Agent Selection** (same pattern as template selection).

---

### Phase 3: Template System (2-4 hours)

**3.1 Create Template Directories**:
```bash
mkdir -p templates/general-purpose
mkdir -p templates/full-stack
mkdir -p templates/batch-data
mkdir -p templates/realtime-workflow
mkdir -p templates/ml-image
mkdir -p templates/realtime-data
```

**3.2 Populate Template Structures**:

**General Purpose** (copy existing):
```bash
cp -r pkg/embedded/templates/specledger/* templates/general-purpose/
```

**Full-Stack** (create structure):
```
templates/full-stack/
├── backend/
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── api/
│   │   ├── models/
│   │   └── services/
│   ├── go.mod
│   └── README.md
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   ├── pages/
│   │   └── App.tsx
│   ├── package.json
│   └── tsconfig.json
├── tests/
├── specledger.yaml
└── README.md
```

**3.3 Update Manifest** (`pkg/embedded/templates/manifest.yaml`):
```yaml
version: 2.0.0

templates:
  - id: general-purpose
    name: General Purpose
    description: Default template for any project type
    characteristics: [Go, CLI, Single Binary]
    path: templates/general-purpose
    is_default: true

  - id: full-stack
    name: Full-Stack Application
    description: Go backend with TypeScript/React frontend
    characteristics: [Go, TypeScript, React, REST API]
    path: templates/full-stack
    is_default: false

  # Add other 4 templates...
```

**3.4 Update Embedded FS** (`pkg/embedded/templates.go`):
```go
//go:embed templates
var TemplatesFS embed.FS

func LoadTemplates() ([]models.TemplateDefinition, error) {
    data, err := TemplatesFS.ReadFile("templates/manifest.yaml")
    if err != nil {
        return nil, err
    }

    var manifest struct {
        Templates []models.TemplateDefinition `yaml:"templates"`
    }
    if err := yaml.Unmarshal(data, &manifest); err != nil {
        return nil, err
    }

    return manifest.Templates, nil
}
```

---

### Phase 4: Agent Configuration (1-2 hours)

**4.1 Create OpenCode Template Structure**:
```bash
mkdir -p templates/opencode/.opencode/commands
mkdir -p templates/opencode/.opencode/skills
```

**4.2 Port Claude Code Commands to OpenCode**:
```bash
cp -r .claude/commands/* templates/opencode/.opencode/commands/
cp -r .claude/skills/* templates/opencode/.opencode/skills/
```

**4.3 Create `opencode.json`**:
```json
{
  "$schema": "https://opencode.ai/config.json",
  "commands": {},
  "instructions": []
}
```

**4.4 Create `AGENTS.md` Template**:
```markdown
# {PROJECT_NAME} Development Guidelines

{TEMPLATE_SPECIFIC_GUIDELINES}

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
```

---

### Phase 5: Bootstrap Integration (2-3 hours)

**5.1 Update Bootstrap Command** (`pkg/cli/commands/bootstrap.go`):
```go
func CreateProject(answers map[string]string) error {
    // 1. Generate UUID
    projectID := uuid.New()

    // 2. Read template from answers
    templateID := answers["template"]
    template, err := loadTemplateByID(templateID)
    if err != nil {
        return fmt.Errorf("failed to load template: %w", err)
    }

    // 3. Copy template files
    if err := copyTemplateFiles(template, answers["project_dir"]); err != nil {
        return fmt.Errorf("failed to copy template: %w", err)
    }

    // 4. Read agent from answers
    agentID := answers["agent"]
    agent, err := models.GetAgentByID(agentID)
    if err != nil {
        return fmt.Errorf("failed to load agent: %w", err)
    }

    // 5. Create agent config directory (if needed)
    if agent.HasConfig() {
        if err := createAgentConfig(agent, answers["project_dir"]); err != nil {
            slog.Warn("failed to create agent config", "error", err)
            // Non-fatal: continue without agent files
        }
    }

    // 6. Create metadata with UUID, template, agent
    metadata := &metadata.ProjectMetadata{
        Version: "1.1.0",
        Project: metadata.ProjectInfo{
            ID:        projectID,
            Name:      answers["project_name"],
            ShortCode: answers["short_code"],
            Template:  templateID,
            Agent:     agentID,
            Created:   time.Now(),
            Modified:  time.Now(),
            Version:   "0.1.0",
        },
        // ... other fields
    }

    // 7. Write specledger.yaml
    metadataPath := filepath.Join(answers["project_dir"], "specledger.yaml")
    return metadata.Save(metadata, metadataPath)
}
```

---

### Phase 6: CLI Flags (1 hour)

**6.1 Add Flags** (`cmd/specledger/main.go` or `pkg/cli/commands/new.go`):
```go
newCmd.Flags().String("template", "", "Project template ID")
newCmd.Flags().String("agent", "", "Coding agent ID")
newCmd.Flags().Bool("list-templates", false, "List available templates")
```

**6.2 Implement `--list-templates`**:
```go
if listTemplates, _ := cmd.Flags().GetBool("list-templates"); listTemplates {
    templates, err := embedded.LoadTemplates()
    if err != nil {
        return err
    }

    fmt.Println("Available Project Templates:\n")
    for _, tmpl := range templates {
        fmt.Printf("  %s%s\n", tmpl.ID, ternary(tmpl.IsDefault, " (default)", ""))
        fmt.Printf("    %s\n", tmpl.Description)
        if len(tmpl.Characteristics) > 0 {
            fmt.Printf("    Tech: %s\n", strings.Join(tmpl.Characteristics, ", "))
        }
        fmt.Println()
    }

    fmt.Println("Use: sl new --template <id>")
    return nil
}
```

**6.3 Implement Non-Interactive Mode**:
```go
if !isatty.IsTerminal(os.Stdin.Fd()) {
    // Non-interactive mode: require all flags
    requiredFlags := []string{"project-name", "project-dir", "short-code", "template", "agent"}
    for _, flag := range requiredFlags {
        if !cmd.Flags().Changed(flag) {
            return fmt.Errorf("missing required flag: --%s", flag)
        }
    }

    // Collect answers from flags
    answers := map[string]string{
        "project_name": cmd.Flags().GetString("project-name"),
        "project_dir":  cmd.Flags().GetString("project-dir"),
        "short_code":   cmd.Flags().GetString("short-code"),
        "template":     cmd.Flags().GetString("template"),
        "agent":        cmd.Flags().GetString("agent"),
    }

    // Validate template and agent
    if _, err := loadTemplateByID(answers["template"]); err != nil {
        return fmt.Errorf("unknown template: %s", answers["template"])
    }
    if _, err := models.GetAgentByID(answers["agent"]); err != nil {
        return fmt.Errorf("unknown agent: %s", answers["agent"])
    }

    return CreateProject(answers)
}

// Interactive mode: run TUI
```

---

### Phase 7: Testing (3-4 hours)

**7.1 Install Testing Dependency**:
```bash
go get github.com/charmbracelet/x/exp/teatest@latest
```

**7.2 Unit Tests** (`pkg/cli/tui/sl_new_test.go`):
```go
func TestTemplateNavigation(t *testing.T) {
    m := InitialModel("/tmp")
    m.step = stepTemplate

    // Test down navigation
    updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
    model := updatedModel.(Model)
    assert.Equal(t, 1, model.selectedTemplateIndex)

    // Test up navigation (wrap)
    m.selectedTemplateIndex = 0
    updatedModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
    model = updatedModel.(Model)
    assert.Equal(t, len(m.templates)-1, model.selectedTemplateIndex)
}

func TestTemplateSelection(t *testing.T) {
    m := InitialModel("/tmp")
    m.step = stepTemplate
    m.selectedTemplateIndex = 1 // Full-Stack

    updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
    model := updatedModel.(Model)

    assert.Equal(t, "full-stack", model.answers["template"])
    assert.Equal(t, stepAgent, model.step)
}
```

**7.3 Integration Tests** (`tests/integration/bootstrap_tui_test.go`):
```go
func TestInteractiveTemplateSelection(t *testing.T) {
    m := tui.InitialModel("/tmp")
    tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(80, 24))

    // Navigate through steps
    tm.Type("test-project")
    tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
    tm.Type("/tmp/test")
    tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
    tm.Type("tp")
    tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

    // Navigate to full-stack template
    tm.Send(tea.KeyMsg{Type: tea.KeyDown})
    tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

    // Select Claude Code agent (default)
    tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

    // Confirm
    tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

    // Verify final model
    fm := tm.FinalModel(t, teatest.WithFinalTimeout(time.Second))
    finalModel := fm.(tui.Model)

    assert.Equal(t, "full-stack", finalModel.answers["template"])
    assert.Equal(t, "claude-code", finalModel.answers["agent"])
}
```

**7.4 E2E Tests** (`tests/integration/bootstrap_test.go`):
```go
func TestNonInteractiveFullStack(t *testing.T) {
    tempDir := t.TempDir()
    slBinary := buildSLBinary(t, tempDir)

    cmd := exec.Command(slBinary, "new",
        "--project-name", "test-app",
        "--project-dir", tempDir+"/test-app",
        "--short-code", "ta",
        "--template", "full-stack",
        "--agent", "opencode")

    output, err := cmd.CombinedOutput()
    require.NoError(t, err, "sl new failed: %s", output)

    // Verify structure
    assert.DirExists(t, tempDir+"/test-app/backend")
    assert.DirExists(t, tempDir+"/test-app/frontend")
    assert.DirExists(t, tempDir+"/test-app/.opencode")
    assert.FileExists(t, tempDir+"/test-app/specledger.yaml")

    // Verify metadata
    data, err := os.ReadFile(tempDir + "/test-app/specledger.yaml")
    require.NoError(t, err)

    var metadata struct {
        Project struct {
            ID       string `yaml:"id"`
            Template string `yaml:"template"`
            Agent    string `yaml:"agent"`
        } `yaml:"project"`
    }
    require.NoError(t, yaml.Unmarshal(data, &metadata))

    assert.NotEmpty(t, metadata.Project.ID)
    assert.Equal(t, "full-stack", metadata.Project.Template)
    assert.Equal(t, "opencode", metadata.Project.Agent)
}
```

---

## Testing the Feature

### Manual Testing

**1. Interactive Mode**:
```bash
go run cmd/specledger/main.go new
# Navigate through TUI with arrow keys
# Verify template descriptions display correctly
# Verify agent descriptions display correctly
# Confirm project creation
```

**2. Non-Interactive Mode**:
```bash
go run cmd/specledger/main.go new \
  --project-name test \
  --project-dir /tmp/test \
  --short-code t \
  --template full-stack \
  --agent opencode

# Verify output: "Project 'test' created at /tmp/test"
ls /tmp/test/
# Verify: backend/, frontend/, .opencode/, specledger.yaml
```

**3. List Templates**:
```bash
go run cmd/specledger/main.go new --list-templates
# Verify all 6 templates listed with descriptions
```

**4. Error Handling**:
```bash
go run cmd/specledger/main.go new --template invalid
# Verify: "Error: unknown template: invalid"
# Verify: Available templates listed
```

### Automated Testing

```bash
# Run all tests
go test ./...

# Run TUI tests only
go test -v ./pkg/cli/tui/

# Run integration tests only
go test -v ./tests/integration/

# Update golden files
go test ./... -update

# Run with race detector
go test -race ./...
```

---

## Common Issues & Solutions

### Issue: UUID not marshaling to YAML
**Solution**: Ensure `github.com/google/uuid` is imported. UUID type has built-in `MarshalText()`.

### Issue: Template not found in embedded FS
**Solution**: Run `go generate` or rebuild to embed templates. Check `//go:embed` directive.

### Issue: TUI not rendering correctly
**Solution**: Set `lipgloss.SetColorProfile(termenv.Ascii)` in init() for tests. Check terminal size with `teatest.WithInitialTermSize(80, 24)`.

### Issue: Agent config not created
**Solution**: Check `agent.HasConfig()` returns true. Verify agent config directory exists in templates.

### Issue: Non-interactive mode requires TTY
**Solution**: Use `isatty.IsTerminal(os.Stdin.Fd())` to detect non-interactive mode. Require all flags when false.

---

## Performance Tips

1. **Lazy Load Templates**: Only load templates when needed (InitialModel or --list-templates)
2. **Cache Manifest**: Parse manifest once, cache in memory
3. **Embed at Build Time**: Use `//go:embed` to compile templates into binary
4. **Skip Validation in Hot Path**: Validate templates at build time, not runtime

---

## Next Steps After Implementation

1. **Update Documentation**: Add template catalog to README.md
2. **Update Agent Scripts**: Extend `.specledger/scripts/bash/update-agent-context.sh` to support OpenCode
3. **Create Template READMEs**: Document each template's structure and usage
4. **Add Metrics**: Log template/agent selection for usage analytics
5. **Consider P2 Features**: Template preview (`--preview-template`)
6. **Consider P3 Features**: Template recommendations (`--recommend`)

---

## Resources

- **Research**: [research.md](./research.md) - Detailed technical decisions
- **Data Model**: [data-model.md](./data-model.md) - Entity relationships and schemas
- **CLI Contract**: [contracts/cli-interface.md](./contracts/cli-interface.md) - Full interface specification
- **Plan**: [plan.md](./plan.md) - Implementation phases and architecture

---

## Estimated Timeline

| Phase | Estimated Time | Dependencies |
|-------|----------------|--------------|
| Phase 1: Data Structures | 1-2 hours | None |
| Phase 2: TUI Extension | 2-3 hours | Phase 1 |
| Phase 3: Template System | 2-4 hours | Phase 1 |
| Phase 4: Agent Configuration | 1-2 hours | Phase 3 |
| Phase 5: Bootstrap Integration | 2-3 hours | Phases 1-4 |
| Phase 6: CLI Flags | 1 hour | Phase 5 |
| Phase 7: Testing | 3-4 hours | Phases 1-6 |
| **Total** | **12-19 hours** | |

**Recommended Approach**: Implement in order, test incrementally. Phases 3 and 4 can be done in parallel.

---

## Success Criteria

- [ ] All 6 templates create valid project structures
- [ ] All 3 agents create correct configuration directories
- [ ] Interactive TUI flow completes in <60 seconds
- [ ] Non-interactive mode works with all flags
- [ ] Backward compatibility: general-purpose + claude-code = old behavior
- [ ] All tests pass (unit, integration, E2E)
- [ ] Code coverage ≥80% for new code
- [ ] No linter warnings
- [ ] Documentation updated

**Feature complete when**: A user can run `sl new`, select "Full-Stack Application" and "OpenCode", and get a working project with backend/, frontend/, and .opencode/ directories.
