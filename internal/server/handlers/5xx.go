package handlers

import (
	"net/http"

	"github.com/shark-ci/shark-ci/templates"
)

func Error5xx(w http.ResponseWriter, r *http.Request, code int) {
	w.WriteHeader(code)
	templates.Error5xxTmpl.Execute(w, nil)
}
