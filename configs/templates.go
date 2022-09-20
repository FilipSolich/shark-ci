package configs

import (
	"html/template"
	"net/http"
)

var Templates *template.Template

func LoadTemplates(pattern string) {
	Templates = template.Must(template.ParseGlob("templates/*.html"))

}

func RenderTemplate(w http.ResponseWriter, tmpl string, data any) {
	err := Templates.ExecuteTemplate(w, tmpl, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
