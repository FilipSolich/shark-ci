package services

import (
	"context"

	"github.com/FilipSolich/ci-server/db"
	"github.com/FilipSolich/ci-server/models"
	"github.com/google/go-github/v47/github"
	"golang.org/x/oauth2"
	oauth2_github "golang.org/x/oauth2/github"
	"gorm.io/gorm/clause"
)

const GitHubName = "github"

var GitHub GitHubManager

type GitHubManager struct {
	oauth2Config *oauth2.Config
}

func NewGitHubManager(clientID string, clientSecret string) {
	GitHub.oauth2Config = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{"repo"},
		Endpoint:     oauth2_github.Endpoint,
	}
}

func (*GitHubManager) GetServiceName() string {
	return GitHubName
}

func (s *GitHubManager) GetOAuth2Config() *oauth2.Config {
	return s.oauth2Config
}

func (s *GitHubManager) GetOrCreateUserIdentity(ctx context.Context, token *oauth2.Token) (*models.UserIdentity, error) {
	oauth2Client := s.oauth2Config.Client(ctx, token)
	ghClient := github.NewClient(oauth2Client)

	ghUser, _, err := ghClient.Users.Get(ctx, "")
	if err != nil {
		return nil, err
	}

	ui := models.UserIdentity{
		ServiceName: GitHubName,
		Username:    ghUser.GetLogin(),
	}

	return models.GetOrCreateUserIdentity(&ui)
}

// Get repositories which aren't archived and are owned by `user`.
func (s *GitHubManager) GetUsersRepos(user *models.User) ([]*models.Repository, error) {
	ctx := context.Background()
	client := getClientByUser(ctx, user)
	ghRepos, _, err := client.Repositories.List(ctx, "", &github.RepositoryListOptions{
		Type: "owner",
	})

	var repos []*models.Repository
	for _, repo := range ghRepos {
		if !repo.GetArchived() {
			r := &models.Repository{
				ServiceName:   GitHubName,
				ServiceRepoID: repo.GetID(),
				Name:          repo.GetName(),
				FullName:      repo.GetFullName(),
			}
			repos = append(repos, r)
		}
	}

	return repos, err
}

// Legacy code
//func GetGitHubClientByUser(ctx context.Context, user *models.User) *github.Client {
//	token := user.Token.Token
//	client := GitHub.OAuth2Config.Client(ctx, &token)
//	ghClient := github.NewClient(client)
//	return ghClient
//}

//func CreateWebhook(ctx context.Context, user *models.User, repo RepoInfo) (*models.Webhook, error) {
//	client := GetGitHubClientByUser(ctx, user)
//	hook := defaultWebhook()
//	hook, _, err := client.Repositories.CreateHook(ctx, user.Username, repo.Name, hook)
//	if err != nil {
//		return nil, err
//	}
//
//	modelHook := &models.Webhook{
//		ServiceWebhookID: hook.GetID(),
//		Service:          "github",
//		RepoID:           repo.ID,
//		RepoName:         repo.Name,
//		RepoFullName:     repo.FullName,
//		Active:           true,
//	}
//	return modelHook, nil
//}
//
//func DeleteWebhook(ctx context.Context, user *models.User, hook *models.Webhook) error {
//	client := GetGitHubClientByUser(ctx, user)
//	_, err := client.Repositories.DeleteHook(ctx, user.Username, hook.RepoName, int64(hook.ServiceWebhookID))
//	return err
//}
//
//func ActivateWebhook(ctx context.Context, user *models.User, hook *models.Webhook) (*models.Webhook, error) {
//	return changeWebhookState(ctx, user, hook, true)
//}
//
//func DeactivateWebhook(ctx context.Context, user *models.User, hook *models.Webhook) (*models.Webhook, error) {
//	return changeWebhookState(ctx, user, hook, false)
//}
//
//func UpdateStatus(ctx context.Context, user *models.User, repo string, ref string) {
//	//client := GetGitHubClientByUser(ctx, user)
//	//status := github.RepoStatus{}
//	//client.Repositories.CreateStatus(ctx, user.Username, repo, ref)
//}
//
//func defaultWebhookConfig() map[string]any {
//	return map[string]any{
//		"url":          "https://" + configs.Host + configs.EventHandlerPath,
//		"content_type": "json",
//		"secret":       configs.WebhookSecret,
//	}
//}
//
//func defaultWebhook() *github.Hook {
//	return &github.Hook{
//		Config: defaultWebhookConfig(),
//		Events: []string{"push", "pull_request"},
//		Active: github.Bool(true),
//	}
//}

func getClientByUser(ctx context.Context, user *models.User) *github.Client {
	var identity *models.UserIdentity
	err := db.DB.Preload(clause.Associations).First(identity, &models.UserIdentity{UserID: user.ID}).Error
	if err != nil {
		return nil
	}

	token := identity.Token.Token
	client := GitHub.oauth2Config.Client(ctx, &token)
	return github.NewClient(client)
}

//func changeWebhookState(ctx context.Context, user *models.User, hook *models.Webhook, active bool) (*models.Webhook, error) {
//	client := GetGitHubClientByUser(ctx, user)
//	ghHook := defaultWebhook()
//	ghHook.Active = github.Bool(active)
//	_, _, err := client.Repositories.EditHook(ctx, user.Username, hook.RepoName, int64(hook.ServiceWebhookID), ghHook)
//	if err == nil {
//		hook.Active = active
//	}
//	return hook, err
//}
