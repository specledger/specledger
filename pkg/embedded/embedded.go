package embedded

import (
	"embed"
)

//go:embed skills
var SkillsFS embed.FS

// TemplatesFS provides template file system access
// TODO: Populate with actual templates
var TemplatesFS embed.FS
