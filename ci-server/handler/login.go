package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/oauth2"

	"github.com/FilipSolich/shark-ci/ci-server/service"
	"github.com/FilipSolich/shark-ci/ci-server/store"
	"github.com/FilipSolich/shark-ci/ci-server/template"
	"github.com/FilipSolich/shark-ci/shared/model"
)

type LoginHandler struct {
	l        *zap.SugaredLogger
	s        store.Storer
	services service.Services
}

func NewLoginHandler(l *zap.SugaredLogger, s store.Storer, services service.Services) *LoginHandler {
	return &LoginHandler{
		l:        l,
		s:        s,
		services: services,
	}
}

func (h *LoginHandler) HandleLoginPage(w http.ResponseWriter, r *http.Request) {
	state, err := uuid.NewRandom()
	if err != nil {
		h.l.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	oauth2State := model.NewOAuth2State(state.String(), 30*time.Minute)
	if err != nil {
		h.l.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = h.s.CreateOAuth2State(r.Context(), oauth2State)
	if err != nil {
		h.l.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data := map[string]string{}
	for _, s := range h.services {
		config := s.OAuth2Config()
		url := config.AuthCodeURL(oauth2State.State, oauth2.AccessTypeOffline)
		data[s.Name()+"URL"] = url
	}

	template.RenderTemplate(w, "login.html", map[string]any{
		"URLs": data,
	})
}
