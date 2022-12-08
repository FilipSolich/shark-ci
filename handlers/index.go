package handlers

import (
	"net/http"

	"github.com/shark-ci/shark-ci/configs"
	"github.com/shark-ci/shark-ci/middlewares"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := middlewares.UserFromContext(r.Context(), w)
	if !ok {
		return
	}

	configs.RenderTemplate(w, "index.html", map[string]any{
		"ID": user.ID,
	})
}
