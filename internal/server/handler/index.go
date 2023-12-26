package handler

import (
	"net/http"

	"github.com/shark-ci/shark-ci/internal/server/middleware"
	"github.com/shark-ci/shark-ci/internal/server/template"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context(), w)
	if !ok {
		return
	}

	template.RenderTemplate(w, "index.html", map[string]any{
		"ID": user.ID,
	})
}
