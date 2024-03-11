package handler

import (
	"net/http"

	"github.com/shark-ci/shark-ci/internal/server/middleware"
	"github.com/shark-ci/shark-ci/internal/server/store"
	"github.com/shark-ci/shark-ci/templates"
)

type IndexHandler struct {
	s store.Storer
}

func NewIndexHandler(s store.Storer) *IndexHandler {
	return &IndexHandler{
		s: s,
	}
}

func (h *IndexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := middleware.UserFromContext(ctx, w)
	if !ok {
		return
	}

	repos, err := h.s.GetUserRepos(ctx, user.ID)
	if err != nil {
		Error5xx(w, http.StatusInternalServerError, "Cannot get user repos.", err)
		return
	}

	err = templates.IndexTmpl.Execute(w, map[string]any{
		"Username": user.Username,
		"Repos":    repos,
	})
	if err != nil {
		Error5xx(w, http.StatusInternalServerError, "Cannot execute template", err)
		return
	}
}
