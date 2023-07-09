package handlers

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"golang.org/x/oauth2"

	"github.com/FilipSolich/shark-ci/ci-server/service"
	"github.com/FilipSolich/shark-ci/ci-server/store"
	"github.com/FilipSolich/shark-ci/ci-server/template"
	"github.com/FilipSolich/shark-ci/model"
)

type LoginHandler struct {
	store      store.Storer
	serviceMap service.ServiceMap
}

func NewLoginHandler(s store.Storer, serviceMap service.ServiceMap) *LoginHandler {
	return &LoginHandler{
		store:      s,
		serviceMap: serviceMap,
	}
}

func (h *LoginHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	state, err := uuid.NewRandom()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	oauth2State := model.NewOAuth2Satate(state.String(), 30*time.Minute)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.store.CreateOAuth2State(r.Context(), oauth2State)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]string{}
	for _, s := range h.serviceMap {
		config := s.OAuth2Config()
		url := config.AuthCodeURL(oauth2State.State, oauth2.AccessTypeOffline)
		data[s.Name()+"URL"] = url
	}

	template.RenderTemplate(w, "login.html", map[string]any{
		"URLs": data,
	})
}
