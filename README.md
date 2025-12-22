# specledger

This system is a **team-scale, LLM-driven Specification-Driven Development (SDD) platform** that treats specifications, plans, and decisions as first-class, persistent artifacts - not ephemeral chat output or secondary documentation.

Built on the foundations popularized by **Kiro IDE**, **spec-kit**, and **Google Conductor**, the system extends SDD beyond single-developer workflows into a **shared, auditable, and scalable collaboration model** for humans and LLM agents.

At its core, the platform introduces a **remote control plane** that captures and links:

* Specifications (PRDs, user stories, acceptance criteria)
* Implementation plans, research, and technical decisions
* Generated task graphs with dependencies and priorities
* Execution history from LLM coding agents and human interventions

Each phase of work - **specification, planning, task generation, and implementation** - is executed through LLM-assisted commands, but every decision, clarification, and alternative explored is **checkpointed and versioned** on a shared server. This enables branching, comparison of approaches, and safe rollback, while preserving the full reasoning trail behind every outcome.

The system integrates a **spec-native issue tracker** that maps cleanly to SDD concepts:

* Epics → Specification branches
* Features → Planned phases and user stories
* Tasks → Executable, agent-driven units of work

LLM coding agents operate directly against this tracker, pulling “ready” work, executing tasks in isolated shells, and attaching their full session context (prompts, file diffs, decisions) back to the task. This ensures that **implementation history becomes shared knowledge**, accessible to both humans and machines.

By decoupling SDD from any single IDE or model and combining it with persistent session storage (inspired by systems like **Depot**), the platform delivers:

* Model-agnostic, CLI-first workflows
* Team-wide consistency and governance
* Traceability from intent → plan → code → production feedback

In short, this system turns SDD from a powerful individual practice into a **repeatable, observable, and collaborative engineering discipline**, enabling teams to truly *measure twice and code once* - even in an AI-accelerated world.
