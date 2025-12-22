<!--
SYNC IMPACT REPORT

Version Change: [No prior version] → 1.0.0

Modified Principles:
- Initial constitution creation with 7 core principles

Added Sections:
- Core Principles (7 principles)
- Performance & Scalability Standards
- Development Workflow & Quality Gates
- Governance

Removed Sections:
- None (initial version)

Templates Requiring Updates:
✅ plan-template.md - Constitution Check section aligns with principles
✅ spec-template.md - User scenarios and requirements align with testing standards
✅ tasks-template.md - Task structure supports test-first workflow
✅ agent-file-template.md - No updates required (general template)

Follow-up TODOs:
- None
-->

# SpecLedger Constitution

## Core Principles

### I. Specification-First Development

Every feature begins with a complete specification before any code is written. Specifications MUST be:
- Independently testable with clear acceptance criteria
- Prioritized by user value (P1, P2, P3+)
- Traceable from requirement to implementation via Beads issue tracker
- Version-controlled and peer-reviewed before planning begins

**Rationale**: Specifications are first-class artifacts, not disposable chat output. This ensures shared understanding, prevents rework, and creates an audit trail from intent to implementation.

### II. Test-First Development (NON-NEGOTIABLE)

Test-Driven Development is mandatory for all implementations:
- Contract tests MUST be written before implementation code
- Integration tests MUST be written for user journeys before feature work
- All tests MUST fail initially (red), then pass after implementation (green)
- Tests are approval gates: User approves tests → Tests fail → Then implement

**Rationale**: Tests define the contract and prevent regression. Writing tests first ensures testability and forces clarity on expected behavior before investing in implementation.

### III. Code Quality & Maintainability

Code MUST be:
- Self-documenting with clear naming conventions
- Linted and formatted using project-standard tools
- Free from complexity unless justified (see Complexity Tracking in plans)
- Reviewed for security vulnerabilities (OWASP Top 10 minimum)
- Minimal in scope: Only implement what is specified, no gold-plating

**Rationale**: Maintainable code reduces technical debt and cognitive load. Simple solutions are easier to debug, extend, and hand off to team members or AI agents.

### IV. User Experience Consistency

User-facing features MUST:
- Follow established interaction patterns within the project
- Provide clear, actionable error messages
- Support accessibility standards (WCAG 2.1 AA minimum where applicable)
- Degrade gracefully when dependencies fail
- Document user flows in acceptance scenarios (Given/When/Then format)

**Rationale**: Consistent UX reduces training time and user errors. Specifications enforce consistency by requiring explicit user scenarios before implementation.

### V. Performance Requirements

Every feature specification MUST define:
- Target performance metrics (e.g., <200ms p95 latency, 1000 req/s throughput)
- Performance constraints (e.g., <100MB memory, offline-capable)
- Scale expectations (e.g., 10k concurrent users, 1M records)
- Acceptance criteria MUST include measurable performance outcomes

Performance regressions blocking merge MUST be justified and documented.

**Rationale**: Performance is a feature, not an afterthought. Defining metrics upfront prevents costly rewrites and ensures user satisfaction under load.

### VI. Observability & Debuggability

All implementations MUST:
- Log key decision points and errors using structured logging
- Emit metrics for performance-critical operations
- Provide clear error messages with context (no bare exceptions)
- Support tracing through distributed systems where applicable
- Document logging patterns in implementation plans

**Rationale**: Observable systems are debuggable systems. Structured logs and metrics enable root cause analysis without guesswork, especially in production.

### VII. Issue Tracking Discipline

All work MUST be tracked in Beads (`bd`) with:
- Epic → Feature → Task hierarchy maintained
- Dependencies explicitly defined via `bd dep add`
- Status updates via `bd update` (open, in progress, blocked, closed)
- Progress comments via `bd comments add` for context preservation
- Queries MUST use filters to avoid context exhaustion (see CLAUDE.md)

**Rationale**: Persistent issue tracking enables context preservation across sessions, traceability from spec to code, and safe parallelization of work by humans and AI agents.

## Performance & Scalability Standards

### Baseline Targets (Override in Feature Specs if Needed)

