package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"golang.org/x/exp/slog"
	"golang.org/x/oauth2"

	"github.com/shark-ci/shark-ci/ci-server/models"
	"github.com/shark-ci/shark-ci/ci-server/service"
	"github.com/shark-ci/shark-ci/ci-server/store"
	"github.com/shark-ci/shark-ci/ci-server/template"
)

type LoginHandler struct {
	l        *slog.Logger
	s        store.Storer
	services service.Services
}

func NewLoginHandler(l *slog.Logger, s store.Storer, services service.Services) *LoginHandler {
	return &LoginHandler{
		l:        l,
		s:        s,
		services: services,
	}
}

func (h *LoginHandler) HandleLoginPage(w http.ResponseWriter, r *http.Request) {
	state, err := uuid.NewRandom()
	if err != nil {
		h.l.Error("cannot generate UUID", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	oauth2State := &models.OAuth2State{
		State:  state,
		Expire: time.Now().Add(30 * time.Minute),
	}
	err = h.s.CreateOAuth2State(r.Context(), oauth2State)
	if err != nil {
		h.l.Error("store: cannot create OAuth2 state", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data := map[string]string{}
	for _, s := range h.services {
		config := s.OAuth2Config()
		url := config.AuthCodeURL(oauth2State.State.String(), oauth2.AccessTypeOffline)
		data[s.Name()+"URL"] = url
	}

	template.RenderTemplate(w, "login.html", map[string]any{
		"URLs": data,
	})
}
