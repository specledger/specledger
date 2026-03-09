# sl Audit Skill

## Overview

Codebase reconnaissance skill for understanding project structure, tech stack, and module organization. This skill provides patterns for efficient codebase exploration without AI orchestration overhead.

**Layer**: L3 (Skill) - Passive context injection
**Use when**: You need to understand an unfamiliar codebase or validate architecture assumptions

## When to Load

Load this skill when:
- First encounter with an unfamiliar codebase
- Need to understand project structure before implementation
- Validating tech stack assumptions
- Preparing context for feature planning
- Building comprehensive codebase documentation

**Don't load when**:
- Already familiar with the codebase
- Only need simple file searches
- Working on isolated, well-understood modules

## Key Concepts

### Two-Phase Audit Strategy

1. **Quick Reconnaissance** (~15 min)
   - Tech stack detection via config files
   - Directory structure mapping
   - Entry point identification
   - Architecture pattern recognition

2. **Deep Module Analysis** (~30+ min per module)
   - Logical module discovery
   - Key function extraction
   - Data model identification
   - Dependency graph building

### Architecture Pattern Recognition

| Pattern | Key Indicators |
|---------|---------------|
| Monorepo | `packages/`, `apps/`, `libs/`, `pnpm-workspace.yaml` |
| Clean Architecture | `domain/`, `infrastructure/`, `application/` |
| MVC | `controllers/`, `models/`, `views/` |
| Feature-Sliced | `features/`, `entities/`, `shared/` |
| Microservices | Multiple `cmd/` or `services/` directories |

### Module Clustering Strategies

When codebase lacks clear structure:
1. **Filename prefix clustering** - Group by common prefixes
2. **Import/dependency analysis** - Follow import chains
3. **Route/URL clustering** - Group by API routes
4. **Database table clustering** - Group by data entities

## Decision Patterns

### Choosing Audit Depth

```
IF first_time_in_codebase:
    RUN full_audit (quick + deep)
ELIF validating_specific_assumption:
    RUN quick_reconnaissance ONLY
ELIF need_module_details:
    RUN deep_module_analysis FOR target_module
```

### Output Format Selection

| Format | When to Use |
|--------|-------------|
| `markdown` | Human review, documentation |
| `json` | Scripting, automation, caching |

### Cache Strategy

- **Quick audit**: Cache to `scripts/audit-quick.json`
- **Deep audit**: Cache to `scripts/audit-cache.json`
- **Force re-analysis**: Use `--force` flag

## Detection Commands

### Tech Stack Detection

```bash
# Identify language/framework
ls -la | grep -E "package.json|go.mod|requirements.txt|pyproject.toml|pom.xml|Cargo.toml"

# JavaScript/TypeScript
cat package.json | jq '{name, dependencies, devDependencies, scripts}'

# Go
cat go.mod | grep "^module\|^require"

# Python
cat requirements.txt
cat pyproject.toml
```

### Structure Analysis

```bash
# Directory tree (3 levels)
tree -L 3 -d -I 'node_modules|.git|dist|build|vendor|target|__pycache__|coverage'

# File counts
find . -type f -name "*.ts" -o -name "*.go" -o -name "*.py" | wc -l
```

### Entry Point Discovery

```bash
# Find main files
find . -name "main.go" -o -name "main.ts" -o -name "index.ts" -o -name "app.py"
```

## Module Extraction Patterns

### Go Projects

```bash
grep -r "^package " [MODULE_PATH] | cut -d: -f2 | sort -u
grep -r "^func [A-Z]" [MODULE_PATH]
grep -r "^type [A-Z]" [MODULE_PATH]
grep -A 5 "^type .* struct" [MODULE_PATH]
```

### TypeScript Projects

```bash
grep -r "^export " [MODULE_PATH] | head -30
grep -r "export.*function\|export.*Component" [MODULE_PATH]
find [MODULE_PATH] -path "*/api/*" -name "*.ts"
grep -r "^export type\|^export interface" [MODULE_PATH]
```

### Python Projects

```bash
grep -r "^class " [MODULE_PATH]
grep -r "@app.get\|@app.post\|path(" [MODULE_PATH]
grep -r "@dataclass\|BaseModel" [MODULE_PATH]
```

## Dependency Graph Building

```bash
# Go imports
grep -r "^import " [MODULE_PATH] | grep -o '".*"' | sort -u

# TypeScript imports
grep -r "^import .* from" [MODULE_PATH] | grep -o "from ['\"].*['\"]" | sort -u

# Python imports
grep -r "^import \|^from " [MODULE_PATH] | sort -u
```

## JSON Output Schema

```json
{
  "metadata": {
    "project_name": "...",
    "project_type": "monorepo|single-package|microservices",
    "language": "typescript|go|python",
    "framework": "nextjs|express|gin|fastapi",
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

## Common Pitfalls

1. **Over-auditing**: Start with quick reconnaissance, go deep only when needed
2. **Ignoring cache**: Reuse cached results to save time
3. **Wrong scope**: Ensure you're in project root for accurate detection
4. **Missing dependencies**: Always check cross-cutting concerns (auth, logging, config)

## CLI Reference

> **Note**: The `sl audit` command doesn't exist yet. This skill provides patterns for manual codebase reconnaissance using standard tools.

### Essential Detection Commands

| Action | Command |
|--------|---------|
| Tech stack detection | `ls -la | grep -E "package.json|go.mod|requirements.txt"` |
| Directory tree | `tree -L 3 -d -I "node_modules|.git|dist|build"` |
| Entry points | `find . -name "main.go" -o -name "main.ts" -o -name "app.py"` |

> **Full syntax**: See `tree --help`, `find --help`, `grep --help` for complete command reference.

## Troubleshooting

**If audit cache is stale**:
```bash
rm scripts/audit-quick.json scripts/audit-cache.json
```

And re-run the detection commands.

**If no project structure detected**:
- Verify you're in the project root directory
- Check for hidden config files (e.g., `.github/workflows/`)
- Look for `README.md` for project setup instructions

**If dependency graph is incomplete**:
- Some imports may be dynamically loaded
- Check for reflection-based dependency injection
- Look for configuration files that specify module relationships

## CLI Reference

> **Note**: The `sl audit` command doesn't exist yet. This skill provides patterns for manual codebase reconnaissance using standard tools.

### Essential Detection Commands

| Action | Command |
|--------|---------|
| Tech stack detection | `ls -la | grep -E "package.json|go.mod|requirements.txt"` |
| Directory tree | `tree -L 3 -d -I "node_modules|.git|dist|build"` |
| Entry points | `find . -name "main.go" -o -name "main.ts" -o -name "app.py"` |

> **Full syntax**: See `tree --help`, `find --help`, `grep --help` for complete command reference.

## Troubleshooting

**If audit cache is stale**:
```bash
rm scripts/audit-quick.json scripts/audit-cache.json
```

And re-run the detection commands.

**If no project structure detected**:
- Verify you're in the project root directory
- Check for hidden config files (e.g., `.github/workflows/`)
- Look for `README.md` for project setup instructions

**If dependency graph is incomplete**:
- Some imports may be dynamically loaded
- Check for reflection-based dependency injection
- Look for configuration files that specify module relationships
