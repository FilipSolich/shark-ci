package templates

import (
	"embed"
	"html/template"
)

//go:embed *.html base/*.html partials/*.html
var templates embed.FS

var (
	IndexTmpl = template.Must(template.New("base.html").Funcs(FuncMap).ParseFS(templates, "base/base.html", "base/layout.html", "index.html", "partials/repo.html"))
	LoginTmpl = template.Must(template.New("base.html").Funcs(FuncMap).ParseFS(templates, "base/base.html", "login.html"))

	Error5xxTmpl = template.Must(template.New("base.html").Funcs(FuncMap).ParseFS(templates, "base/base.html", "errors/5xx.html"))
)

var FuncMap = template.FuncMap{
	"Modulo": Modulo,
}

func Modulo(a int, b int) bool {
	return a%b == 0
}
