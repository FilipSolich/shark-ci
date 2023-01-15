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
	store      store.Storer
	serviceMap services.ServiceMap
}

func NewLoginHandler(s store.Storer, serviceMap services.ServiceMap) *LoginHandler {
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

	oauth2State := models.OAuth2State{State: state.String()}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]string{}
	for _, service := range h.serviceMap {
		config := service.OAuth2Config()
		url := config.AuthCodeURL(oauth2State.State, oauth2.AccessTypeOffline)
		data[service.ServiceName()+"URL"] = url
	}

	configs.RenderTemplate(w, "login.html", map[string]any{
		"URLs": data,
	})
}
