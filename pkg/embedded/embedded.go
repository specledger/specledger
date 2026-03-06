package embedded

import (
	"embed"
)

//go:embed all:templates
var TemplatesFS embed.FS
