package handler

import (
	"net/http"

	"github.com/shark-ci/shark-ci/internal/server/middleware"
	"github.com/shark-ci/shark-ci/internal/server/templates"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context(), w)
	if !ok {
		return
	}

	templates.IndexTmpl.Execute(w, map[string]any{
		"ID": user.ID,
	})
	//templates.RenderTemplate(w, "index.html", map[string]any{
	//	"ID": user.ID,
	//})
}
