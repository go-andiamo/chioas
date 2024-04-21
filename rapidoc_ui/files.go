package rapidoc_ui

import "embed"

// RapidocUIStaticFiles is the embedded file system that serves the rapidoc static files
//
//go:embed *.css *.js
var RapidocUIStaticFiles embed.FS
