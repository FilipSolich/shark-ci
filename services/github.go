package services

import (
	"context"
	"fmt"
	"os"

	"github.com/FilipSolich/ci-server/models"
	"github.com/google/go-github/v47/github"
	"golang.org/x/oauth2"
	oauth2_github "golang.org/x/oauth2/github"
	"gorm.io/gorm"
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

func CreateWebhook(ctx context.Context, user *models.User, repoName string, repoID int64) (*models.Webhook, error) {
	client := GetGitHubClientByUser(ctx, user)
	hook := &github.Hook{
		Config: map[string]any{
			"url":          "https://" + os.Getenv("HOSTNAME") + "/webhooks",
			"content_type": "json",
			"secret":       os.Getenv("WEBHOOK_SECRET"),
		},
		Events: []string{"push", "pull_request"},
		Active: github.Bool(true),
	}
	// TODO: Log error.
	hook, _, err := client.Repositories.CreateHook(ctx, user.Username, repoName, hook)
	if err != nil {
		fmt.Println("cannot create webhook", err.Error())
		return nil, err
	}

	modelHook := &models.Webhook{
		Model:    gorm.Model{ID: uint(hook.GetID())},
		Service:  "github",
		RepoID:   repoID,
		RepoName: repoName,
	}
	return modelHook, nil
}

func DeleteWebhook(ctx context.Context, user *models.User, hook *models.Webhook) error {
	client := GetGitHubClientByUser(ctx, user)
	fmt.Println("HERE", user.Username, hook.RepoName, hook.ID)
	_, err := client.Repositories.DeleteHook(ctx, user.Username, hook.RepoName, int64(hook.ID))
	return err
}

func ActivateWebhook() {
}

func DeactivateWebhook() {
}
