package static

import "embed"

//go:embed htmx.min.js bootstrap.min.css
var StaticFS embed.FS
