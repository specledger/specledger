---
description: Complete codebase audit - quick reconnaissance scan followed by deep module analysis with dependency graphs and JSON cache generation
---

## User Input

```text
$ARGUMENTS
```

Optional flags:
- `--format json|markdown`: Output format (default: markdown)
- `--scope [PATH]`: Only analyze specific directory
- `--module [NAME]`: Analyze specific module only (deep phase)
- `--force`: Re-analyze even if cache exists

## Purpose

Perform a complete two-phase audit of a codebase:
1. **Quick Reconnaissance** (~15 minutes): Identify tech stack, architecture pattern, and entry points
2. **Deep Module Analysis** (~30+ minutes): Discover logical modules, extract key functions, data models, and build dependency graphs

## When to Use

- First encounter with an unfamiliar codebase
- Quick project overview before starting work
- Validating tech stack assumptions
- Preparing context for `/specledger.adopt`
- Building comprehensive codebase documentation
- When you need detailed module understanding

---

# PART 1: Quick Reconnaissance Scan

## Phase 1: Tech Stack Detection (5 minutes)

1. **Detect Language & Framework**
   ```bash
   # Scan root for identifying files
   ls -la | grep -E "package.json|go.mod|requirements.txt|pyproject.toml|pom.xml|build.gradle|composer.json|Cargo.toml|Gemfile"
   ```

   Apply detection patterns from `.specledger/templates/partials/language-detection.md`:
   - **JavaScript/TypeScript**: `package.json`, `tsconfig.json`
   - **Python**: `requirements.txt`, `pyproject.toml`
   - **Go**: `go.mod`, `go.sum`
   - **Java/Kotlin**: `pom.xml`, `build.gradle`
   - **PHP**: `composer.json`
   - **Rust**: `Cargo.toml`
   - **Ruby**: `Gemfile`

2. **Extract Framework Information**
   ```bash
   # For Node.js projects
   cat package.json | jq '{name, dependencies, devDependencies, scripts}' 2>/dev/null

   # For Go projects
   cat go.mod | grep "^module\|^require" 2>/dev/null
   ```

## Phase 2: Directory Structure Mapping (5 minutes)

1. **Get Project Tree**
   ```bash
   tree -L 3 -d -I 'node_modules|.git|dist|build|vendor|target|__pycache__|coverage|.next'
   ```

2. **Identify Architecture Pattern**
   - **Monorepo**: `packages/`, `apps/`, `libs/`, `pnpm-workspace.yaml`
   - **Clean Architecture**: `domain/`, `infrastructure/`, `application/`
   - **MVC**: `controllers/`, `models/`, `views/`
   - **Feature-Sliced**: `features/`, `entities/`, `shared/`
   - **Microservices**: Multiple `cmd/` or `services/` directories

3. **Count Files by Type**
   ```bash
   find . -type f -name "*.ts" -o -name "*.tsx" -o -name "*.go" -o -name "*.py" 2>/dev/null | wc -l
   ```

## Phase 3: Entry Point Detection (5 minutes)

1. **Find Entry Points**
   ```bash
   find . -name "main.go" -o -name "main.ts" -o -name "index.ts" -o -name "app.py" -o -name "server.js" 2>/dev/null
   ```

2. **Check Build Scripts**
   ```bash
   cat package.json | jq '.scripts' 2>/dev/null
   grep "^[a-z].*:" Makefile 2>/dev/null
   ```

## Quick Audit Output

Generate a quick overview and save to `scripts/audit-quick.json`:

```json
{
  "project_name": "...",
  "tech_stack": {
    "language": "...",
    "framework": "...",
    "build_tool": "..."
  },
  "architecture": {
    "pattern": "...",
    "key_directories": []
  },
  "entry_points": [],
  "file_count": 0,
  "audit_type": "quick"
}
```

---

# PART 2: Deep Module Analysis

## Phase 4: Load Quick Audit Results (2 minutes)

1. **Use Quick Audit Context**
   - Project name and type
   - Primary language and framework
   - Key directories for analysis

## Phase 5: Module Discovery (15-30 minutes per module)

Apply clustering strategies from `.specledger/templates/partials/module-clustering.md`:

1. **If Clear Structure Exists** (e.g., `src/modules/`, `packages/`):
   - Each top-level directory = one module
   - Use directory name as module ID

2. **If Unstructured** (flat files):
   - Apply filename prefix clustering
   - Apply import/dependency analysis
   - Apply route/URL clustering
   - Apply database table clustering

