package configs

import (
	"html/template"
	"net/http"
	"path"
)

var (
	Templates = make(map[string]*template.Template)

	prefix         = "ci-server/templates"
	base           = path.Join(prefix, "_base.html")
	layout         = path.Join(prefix, "_layout.html")
	templatesFiles = map[string][]string{
		"login.html": {base, path.Join(prefix, "login.html")},
		"index.html": {base, layout, path.Join(prefix, "index.html")},
		"repos.html": {base, layout, path.Join(prefix, "repos.html")},
	}
)

func LoadTemplates() {
	for tmpl, files := range templatesFiles {
		Templates[tmpl] = template.Must(template.ParseFiles(files...))
	}
}

func RenderTemplate(w http.ResponseWriter, tmpl string, data any) {
	err := Templates[tmpl].ExecuteTemplate(w, "_base", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
