package templates

import (
	"embed"
	"html/template"
)

//go:embed *.html base/*.html partials/*.html errors/*.html
var templates embed.FS

var (
	IndexTmpl = template.Must(template.New("base.html").Funcs(FuncMap).ParseFS(templates, "base/base.html", "base/layout.html", "index.html"))
	LoginTmpl = template.Must(template.New("base.html").Funcs(FuncMap).ParseFS(templates, "base/base.html", "login.html"))

	ReposRegisterTmpl = template.Must(template.ParseFS(templates, "partials/repos_register.html"))

	Error400Tmpl = template.Must(template.New("base.html").ParseFS(templates, "base/base.html", "errors/400.html"))
	Error404Tmpl = template.Must(template.New("base.html").ParseFS(templates, "base/base.html", "errors/404.html"))
	Error5xxTmpl = template.Must(template.New("base.html").ParseFS(templates, "base/base.html", "errors/5xx.html"))
)

var FuncMap = template.FuncMap{
	"Modulo": Modulo,
}

func Modulo(a int, b int) bool {
	return a%b == 0
}
