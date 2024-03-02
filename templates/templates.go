package templates

import (
	"embed"
	"html/template"
	"log/slog"
	"os"
)

//go:embed *.html
var templates embed.FS
var IndexTmpl *template.Template
var LoginTmpl *template.Template
var ReposTmpl *template.Template

func ParseTemplates() {
	parseTemplate(&IndexTmpl, "_base.html", "_layout.html", "index.html")
	parseTemplate(&LoginTmpl, "_base.html", "login.html")
	parseTemplate(&ReposTmpl, "_base.html", "_layout.html", "repos.html")
}

func parseTemplate(tmpl **template.Template, filename ...string) {
	var err error
	*tmpl, err = template.ParseFS(templates, filename...)
	if err != nil {
		slog.Error("Parsing template failed", "template", filename[len(filename)-1], "err", err)
		os.Exit(1)
	}
}
