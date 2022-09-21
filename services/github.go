package services

import (
	"context"

	"github.com/FilipSolich/ci-server/models"
	"github.com/google/go-github/v47/github"
	"golang.org/x/oauth2"
	oauth2_github "golang.org/x/oauth2/github"
)

var GitHubOAut2Config *oauth2.Config

func NewGitHubOAuth2Config(clientID string, clientSecret string) {
	GitHubOAut2Config = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{"repo"},
		Endpoint:     oauth2_github.Endpoint,
	}
}

func GetGitHubClientByUser(ctx context.Context, user *models.User) *github.Client {
	token := user.Token.Token
	client := GitHubOAut2Config.Client(ctx, &token)
	ghClient := github.NewClient(client)
	return ghClient
}

func RegisterWebhook() {

}

func DeleteWebhook() {

}
