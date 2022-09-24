package handlers

import (
	"net/http"

	"github.com/FilipSolich/ci-server/configs"
	"github.com/FilipSolich/ci-server/services"
	"golang.org/x/oauth2"
)

func Login(w http.ResponseWriter, r *http.Request) {
	GitHubLoginURL := services.GitHub.OAuth2Config.AuthCodeURL("state", oauth2.AccessTypeOffline)

	configs.RenderTemplate(w, "login.html", map[string]any{
		"GitHubLoginURL": GitHubLoginURL,
	})
}
