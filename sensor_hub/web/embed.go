package web

import "embed"

//go:embed all:dist
var distFS embed.FS

//go:embed all:docs
var docsFS embed.FS
