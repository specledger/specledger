// Package skills provides the skills.sh registry client, source parsing,
// lock file management, and skill installation logic for the sl skill command.
//
// It integrates with Vercel's skills.sh registry via public HTTP APIs:
//   - skills.sh/api/search — skill discovery by keyword
//   - add-skill.vercel.sh/audit — security risk assessments (ATH, Socket, Snyk)
//   - add-skill.vercel.sh/t — install/remove telemetry
//   - api.github.com — GitHub Trees API for repo skill enumeration
//   - raw.githubusercontent.com — SKILL.md content fetch
//
// Skills are installed to agent-specific directories resolved from the agent
// registry and tracked in a Vercel-compatible skills-lock.json.
package skills
