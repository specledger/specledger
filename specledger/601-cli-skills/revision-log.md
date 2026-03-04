# Revision Log: 601-cli-skills

**Created**: 2026-03-05
**Purpose**: Track document revision decisions for comment feedback

---

## Revision 1: D21 Context Constraints in User Stories

**Date**: 2026-03-05
**Comment**: "was hoping more user stories mention about context constraints and progressive disclosure as in 'Token Efficient CLI output'"

### Options Presented

| Option | Description |
|--------|-------------|
| **A. Add to existing acceptance scenarios** | Enhance current scenarios (US1-US3) with token budget and progressive disclosure language |
| B. Add new dedicated scenarios | Add 1-2 new acceptance scenarios per user story that explicitly test D21 compliance |
| C. Add context note to each story header | Add a brief 'Context: Token Budget' note at the start of US1-US3 |

### Choice Made

**Option A**: Add to existing acceptance scenarios

### Rationale

- Keeps spec concise without adding new scenario numbers
- Token budgets naturally fit into existing compact mode scenarios
- Progressive disclosure pattern (list=compact, show=full) directly maps to existing scenarios

### Changes Applied

#### spec.md

**US1 - sl comment list Command**:
- Scenario 1: Added `content_preview (truncated to 120 chars)` and `reply_count` fields
- Scenario 6: Enhanced with `~500 tokens for 25 comments` and `progressive disclosure for agent context efficiency`
- Scenario 7 (new): Added footer hint for drill-down guidance

**US2 - sl comment show Command**:
- Scenario 1: Added `no truncation` and `drill-down command for progressive disclosure after list scan`
- Scenario 5 (new): Added `~200 tokens for 1 comment + 3 replies`

**US3 - sl comment reply/resolve Commands**:
- Scenario 1: Added `minimal output (~30 tokens) for agent context efficiency`
- Scenario 3: Added `minimal confirmation output`

#### tasks.md (DoD items)

- SL-a801e5: Changed `80-char truncation` → `120-char truncation`, added `Footer hint for drill-down`, `~500 tokens for 25 comments`
- SL-c2922e: Added `no truncation`, `~200 tokens for 1 comment + 3 replies`
- SL-34e30: Added `Minimal output (~30 tokens)`
- SL-638e23: Added `Minimal confirmation output`

#### plan.md

No changes needed — already references D21 in:
- Line 8: Architecture decisions
- Line 35: Constitution check
- Lines 160-169: Dedicated "Token-efficient output (D21)" section

### Comments Addressed

- [x] Comment 1 (spec.md, "sl comment" target)
