Spec Driven Development is an LLM driven software development workflow that leverages a Spec  (User Stories, Functional Requirements, Edge Case clarifications, ...) First methodology. The concepts were made popular by the launch of AWS's Kiro IDE which is centered around this workflow. Many similar work followed such as GitHub's spec-kit and Google's Conductor launched last week.
With spec driven development the most important artifacts collected during feature work and software development is no longer just the code but also all the decision record along the way.

The full Spec-Kit flow is defined in concepts document below

<spec_kit_concepts>
# Specification-Driven Development (SDD)

## The Power Inversion

For decades, code has been king. Specifications served code—they were the scaffolding we built and then discarded once the "real work" of coding began. We wrote PRDs to guide development, created design docs to inform implementation, drew diagrams to visualize architecture. But these were always subordinate to the code itself. Code was truth. Everything else was, at best, good intentions. Code was the source of truth, and as it moved forward, specs rarely kept pace. As the asset (code) and the implementation are one, it's not easy to have a parallel implementation without trying to build from the code.

Spec-Driven Development (SDD) inverts this power structure. Specifications don't serve code—code serves specifications. The Product Requirements Document (PRD) isn't a guide for implementation; it's the source that generates implementation. Technical plans aren't documents that inform coding; they're precise definitions that produce code. This isn't an incremental improvement to how we build software. It's a fundamental rethinking of what drives development.

The gap between specification and implementation has plagued software development since its inception. We've tried to bridge it with better documentation, more detailed requirements, stricter processes. These approaches fail because they accept the gap as inevitable. They try to narrow it but never eliminate it. SDD eliminates the gap by making specifications and their concrete implementation plans born from the specification executable. When specifications and implementation plans generate code, there is no gap—only transformation.

This transformation is now possible because AI can understand and implement complex specifications, and create detailed implementation plans. But raw AI generation without structure produces chaos. SDD provides that structure through specifications and subsequent implementation plans that are precise, complete, and unambiguous enough to generate working systems. The specification becomes the primary artifact. Code becomes its expression (as an implementation from the implementation plan) in a particular language and framework.

In this new world, maintaining software means evolving specifications. The intent of the development team is expressed in natural language ("**intent-driven development**"), design assets, core principles and other guidelines. The **lingua franca** of development moves to a higher level, and code is the last-mile approach.

