package services

import (
	"context"

	"github.com/FilipSolich/ci-server/configs"
	"github.com/FilipSolich/ci-server/models"
	"github.com/google/go-github/v47/github"
	"golang.org/x/oauth2"
	oauth2_github "golang.org/x/oauth2/github"
)

var GitHub GitHubService

type GitHubService struct {
	OAuth2Config *oauth2.Config
}

func NewGitHub(clientID string, clientSecret string) {
	GitHub.OAuth2Config = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{"repo"},
		Endpoint:     oauth2_github.Endpoint,
	}
}

func (s *GitHubService) GetServiceName() string {
	return "github"
}

func GetGitHubClientByUser(ctx context.Context, user *models.User) *github.Client {
	token := user.Token.Token
	client := GitHub.OAuth2Config.Client(ctx, &token)
	ghClient := github.NewClient(client)
	return ghClient
}

func CreateWebhook(ctx context.Context, user *models.User, repo RepoInfo) (*models.Webhook, error) {
	client := GetGitHubClientByUser(ctx, user)
	hook := defaultWebhook()
	hook, _, err := client.Repositories.CreateHook(ctx, user.Username, repo.Name, hook)
	if err != nil {
		return nil, err
	}

	modelHook := &models.Webhook{
		ServiceWebhookID: hook.GetID(),
		Service:          "github",
		RepoID:           repo.ID,
		RepoName:         repo.Name,
		RepoFullName:     repo.FullName,
		Active:           true,
	}
	return modelHook, nil
}

func DeleteWebhook(ctx context.Context, user *models.User, hook *models.Webhook) error {
	client := GetGitHubClientByUser(ctx, user)
	_, err := client.Repositories.DeleteHook(ctx, user.Username, hook.RepoName, int64(hook.ServiceWebhookID))
	return err
}

func ActivateWebhook(ctx context.Context, user *models.User, hook *models.Webhook) (*models.Webhook, error) {
	return changeWebhookState(ctx, user, hook, true)
}

func DeactivateWebhook(ctx context.Context, user *models.User, hook *models.Webhook) (*models.Webhook, error) {
	return changeWebhookState(ctx, user, hook, false)
}

func UpdateStatus() {

}

func defaultWebhookConfig() map[string]any {
	return map[string]any{
		"url":          "https://" + configs.Host + configs.EventHandlerPath,
		"content_type": "json",
		"secret":       configs.WebhookSecret,
	}
}

func defaultWebhook() *github.Hook {
	return &github.Hook{
		Config: defaultWebhookConfig(),
		Events: []string{"push", "pull_request"},
		Active: github.Bool(true),
	}
}

func changeWebhookState(ctx context.Context, user *models.User, hook *models.Webhook, active bool) (*models.Webhook, error) {
	client := GetGitHubClientByUser(ctx, user)
	ghHook := defaultWebhook()
	ghHook.Active = github.Bool(active)
	_, _, err := client.Repositories.EditHook(ctx, user.Username, hook.RepoName, int64(hook.ServiceWebhookID), ghHook)
	if err == nil {
		hook.Active = active
	}
	return hook, err
}