3. **For Each Module, Extract:**

   **Go Projects:**
   ```bash
   grep -r "^package " [MODULE_PATH] | cut -d: -f2 | sort -u
   grep -r "^func [A-Z]" [MODULE_PATH]
   grep -r "^type [A-Z]" [MODULE_PATH]
   grep -A 5 "^type .* struct" [MODULE_PATH]
   ```

   **TypeScript Projects:**
   ```bash
   grep -r "^export " [MODULE_PATH] | head -30
   grep -r "export.*function\|export.*Component" [MODULE_PATH]
   find [MODULE_PATH] -path "*/api/*" -name "*.ts"
   grep -r "^export type\|^export interface" [MODULE_PATH]
   ```

   **Python Projects:**
   ```bash
   grep -r "^class " [MODULE_PATH]
   grep -r "@app.get\|@app.post\|path(" [MODULE_PATH]
   grep -r "@dataclass\|BaseModel" [MODULE_PATH]
   ```

## Phase 6: Dependency Graph Building (10 minutes)

1. **Extract Imports**
   ```bash
   # Go
   grep -r "^import " [MODULE_PATH] | grep -o '".*"' | sort -u

   # TypeScript
   grep -r "^import .* from" [MODULE_PATH] | grep -o "from ['\"].*['\"]" | sort -u

   # Python
   grep -r "^import \|^from " [MODULE_PATH] | sort -u
   ```

2. **Identify Integration Points**
   - Database access patterns
   - External API calls
   - Message queues, event buses
   - File system operations

3. **Detect Cross-Cutting Concerns**
   - Authentication/Authorization
   - Logging, Monitoring
   - Error handling patterns
   - Configuration management

## Phase 7: Generate JSON Cache (5 minutes)

Create `scripts/audit-cache.json`:

```json
{
  "metadata": {
    "project_name": "...",
    "project_type": "monorepo|single-package|microservices",
    "language": "typescript|go|python|...",
    "framework": "nextjs|express|gin|fastapi|...",
    "analyzed_at": "ISO timestamp",
    "total_loc": 0,
    "file_count": 0
  },
  "global_context": {
    "architecture_style": "...",
    "api_pattern": "REST|GraphQL|gRPC",
    "auth_pattern": "...",
    "database": "...",
    "common_patterns": []
  },
  "modules": [
    {
      "id": "module-id",
      "name": "Human Readable Name",
      "description": "What this module does",
      "type": "core-domain|infrastructure|api|integration|ui|utility",
      "paths": ["path/to/files"],
      "entry_point": "main file",
      "loc": 0,
      "key_functions": [],
      "data_models": [],
      "api_contracts": [],
      "dependencies": []
    }
  ]
}
```

---

## Final Output

### Markdown Output (default)

```markdown
# Complete Audit: [PROJECT_NAME]

## Tech Stack
- **Language**: [Primary language]
- **Framework**: [Framework name]
- **Build Tool**: [npm/yarn/go/etc]

## Architecture
- **Pattern**: [Monorepo/MVC/Clean/etc]
- **Structure**: [Brief description]

## Entry Points
- [List of main entry files]

## File Statistics
- Total source files: [count]
- Primary directories: [list]

---

## Modules Discovered: [COUNT]

| Module | Type | Files | LOC | Key Functions |
|--------|------|-------|-----|---------------|
| [name] | [type] | [n] | [loc] | [count] |

## Dependency Graph
[Module A] → [Module B] → [Module C]

## Next Steps
Run `/specledger.adopt --module-id [ID] --from-audit` to create specs.
```

### JSON Output (--format json)

Both `scripts/audit-quick.json` and `scripts/audit-cache.json` are generated.

## Error Handling

- **No recognizable project files**: "Cannot detect project type. Ensure you're in a project root directory."
- **Empty directory**: "No source files found in the specified scope."
- **Permission denied**: "Cannot read some directories. Check file permissions."
- **No modules found**: "Could not identify module boundaries. Project may be too flat."
- **Analysis timeout**: "Module [X] analysis exceeded time limit. Skipping detailed analysis."

## Examples

```bash
# Complete audit (quick + deep)
/specledger.audit

# Audit specific subdirectory
/specledger.audit --scope src/api

# Output as JSON for scripting
/specledger.audit --format json

# Analyze specific module only
/specledger.audit --module user-management

# Force re-analysis
/specledger.audit --force
```
