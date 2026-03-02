package context

type PlanMetadata struct {
	Language     string `json:"language"`
	Dependencies string `json:"dependencies"`
	Storage      string `json:"storage"`
	ProjectType  string `json:"project_type"`
}

type AgentFileMapping struct {
	AgentType string
	FilePath  string
}

type UpdateResult struct {
	AgentType string `json:"agent_type"`
	FilePath  string `json:"file_path"`
	Action    string `json:"action"`
}

var AgentFileMappings = []AgentFileMapping{
	{AgentType: "claude", FilePath: "CLAUDE.md"},
	{AgentType: "gemini", FilePath: "GEMINI.md"},
	{AgentType: "copilot", FilePath: ".github/agents/copilot-instructions.md"},
	{AgentType: "cursor-agent", FilePath: ".cursor/rules/specify-rules.mdc"},
	{AgentType: "opencode", FilePath: "AGENTS.md"},
	{AgentType: "codex", FilePath: "AGENTS.md"},
	{AgentType: "amp", FilePath: "AGENTS.md"},
	{AgentType: "q", FilePath: "AGENTS.md"},
	{AgentType: "bob", FilePath: "AGENTS.md"},
	{AgentType: "windsurf", FilePath: ".windsurf/rules/specify-rules.md"},
	{AgentType: "qwen", FilePath: "QWEN.md"},
	{AgentType: "kilocode", FilePath: "KILOCODE.md"},
	{AgentType: "auggie", FilePath: "AUGGIE.md"},
	{AgentType: "roo", FilePath: "ROO.md"},
	{AgentType: "codebuddy", FilePath: "CODEBUDDY.md"},
	{AgentType: "qoder", FilePath: "QODER.md"},
	{AgentType: "shai", FilePath: "SHAI.md"},
}

const (
	ManualAdditionsStart = "<!-- MANUAL ADDITIONS START -->"
	ManualAdditionsEnd   = "<!-- MANUAL ADDITIONS END -->"
	MaxTechStackEntries  = 5
	MaxRecentChanges     = 3
)
