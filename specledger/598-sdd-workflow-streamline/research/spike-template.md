# Spike Template

Standard template for time-boxed exploratory research. Used by `/specledger.spike` when creating research documents in `specledger/<spec>/research/yyyy-mm-dd-<name>.md`.

---

## Template Structure

```markdown
# Spike: [Question to Answer]

## Objective
[Single sentence: What are we trying to learn?]

## Investigation Plan
- [ ] Research approach: [What will you read/test?]
- [ ] Prototype scope: [What will you build to test this?]
- [ ] Success criteria: [How will you know you have an answer?]
- [ ] Time box: [1-3 days]

## Research & Findings
[Document what you learn - links, patterns, gotchas]

## Prototype Results
[What did you actually build? Did it work?]

## Decision / Recommendation
[What should we do based on this spike?]
- [ ] Proceed with implementation
- [ ] Need more research
- [ ] Different approach recommended
- [ ] Not viable / reject direction
```

---

## Quick Spike Workflow

1. **Define the question** (30 min): What specific uncertainty are you reducing?
2. **Research existing patterns** (2-4 hours): Docs, code examples, prior art
3. **Prototype/test** (4-8 hours): Build minimal proof-of-concept
4. **Document findings** (1-2 hours): What did you learn? What's next?
5. **Make decision** (30 min): Proceed, iterate, or reject
