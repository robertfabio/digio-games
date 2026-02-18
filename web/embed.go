package web

import (
	"embed"
	"html/template"
	"io/fs"
)

//go:embed templates/*.html
var templateFS embed.FS

//go:embed static
var staticFS embed.FS

var cachedTemplates *template.Template

func init() {
	cachedTemplates = template.Must(template.ParseFS(templateFS, "templates/*.html"))
}

func Templates() *template.Template {
	return cachedTemplates
}

func StaticFS() fs.FS {
	sub, _ := fs.Sub(staticFS, "static")
	return sub
}
