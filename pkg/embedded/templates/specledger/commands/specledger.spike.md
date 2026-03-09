---
description: Time-boxed exploratory research for investigating technologies, patterns, or solutions. Output saved to specledger/<spec>/research/yyyy-mm-dd-<topic>.md
---

## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

## Purpose

Conduct time-boxed exploratory research to investigate a technology, pattern, or solution. Capture findings in a structured research document for future reference.

**When to use**: Before planning when you need to explore unknowns, evaluate alternatives, or prototype approaches.

## Outline

Goal: Investigate a research topic and produce a structured report with findings, decisions, and recommendations.

Execution steps:

1. Run `sl spec info --json --paths-only` to get `FEATURE_DIR` and `FEATURE_SPEC`.

2. Parse the research topic from `$ARGUMENTS`. If empty, ask the user what they want to research.

3. Create a research file at `specledger/<spec>/research/yyyy-mm-dd-<topic-slug>.md`:
   - Use today's date
   - Create a URL-safe slug from the topic (lowercase, hyphens for spaces)
   - If file exists, append `-2`, `-3`, etc.

4. Conduct the research:
   - Search documentation, codebases, and web resources
   - Prototype or test if needed
   - Capture code snippets and examples
   - Note trade-offs and alternatives considered

5. Write the research file with this structure:

   ```markdown
   # Research: <Topic>

   **Date**: YYYY-MM-DD
   **Context**: <Why this research was needed>
   **Time-box**: <Duration spent>

   ## Question

   <The question or problem being investigated>

   ## Findings

   <Key discoveries, organized by subtopic>

   ### Finding 1: <Title>

   <Details, code examples, measurements>

   ### Finding 2: <Title>

   <Details, code examples, measurements>

   ## Decisions

   <Conclusions drawn from the research>

   - Decision 1: <What was decided and why>
   - Decision 2: <What was decided and why>

   ## Recommendations

   <Actionable next steps based on findings>

   1. <Recommendation 1>
   2. <Recommendation 2>

   ## References

   - <Links to documentation, articles, or code>
   ```

6. Report completion:
   - Path to research file
   - Summary of key findings (2-3 bullets)
   - Recommended next steps

## Behavior Rules

- Time-box research to prevent scope creep (default: 30 minutes)
- Focus on answering the specific question, not general exploration
- Include code examples when relevant
- Note confidence level in findings (high/medium/low)
- If research reveals the question was wrong, document the better question
- Link to relevant sections of the spec if findings impact requirements

## Example Usage

```bash
# Research authentication patterns
/specledger.spike "OAuth2 vs JWT for API authentication"

# Research performance optimization
/specledger.spike "Database query optimization for large datasets"

# Research technology choice
/specledger.spike "React vs Vue for admin dashboard"
```
