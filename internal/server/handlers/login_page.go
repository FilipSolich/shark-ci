package handlers

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"golang.org/x/oauth2"

	"github.com/shark-ci/shark-ci/internal/server/service"
	"github.com/shark-ci/shark-ci/internal/server/store"
	"github.com/shark-ci/shark-ci/internal/server/types"
	"github.com/shark-ci/shark-ci/templates"
)

type LoginHandler struct {
	s        store.Storer
	services service.Services
}

func NewLoginHandler(s store.Storer, services service.Services) *LoginHandler {
	return &LoginHandler{
		s:        s,
		services: services,
	}
}

func (h *LoginHandler) HandleLoginPage(w http.ResponseWriter, r *http.Request) {
	state, err := uuid.NewRandom()
	if err != nil {
		slog.Error("cannot generate UUID", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	oauth2State := types.OAuth2State{
		State:  state,
		Expire: time.Now().Add(30 * time.Minute),
	}
	err = h.s.CreateOAuth2State(r.Context(), oauth2State)
	if err != nil {
		slog.Error("store: cannot create OAuth2 state", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data := map[string]string{}
	for _, s := range h.services {
		config := s.OAuth2Config()
		url := config.AuthCodeURL(oauth2State.State.String(), oauth2.AccessTypeOffline)
		data[s.Name()+"URL"] = url
	}

	err = templates.LoginTmpl.Execute(w, map[string]any{"URLs": data})
	if err != nil {
		slog.Error("Cannot execute template.", "template", templates.LoginTmpl.Name(), "err", err)
		Error5xx(w, r, http.StatusInternalServerError)
		return
	}
}
