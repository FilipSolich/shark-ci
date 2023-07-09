package handlers

import (
	"net/http"

	"github.com/FilipSolich/shark-ci/ci-server/middlewares"
	"github.com/FilipSolich/shark-ci/ci-server/template"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := middlewares.UserFromContext(r.Context(), w)
	if !ok {
		return
	}

	template.RenderTemplate(w, "index.html", map[string]any{
		"ID": user.ID,
	})
}
