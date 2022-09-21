package configs

import (
	"html/template"
	"net/http"
	"path"
)

var Templates = make(map[string]*template.Template)

func LoadTemplates(pattern string) {
	prefix := "templates"
	base := path.Join(prefix, "_base.html")
	files := []string{
		"index.html",
		"login.html",
		"registered_repos.html",
	}

	for _, file := range files {
		Templates[file] = template.Must(template.ParseFiles(path.Join(prefix, file), base))
	}
}

func RenderTemplate(w http.ResponseWriter, tmpl string, data any) {
	err := Templates[tmpl].ExecuteTemplate(w, "_base", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
