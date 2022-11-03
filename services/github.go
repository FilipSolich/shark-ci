package services

import (
	"context"
	"fmt"
	"log"

	"github.com/google/go-github/v47/github"
	"golang.org/x/oauth2"
	oauth2_github "golang.org/x/oauth2/github"

	"github.com/FilipSolich/ci-server/configs"
	"github.com/FilipSolich/ci-server/db"
)

const GitHubName = "GitHub"                                          // Service name.
const EventHandlerPath = configs.EventHandlerPath + "/" + GitHubName // URL path for events webhook.

var GitHub GitHubManager // Global instance of GitHubManager.

// Manager struct for service config.
type GitHubManager struct {
	oauth2Config *oauth2.Config
}

// Craete new global instance of GitHubManager.
// Needs clientID and clientSecret generated from GitHub.
func NewGitHubManager(clientID string, clientSecret string) {
	GitHub.oauth2Config = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{"repo"},
		Endpoint:     oauth2_github.Endpoint,
	}
}

// Return service name.
func (*GitHubManager) GetServiceName() string {
	return GitHubName
}

func (ghm *GitHubManager) GetOAuth2Config() *oauth2.Config {
	return ghm.oauth2Config
}

func (ghm *GitHubManager) GetOrCreateUserIdentity(ctx context.Context, user *db.User, token *oauth2.Token) (*db.Identity, error) {
	ghClient := getGitHubClient(ctx, token)

	ghUser, _, err := ghClient.Users.Get(ctx, "")
	if err != nil {
		return nil, err
	}

	identity := db.Identity{
		ServiceName: GitHubName,
		Username:    ghUser.GetLogin(),
		Token: db.OAuth2Token{
			AccessToken:  token.AccessToken,
			TokenType:    token.TokenType,
			RefreshToken: token.RefreshToken,
			Expiry:       token.Expiry,
		},
	}

	return db.GetOrCreateIdentity(ctx, &identity, user)
}

// Get repositories which aren't archived and are owned by user `identity`.
func (*GitHubManager) GetUsersRepos(ctx context.Context, identity *db.Identity) ([]*db.Repo, error) {
	client := getClientByIdentity(ctx, identity)

	ghRepos, _, err := client.Repositories.List(ctx, "", &github.RepositoryListOptions{
		Type: "owner",
	})

	var repos []*db.Repo
	for _, repo := range ghRepos {
		if !repo.GetArchived() {
			r := &db.Repo{
				RepoID:      repo.GetID(),
				ServiceName: GitHubName,
				Name:        repo.GetName(),
				FullName:    repo.GetFullName(),
			}
			r, err := db.GetOrCreateRepo(ctx, r)
			if err != nil {
				log.Print(err)
				continue
			}

			repos = append(repos, r)
		}
	}

	return repos, err
}

func (*GitHubManager) CreateWebhook(ctx context.Context, identity *db.Identity, repo *db.Repo) (*db.Webhook, error) {
	client := getClientByIdentity(ctx, identity)

	hook := defaultWebhook()
	hook, _, err := client.Repositories.CreateHook(ctx, identity.Username, repo.Name, hook)
	if err != nil {
		return nil, err
	}

	dbHook := db.Webhook{
		WebhookID: hook.GetID(),
		Active:    true,
	}

	return &dbHook, nil
}

func (*GitHubManager) DeleteWebhook(ctx context.Context, identity *db.Identity, repo *db.Repo, hook *db.Webhook) error {
	client := getClientByIdentity(ctx, identity)

	_, err := client.Repositories.DeleteHook(ctx, identity.Username, repo.Name, hook.WebhookID)
	return err
}

func (*GitHubManager) ChangeWebhookState(ctx context.Context, identity *db.Identity, repo *db.Repo, hook *db.Webhook, active bool) (*db.Webhook, error) {
	client := getClientByIdentity(ctx, identity)

	ghHook := defaultWebhook()
	ghHook.Active = github.Bool(active)
	_, _, err := client.Repositories.EditHook(ctx, identity.Username, repo.Name, hook.WebhookID, ghHook)
	if err != nil {
		return nil, err
	}

	hook.Active = active
	return hook, nil
}

//func (*GitHubManager) CreateJob(ctx context.Context, r *http.Request) (*models.Job, error) {
//	payload, err := github.ValidatePayload(r, []byte(configs.WebhookSecret))
//	if err != nil {
//		return nil, err
//	}
//
//	event, err := github.ParseWebHook(github.WebHookType(r), payload)
//	if err != nil {
//		return nil, err
//	}
//
//	switch event := event.(type) {
//	case *github.PushEvent:
//		commit := event.Commits[len(event.Commits)-1]
//
//		username := event.Repo.Owner.GetLogin()
//		var identity models.UserIdentity
//		err = db.DB.Preload(clause.Associations).Where("username = ?", username).First(&identity).Error
//		if err != nil {
//			return nil, err
//		}
//
//		job := &models.Job{
//			OAuth2TokenID: identity.Token.ID,
//			CommitSHA:     commit.GetID(),
//			CloneURL:      event.Repo.GetCloneURL(),
//		}
//
//		return job, nil
//	}
//	return nil, nil
//}
//
//func (*GitHubManager) UpdateStatus(ctx context.Context, user *models.User, status Status, job *models.Job) error {
//	//identity, client, err := getIdentityClientByUser(ctx, user)
//	//if err != nil {
//	//	return err
//	//}
//
//	//client.Repositories.CreateStatus(ctx, identity.Username, job.CommitSHA)
//
//	return nil
//}

//func UpdateStatus(ctx context.Context, user *models.User, repo string, ref string) {
//	//client := GetGitHubClientByUser(ctx, user)
//	//status := github.RepoStatus{}
//	//client.Repositories.CreateStatus(ctx, user.Username, repo, ref)
//}

func defaultWebhook() *github.Hook {
	return &github.Hook{
		Active: github.Bool(true),
		Events: []string{"push", "pull_request"},
		Config: map[string]any{
			"url":          fmt.Sprintf("https://%s:%s%s", configs.Host, configs.Port, EventHandlerPath),
			"content_type": "json",
			"secret":       configs.WebhookSecret,
		},
	}
}

func getGitHubClient(ctx context.Context, token *oauth2.Token) *github.Client {
	client := GitHub.oauth2Config.Client(ctx, token)
	return github.NewClient(client)
}

func getClientByIdentity(ctx context.Context, identity *db.Identity) *github.Client {
	token := oauth2.Token{
		AccessToken:  identity.Token.AccessToken,
		TokenType:    identity.Token.TokenType,
		RefreshToken: identity.Token.RefreshToken,
		Expiry:       identity.Token.Expiry,
	}
	return getGitHubClient(ctx, &token)
}
