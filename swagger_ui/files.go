package swagger_ui

import "embed"

// SwaggerUIStaticFiles is the embedded file system that serves the swagger-ui static files
//
//go:embed *.html *.css *.js *.png
var SwaggerUIStaticFiles embed.FS
