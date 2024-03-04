package handlers

import (
	"log/slog"
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
		slog.Error("Cannot get user repos.", "userID", user.ID, "err", err)
		Error5xx(w, r, http.StatusInternalServerError)
		return
	}

	err = templates.IndexTmpl.Execute(w, map[string]any{
		"Username": user.Username,
		"Repos":    repos,
	})
	if err != nil {
		slog.Error("Cannot execute template.", "template", "index", "err", err)
		Error5xx(w, r, http.StatusInternalServerError)
		return
	}
}
