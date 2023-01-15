package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"golang.org/x/oauth2"

	"github.com/shark-ci/shark-ci/ci-server/configs"
	"github.com/shark-ci/shark-ci/ci-server/services"
	"github.com/shark-ci/shark-ci/ci-server/store"
	"github.com/shark-ci/shark-ci/models"
)

type LoginHandler struct {
	store store.Storer
}

func NewLoginHandler(s store.Storer) *LoginHandler {
	return &LoginHandler{
		store: s,
	}
}

func (h *LoginHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	state, err := uuid.NewRandom()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	oauth2State := models.OAuth2State{State: state.String()}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]string{}
	for _, service := range services.Services {
		config := service.GetOAuth2Config()
		url := config.AuthCodeURL(oauth2State.State, oauth2.AccessTypeOffline)
		data[service.GetServiceName()+"URL"] = url
	}

	configs.RenderTemplate(w, "login.html", map[string]any{
		"URLs": data,
	})
}
