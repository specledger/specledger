# Implementation Plan: Generate .claudeignore

**Spec**: 322-generate-claudeignore
**Feature Branch**: `322-generate-claudeignore`

## Default Ignore Patterns

The generated .claudeignore file should include patterns for:

**Universal**:
- `.git/`
- `.gitignore`
- `.env`
- `.env.local`
- `.DS_Store`
- `*.log`

**Node.js**:
- `node_modules/`
- `dist/`
- `build/`
- `.next/`
- `coverage/`

**Python**:
- `__pycache__/`
- `*.pyc`
- `venv/`
- `.venv/`
- `dist/`
- `build/`

**Go**:
- `vendor/`
- `bin/`
- `*.o`

**General Build Artifacts**:
- `*.min.js`
- `*.min.css`
- `*.map`

## File Generation Logic

1. Check if .claudeignore exists
2. If not, create from embedded static template
3. If yes, preserve existing file and queue improvement suggestions
4. Agent instructions are generated to explore codebase and enhance patterns
5. Log action to user

## Agent Enhancement Instructions

When generating or updating .claudeignore, the agent should:
1. Detect project type (language, framework, build tools)
2. Explore codebase for build artifacts, dependency directories, generated files
3. Add project-specific patterns beyond the static template
4. For existing files: suggest additive improvements only (no removals without user approval)

## Integration Points

- `sl init` command handler
- `sl new` command handler
- Project scaffolding templates

## Testing Strategy

### Unit Tests
- Test pattern matching logic
- Test file creation with various project states
- Test preservation of existing files

### Integration Tests
- Test `sl init` with and without existing .claudeignore
- Test `sl new` project generation
- Test pattern effectiveness with sample files

### Manual Testing
- Verify .claudeignore is created in new projects
- Verify existing files are not overwritten
- Verify patterns work as expected with Claude

## Documentation

- Add .claudeignore format documentation to user guide
- Include examples for common project types
- Document how to customize patterns
- Add troubleshooting section

## Future Considerations

- **Multi-Agent Support**: Future versions may support generating ignore files for other AI tools (e.g., `.cursorignore`, `.aiderignore`). This would integrate with a config-based approach where the agent reads project configuration and generates appropriate ignore files per configured agent.
- **Template Versioning**: Track template versions to enable upgrade prompts when newer templates are available.