Debugging means fixing specifications and their implementation plans that generate incorrect code. Refactoring means restructuring for clarity. The entire development workflow reorganizes around specifications as the central source of truth, with implementation plans and code as the continuously regenerated output. Updating apps with new features or creating a new parallel implementation because we are creative beings, means revisiting the specification and creating new implementation plans. This process is therefore a 0 -> 1, (1', ..), 2, 3, N.

The development team focuses in on their creativity, experimentation, their critical thinking.

## The SDD Workflow in Practice

The workflow begins with an idea—often vague and incomplete. Through iterative dialogue with AI, this idea becomes a comprehensive PRD. The AI asks clarifying questions, identifies edge cases, and helps define precise acceptance criteria. What might take days of meetings and documentation in traditional development happens in hours of focused specification work. This transforms the traditional SDLC—requirements and design become continuous activities rather than discrete phases. This is supportive of a **team process**, where team-reviewed specifications are expressed and versioned, created in branches, and merged.

When a product manager updates acceptance criteria, implementation plans automatically flag affected technical decisions. When an architect discovers a better pattern, the PRD updates to reflect new possibilities.

Throughout this specification process, research agents gather critical context. They investigate library compatibility, performance benchmarks, and security implications. Organizational constraints are discovered and applied automatically—your company's database standards, authentication requirements, and deployment policies seamlessly integrate into every specification.

From the PRD, AI generates implementation plans that map requirements to technical decisions. Every technology choice has documented rationale. Every architectural decision traces back to specific requirements. Throughout this process, consistency validation continuously improves quality. AI analyzes specifications for ambiguity, contradictions, and gaps—not as a one-time gate, but as an ongoing refinement.

Code generation begins as soon as specifications and their implementation plans are stable enough, but they do not have to be "complete." Early generations might be exploratory—testing whether the specification makes sense in practice. Domain concepts become data models. User stories become API endpoints. Acceptance scenarios become tests. This merges development and testing through specification—test scenarios aren't written after code, they're part of the specification that generates both implementation and tests.

The feedback loop extends beyond initial development. Production metrics and incidents don't just trigger hotfixes—they update specifications for the next regeneration. Performance bottlenecks become new non-functional requirements. Security vulnerabilities become constraints that affect all future generations. This iterative dance between specification, implementation, and operational reality is where true understanding emerges and where the traditional SDLC transforms into a continuous evolution.

## Why SDD Matters Now

Three trends make SDD not just possible but necessary:

First, AI capabilities have reached a threshold where natural language specifications can reliably generate working code. This isn't about replacing developers—it's about amplifying their effectiveness by automating the mechanical translation from specification to implementation. It can amplify exploration and creativity, support "start-over" easily, and support addition, subtraction, and critical thinking.

Second, software complexity continues to grow exponentially. Modern systems integrate dozens of services, frameworks, and dependencies. Keeping all these pieces aligned with original intent through manual processes becomes increasingly difficult. SDD provides systematic alignment through specification-driven generation. Frameworks may evolve to provide AI-first support, not human-first support, or architect around reusable components.

Third, the pace of change accelerates. Requirements change far more rapidly today than ever before. Pivoting is no longer exceptional—it's expected. Modern product development demands rapid iteration based on user feedback, market conditions, and competitive pressures. Traditional development treats these changes as disruptions. Each pivot requires manually propagating changes through documentation, design, and code. The result is either slow, careful updates that limit velocity, or fast, reckless changes that accumulate technical debt.

SDD can support what-if/simulation experiments: "If we need to re-implement or change the application to promote a business need to sell more T-shirts, how would we implement and experiment for that?"

SDD transforms requirement changes from obstacles into normal workflow. When specifications drive implementation, pivots become systematic regenerations rather than manual rewrites. Change a core requirement in the PRD, and affected implementation plans update automatically. Modify a user story, and corresponding API endpoints regenerate. This isn't just about initial development—it's about maintaining engineering velocity through inevitable changes.

## Core Principles

**Specifications as the Lingua Franca**: The specification becomes the primary artifact. Code becomes its expression in a particular language and framework. Maintaining software means evolving specifications.

**Executable Specifications**: Specifications must be precise, complete, and unambiguous enough to generate working systems. This eliminates the gap between intent and implementation.

**Continuous Refinement**: Consistency validation happens continuously, not as a one-time gate. AI analyzes specifications for ambiguity, contradictions, and gaps as an ongoing process.

**Research-Driven Context**: Research agents gather critical context throughout the specification process, investigating technical options, performance implications, and organizational constraints.

**Bidirectional Feedback**: Production reality informs specification evolution. Metrics, incidents, and operational learnings become inputs for specification refinement.

**Branching for Exploration**: Generate multiple implementation approaches from the same specification to explore different optimization targets—performance, maintainability, user experience, cost.

## Implementation Approaches

Today, practicing SDD requires assembling existing tools and maintaining discipline throughout the process. The methodology can be practiced with:

- AI assistants for iterative specification development
- Research agents for gathering technical context
- Code generation tools for translating specifications to implementation
- Version control systems adapted for specification-first workflows
- Consistency checking through AI analysis of specification documents

The key is treating specifications as the source of truth, with code as the generated output that serves the specification rather than the other way around.

## Streamlining SDD with Commands

The SDD methodology is significantly enhanced through three powerful commands that automate the specification → planning → tasking workflow:

### The `/speckit.specify` Command

This command transforms a simple feature description (the user-prompt) into a complete, structured specification with automatic repository management:

1. **Automatic Feature Numbering**: Scans existing specs to determine the next feature number (e.g., 001, 002, 003)
2. **Branch Creation**: Generates a semantic branch name from your description and creates it automatically
3. **Template-Based Generation**: Copies and customizes the feature specification template with your requirements
4. **Directory Structure**: Creates the proper `specs/[branch-name]/` structure for all related documents

### The `/speckit.plan` Command

Once a feature specification exists, this command creates a comprehensive implementation plan:

1. **Specification Analysis**: Reads and understands the feature requirements, user stories, and acceptance criteria
2. **Constitutional Compliance**: Ensures alignment with project constitution and architectural principles
3. **Technical Translation**: Converts business requirements into technical architecture and implementation details
4. **Detailed Documentation**: Generates supporting documents for data models, API contracts, and test scenarios
5. **Quickstart Validation**: Produces a quickstart guide capturing key validation scenarios


### The `/speckit.tasks` Command

After a plan is created, this command analyzes the plan and related design documents to generate an executable task list:

1. **Inputs**: Reads `plan.md` (required) and, if present, `data-model.md`, `contracts/`, and `research.md`
2. **Task Derivation**: Converts contracts, entities, and scenarios into specific tasks
3. **Parallelization**: Marks independent tasks `[P]` and outlines safe parallel groups
4. **Output**: Writes `tasks.md` in the feature directory, ready for execution by a Task agent
</spec_kit_concepts>
source: https://raw.githubusercontent.com/github/spec-kit/refs/heads/main/spec-driven.md

Conductor has very similar phases
https://developers.googleblog.com/conductor-introducing-context-driven-development-for-gemini-cli/

Important ideas: Measure twice, code once

Benjamin Franklin said: "Failing to plan is planning to fail". Yet, in the age of AI, we often dive straight into implementation without establishing a clear understanding of what we are building. Conductor, a new extension now available in preview for Gemini CLI, changes this workflow by using context-driven development. Rather than depending on impermanent chat logs, Conductor helps you create formal specs and plans that live alongside your code in persistent Markdown files. This allows you to plan before you build, review plans before code is written, and keep the human developer firmly in the driver's seat.

The philosophy behind Conductor is simple: control your code.

Instead of diving straight into implementation, Conductor helps you formalize your intent. It unlocks context-driven development by shifting the context of your project out of the chat window and directly into your codebase. By treating context as a managed artifact alongside your code, you transform your repository into a single source of truth that drives every agent interaction with deep, persistent project awareness.

What Conductor adds (and isn't clear in Kiro or GitHub/spec-kit) is how to scale this across a team with this section:
Conductor for teams
Conductor allows you to set project-level context for your product, your tech stack, and your workflow preferences. You can set these preferences once, and they become a shared foundation for every feature your team builds. For example, teams can define an established testing strategy that would automatically be used by Gemini.

By centralizing your technical constraints and coding standards, you ensure that every AI-generated contribution adheres to your specific guidelines, regardless of which developer runs the command. This shared configuration accelerates onboarding for new team members and guarantees that features built by different people feel like they were written by a single, cohesive engineering team.

However this misses out important features of Kiro:
- Each task in Kiro is executed in a single Coding Agent CLI session and completed tasks keep track of the original session history that completed the implementation of a task

This context, should be available to all Humans and LLMs involved in the software development effort.
An interesting solution to this could be Depot's remote session storage solution https://depot.dev/blog/now-available-claude-code-sessions-in-depot

Kiro has amazing features, but is tied to its own IDE and does not provide the cross LLM CLI and Model flexibility of GitHub/spec-kit.

Additionally, an issue  with most of these systems is lack of proper issue tracking, where issues are divided into types of:
1. Epic -> speckit.specify "branch"
2. Feature -> Phases and User stories within a speckit.plan
3. Task -> Tasks within a feature (Phase)

to achieve this, a proper issue tracker which should be accessible to all actors involved (the local LLM agent executing a task in it's shell .. which session is attached back to the task - as well as other team members and humans)

The goal is to build a system on top of the Spec Driven Development with a remote server to fill the gaps and make this LLM Driven workflow scalable across teams.

This includes but is not limited to:
1. Spec generation session history (track the LLM session that created the spec generation in collaboration with the human, .. this should allow frequent check in points across clarification phases and edge case detection plus track how these important questions are address by initial human drive creating the final spec for everyone to review)
2. Plan and research phase (this is a single instance / approach towards implementing the identified user stories and functional requirements, here a technical stack decision is made, initial research is done reviewing code libraries and services that can be used to implement the Spec. The artifacts collected are generally: Plan (high level with research decisions made), Research (tech and alternatives considered), Quickstart (a UX preview of how each user story will be achieved with the selection of techs considered) - additionally this includes data models, data contracts, ... - these artifacts must be tracked with checkpoints and context of involved sessions, allowing different branch out to come to an agreed final version of them comparing pros and cons or reverting a partial implementation to try out a different branch (this is tracked in a remote server)
3. Task generation phase: This is an LLM driven process where all phases are defined (standard phases such as initial foundational set up, folder structure, module init, dep install) and phases for each user story with a priority on those that are defined as required to reach MVP - Again the actual task generation involves human oversight and context needs to be captured in remote system. The full plan is registered with interdependencies (parent-child, blocker, ...) for structured approach towards the next workflow step of "implementation"
4. Actual implementation workflow step where issue tracker API is central to serve "ready work" and is tightly integrated within the LLM Agent Shell (to keep conversation history, file edits, and LLM + human decision flow recorded for easy branching and exploring / reverting alternative options)

and other requirements.

Generate a short description of this LLM Driven workflow system 
