# SDD Framework Overview

SpecLedger supports multiple Specification-Driven Development (SDD) frameworks.

## Supported Frameworks

### Spec Kit
[GitHub Spec Kit](https://github.com/github/spec-kit)

Structured, phase-gated workflow with specifications, plans, and tasks.

**Best for**: Teams that want structured development phases

**Features**:
- Specification-first development
- Implementation planning
- Task generation with Beads
- Phase gates and quality checks

### OpenSpec
[OpenSpec](https://github.com/fission-ai/openspec)

Lightweight, iterative specification workflow.

**Best for**: Teams that prefer flexibility and iteration

**Features**:
- Lightweight specification format
- AI-friendly content structure
- Minimal process overhead
- Quick iteration cycles

## Framework Integration

When you create a project with `sl new --framework <name>`:

1. Framework is installed via mise
2. Framework is initialized in your project
3. Configuration is updated in `specledger.yaml`
4. Project is ready for SDD workflows

## Using Frameworks After Setup

Once your project is created, you use the framework's tools directly:

```bash
# With Spec Kit
specify spec create "User Authentication"
specify plan create
specify tasks generate

# With OpenSpec
openspec spec create "User Authentication"
```

## Framework Detection

SpecLedger automatically detects the framework type when adding dependencies:

```bash
sl deps add git@github.com:org/api-spec
# Detects: Spec Kit, OpenSpec, both, or none
# Sets up AI import path automatically
```

## Choosing the Right Framework

| Need | Recommendation |
|------|----------------|
| Structured phases with gates | Spec Kit |
| Fast iteration and flexibility | OpenSpec |
| Maximum flexibility | Both (choose per project) |
| Custom SDD process | None (use SpecLedger only) |
