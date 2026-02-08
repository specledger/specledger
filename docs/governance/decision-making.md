# Decision Making Process

How proposals become features in SpecLedger.

## Overview

SpecLedger uses a lightweight, transparent decision-making process. This document explains how proposals are reviewed, approved, and implemented.

## Proposal Types

1. **Feature Proposals**: New functionality or major changes
2. **Bug Reports**: Issues requiring fixes
3. **Enhancement Requests**: Improvements to existing features
4. **Process Changes**: Changes to how the project is run

## Proposal Process

### Step 1: Create Proposal

1. Create a GitHub issue
2. Add `proposal` label
3. Use template (if available) or include:
   - Problem statement
   - Proposed solution
   - Alternatives considered
   - Impact assessment

### Step 2: Community Discussion

1. Proposal is open for community discussion
2. Maintainers review and ask clarifying questions
3. Discussion period: Minimum 1 week for major features
4. All feedback considered

### Step 3: Maintainer Review

1. Maintainers review proposal against criteria:
   - Alignment with project goals
   - Technical feasibility
   - Resource requirements
   - Community benefit
   - Breaking changes impact

2. Possible outcomes:
   - **Approved**: Move to implementation planning
   - **Needs Revision**: Feedback provided, proposer to update
   - **Deferred**: Not now, but reconsider later
   - **Rejected**: Doesn't fit project direction

### Step 4: Implementation Planning

For approved proposals:
1. Spec created using `/specledger.specify`
2. Implementation plan created using `/specledger.plan`
3. Tasks generated using `/specledger.tasks`
4. Implementation assigned and tracked

### Step 5: Implementation

1. Work tracked via Beads issue tracker
2. Pull requests reviewed
3. Tests must pass
4. Documentation updated
5. Changelog updated

## Decision Criteria

### Alignment with Project Goals

- Does it support SpecLedger's mission as a bootstrap tool?
- Does it maintain framework agnosticism?
- Does it keep the tool lightweight?

### Technical Feasibility

- Can it be implemented with available resources?
- Does it fit the current architecture?
- Are dependencies manageable?

### Community Benefit

- Does it solve a real user problem?
- Do multiple users benefit?
- Is there demand for this feature?

### Breaking Changes

- Breaking changes require strong justification
- Migration path must be provided
- Deprecation period: 3 months minimum

## Maintainer Role

### Responsibilities

- Review proposals in timely manner
- Facilitate community discussion
- Make final decisions on proposals
- Ensure technical quality of contributions

### Decision Making

- Decisions are consensus-based when possible
- Maintainers have final say when consensus not reached
- All decisions must be justified with project goals

## Conflict Resolution

### Disagreements

1. Focus on technical arguments, not personalities
2. Consider user perspectives and use cases
3. Seek compromise when possible
4. Maintainers make final call if needed

### Escalation

If you disagree with a decision:
1. Provide additional context or use cases
2. Request reconsideration with new information
3. Accept final decision or fork the project

## Governance Changes

Changes to this process require:
1. Proposal with rationale
2. 2/3 maintainer approval
3. Community discussion period
4. Update to this document

## Transparency

All decisions are documented in:
- GitHub issues (proposals)
- GitHub discussions (community input)
- Changelog (implemented features)
- Governance docs (process changes)

## Contact

For questions about the decision-making process:
- Open a GitHub Discussion
- Contact maintainers via issue
- Review [GOVERNANCE.md](../GOVERNANCE.md) for project governance
