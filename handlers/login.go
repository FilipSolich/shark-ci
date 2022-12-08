package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"golang.org/x/oauth2"

	"github.com/shark-ci/shark-ci/configs"
	"github.com/shark-ci/shark-ci/db"
	"github.com/shark-ci/shark-ci/services"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	state, err := uuid.NewRandom()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	oauth2State := db.OAuth2State{State: state.String()}
	_, err = db.NewOAuth2State(&oauth2State)
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
