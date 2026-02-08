package embedded

import (
	"embed"
)

//go:embed templates
//go:embed templates/specledger/.claude
//go:embed templates/specledger/.gitattributes
//go:embed templates/specledger/.specledger
var TemplatesFS embed.FS

//go:embed skills
var SkillsFS embed.FS
