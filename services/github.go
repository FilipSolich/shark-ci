package services

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

var GitHubOAut2Config *oauth2.Config

func NewGitHubOAuth2Config(clientID string, clientSecret string) {
	GitHubOAut2Config = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{"repo"},
		Endpoint:     github.Endpoint,
	}
}
