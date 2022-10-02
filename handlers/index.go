package handlers

import (
	"net/http"

	"github.com/FilipSolich/ci-server/configs"
	"github.com/FilipSolich/ci-server/middlewares"
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