- **API Response Time**: p95 < 200ms, p99 < 500ms
- **Database Queries**: < 100ms for single-record reads, < 1s for complex queries
- **Memory Usage**: Backend services < 512MB baseline, < 2GB under load
- **Frontend Load Time**: Time to Interactive (TTI) < 3s on 3G connection
- **Throughput**: 1000 requests/second per service instance minimum

### Scalability Requirements

- Horizontal scaling MUST be supported (stateless services preferred)
- Database queries MUST use indexes for all WHERE/JOIN clauses
- Batch operations MUST be paginated with configurable limits
- Long-running tasks MUST be async with progress tracking

### Performance Testing Gates

- Load tests MUST validate throughput targets before production deployment
- Performance regressions > 20% trigger rollback or justification requirement
- Database query plans MUST be reviewed for N+1 patterns

## Development Workflow & Quality Gates

### Gate 1: Specification Approval

**Required Artifacts**:
- Complete spec.md with prioritized user stories (P1, P2, P3+)
- Independently testable acceptance scenarios (Given/When/Then)
- Functional requirements with unique IDs (FR-001, FR-002, etc.)
- Success criteria with measurable outcomes (SC-001, SC-002, etc.)
- Beads epic created with `bd create` and linked to spec

**Approval**: Product owner or tech lead sign-off required

### Gate 2: Planning & Design Approval

**Required Artifacts**:
- plan.md with technical context, structure decision, and constitution check
- research.md documenting alternatives considered
- data-model.md defining entities and relationships
- contracts/ directory with API/interface definitions
- Complexity violations justified in plan.md table

**Approval**: Architect or senior engineer review required

### Gate 3: Task Generation & Dependency Mapping

**Required Artifacts**:
- Beads task graph created via `bd create` with correct hierarchy
- Dependencies mapped via `bd dep add` (blocks, parent-child, related)
- Tasks organized by user story with independent testability preserved
- Foundational phase clearly separated from user story work

**Approval**: Tasks review by lead, ensure no circular dependencies

### Gate 4: Implementation & Testing

**Required Steps**:
1. Write contract/integration tests (MUST fail initially)
2. Implement minimal code to pass tests
3. Refactor for clarity and performance
4. Run linters, formatters, security scans
5. Update Beads tasks via `bd update` and `bd comments add`

**Approval**: Peer code review + all tests passing + no security warnings

### Gate 5: Performance Validation (If Spec Defines Metrics)

**Required Steps**:
- Load tests executed against performance targets
- Metrics collection validated (logs, traces, metrics endpoints)
- Performance regression checks pass (< 20% deviation from baseline)

**Approval**: Performance test results attached to PR

### Gate 6: Deployment

**Required Steps**:
- User story independently deployable (verified via quickstart.md)
- Rollback plan documented
- Observability validated (logs flowing, metrics emitting)
- Close Beads tasks via `bd close` with completion context

**Approval**: Deployment lead or automated CD pipeline

## Governance

### Amendment Procedure

1. Propose amendment via pull request to `.specify/memory/constitution.md`
2. Update `CONSTITUTION_VERSION` per semantic versioning rules:
   - **MAJOR**: Backward incompatible governance/principle removals or redefinitions
   - **MINOR**: New principle/section added or materially expanded guidance
   - **PATCH**: Clarifications, wording, typo fixes, non-semantic refinements
3. Update `LAST_AMENDED_DATE` to current date (ISO format YYYY-MM-DD)
4. Document changes in Sync Impact Report (HTML comment at top of file)
5. Propagate changes to dependent templates (plan, spec, tasks, agent-file)
6. Require approval from 2+ senior engineers or architect

### Compliance Review

- All PRs MUST verify compliance with applicable principles (reference in PR description)
- Quarterly audits of closed features against constitution requirements
- Constitution violations MUST be justified in plan.md Complexity Tracking table
- Repeat violations trigger process improvement discussions

### Template Synchronization

When constitution changes, the following files MUST be updated:
- `.specify/templates/plan-template.md` (Constitution Check section)
- `.specify/templates/spec-template.md` (Requirements and success criteria)
- `.specify/templates/tasks-template.md` (Task structure and phases)
- `.specify/templates/agent-file-template.md` (If new coding standards added)

Use `.specify/scripts/bash/update-agent-context.sh` to regenerate agent guidance after amendments.

**Version**: 1.0.0 | **Ratified**: 2025-12-22 | **Last Amended**: 2025-12-22
