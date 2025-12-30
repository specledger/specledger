User Stories (MVP) - specledger CLI "sl":
1. bootstrap (per agent shell focused /commands, CLI, ... to bootstrap - initial very opinionated for Claude Code, Zsh, mise and a CLI)
  - specify, clarify, plan, tasks, analyze, implement, resume commands?
  - claude skills
  - claude code session start and session end similar to beads?
2. controlplane:
  - basic template management for all claude commands, skills bootstrapped into the repo
  - Issue tracker sync (bd sync -> sl API)

sl CLI commands:
1. `sl init` - bootstrap a new specledger project
2. (use bd for now) `bd create -type epic|feature|task|bug -titel "title" -description "description" -design ".."`

User options:
- Choose the Spec Driven Development Playbook: Default, Data Science, SRE, Product
- Choose preferred agent shell: Claude Code, Gemini CLI, Codex, co-pilot

Expansion:

1. Replace `bd` by `sl`
1. Bootstrap: custom per tenant / user (user tunes his own templates) <- how do we propagate updates to user managed templates? do we use patching mechanisms?

core idea:
- Adopt SDD
  - Need to provide use cases? Help with prompts?
  - Need to provide task management
- Needs to be collaborative <- what does this mean?
- Able to integrate with preferred workflows of users (support different agent shells)


User -> SaaS
- Set me up this playbook
- CLI uses playbook to iterate on the tasks


