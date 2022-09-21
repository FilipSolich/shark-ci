package handlers

import (
	"net/http"

	"github.com/FilipSolich/ci-server/configs"
	"github.com/FilipSolich/ci-server/services"
	"golang.org/x/oauth2"
)

type loginData struct {
	GitHubLoginURL string
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	GitHubLoginURL := services.GitHubOAut2Config.AuthCodeURL("state", oauth2.AccessTypeOffline)

	configs.RenderTemplate(w, "login.html", loginData{GitHubLoginURL: GitHubLoginURL})
}
