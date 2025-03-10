package templates

import "embed"

//go:embed *.gohtml config_template.yml
var Files embed.FS
