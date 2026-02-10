package embedded

import (
	"embed"
)

//go:embed all:skills
var SkillsFS embed.FS

//go:embed all:templates
var TemplatesFS embed.FS
