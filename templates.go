package templates

import "embed"

//go:embed templates/*.html
var Files embed.FS
