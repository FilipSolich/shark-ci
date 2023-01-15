package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/go-github/v49/github"
	"golang.org/x/oauth2"
	oauth2_github "golang.org/x/oauth2/github"

	"github.com/shark-ci/shark-ci/ci-server/configs"
	"github.com/shark-ci/shark-ci/ci-server/db"
	"github.com/shark-ci/shark-ci/ci-server/store"
	"github.com/shark-ci/shark-ci/models"
)

const githubName = "GitHub"

type GitHubManager struct {
	name             string
	eventHandlerPath string
	store            store.Storer
	oauth2Config     *oauth2.Config
}

func NewGitHubManager(clientID string, clientSecret string, store store.Storer) *GitHubManager {
	return &GitHubManager{
		name:             githubName,
		eventHandlerPath: configs.EventHandlerPath + "/" + githubName,
		store:            store,
		oauth2Config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Scopes:       []string{"repo"},
			Endpoint:     oauth2_github.Endpoint,
		}}
}

// Return service name.
func (ghm *GitHubManager) ServiceName() string {
	return ghm.name
}

func (*GitHubManager) StatusName(status StatusState) (string, error) {
	switch status {
	case StatusSuccess:
		return "success", nil
	case StatusPending:
		return "pending", nil
	case StatusRunning:
		return "pending", nil
	case StatusError:
		return "error", nil
	}
	return "", errors.New("invalid state")
}

func (ghm *GitHubManager) OAuth2Config() *oauth2.Config {
	return ghm.oauth2Config
}

func (ghm *GitHubManager) GetUserIdentity(ctx context.Context, token *oauth2.Token) (*models.Identity, error) {
	ghClient := getGitHubClient(ctx, token)

	ghUser, _, err := ghClient.Users.Get(ctx, "")
	if err != nil {
		return nil, err
	}

	identity := models.NewIdentity(ghUser.GetLogin(), ghm.name, token)
	return identity, nil
}

// Get repositories which aren't archived and are owned by user `identity`.
func (ghm *GitHubManager) GetUsersRepos(ctx context.Context, identity *models.Identity) ([]*models.Repo, error) {
	client := getClientByIdentity(ctx, identity)

	ghRepos, _, err := client.Repositories.List(ctx, "", &github.RepositoryListOptions{
		Type: "owner",
	})

	var repos []*models.Repo
	for _, repo := range ghRepos {
		if !repo.GetArchived() {
			r := &models.Repo{
				RepoServiceID: repo.GetID(),
				ServiceName:   ghm.ServiceName(),
				Name:          repo.GetName(),
				FullName:      repo.GetFullName(),
			}
			r, err := db.GetOrCreateRepo(ctx, r, identity)
			if err != nil {
				log.Print(err)
				continue
			}

			repos = append(repos, r)
		}
	}

	return repos, err
}

func (*GitHubManager) CreateWebhook(ctx context.Context, identity *models.Identity, repo *models.Repo) (*models.Repo, error) {
	client := getClientByIdentity(ctx, identity)

	hook := defaultWebhook()
	hook, _, err := client.Repositories.CreateHook(ctx, identity.Username, repo.Name, hook)
	if err != nil {
		return nil, err
	}

	repo.WebhookID = hook.GetID()
	repo.WebhookActive = true

	return repo, nil
}

func (*GitHubManager) DeleteWebhook(ctx context.Context, identity *models.Identity, repo *models.Repo) error {
	client := getClientByIdentity(ctx, identity)

	_, err := client.Repositories.DeleteHook(ctx, identity.Username, repo.Name, repo.WebhookID)
	return err
}

func (*GitHubManager) ChangeWebhookState(ctx context.Context, identity *models.Identity, repo *models.Repo, active bool) (*models.Repo, error) {
	client := getClientByIdentity(ctx, identity)

	ghHook := defaultWebhook()
	ghHook.Active = github.Bool(active)
	_, _, err := client.Repositories.EditHook(ctx, identity.Username, repo.Name, repo.WebhookID, ghHook)
	if err != nil {
		return nil, err
	}

	repo.WebhookActive = active
	return repo, nil
}

func (ghm *GitHubManager) CreateJob(ctx context.Context, r *http.Request) (*models.Job, error) {
	payload, err := github.ValidatePayload(r, []byte(configs.WebhookSecret))
	if err != nil {
		return nil, err
	}

	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		return nil, err
	}

	switch event := event.(type) {
	case *github.PushEvent:
		// TODO: Should this be commit which triggred webhook or last commit in repo?
		commit := event.Commits[len(event.Commits)-1]

		username := event.Repo.Owner.GetLogin()
		identity, err := db.GetIdentityByUsername(ctx, username, ghm.name)
		if err != nil {
			return nil, err
		}

		repo, err := db.GetRepoByFullName(ctx, event.Repo.GetFullName(), ghm.name)
		if err != nil {
			return nil, err
		}

		job := &models.Job{
			RepoID:    repo.ID,
			CommitSHA: commit.GetID(),
			CloneURL:  event.Repo.GetCloneURL(),
			Token: models.OAuth2Token{
				AccessToken:  identity.Token.AccessToken,
				TokenType:    identity.Token.TokenType,
				RefreshToken: identity.Token.RefreshToken,
				Expiry:       identity.Token.Expiry,
			},
		}

		return job, nil
	}
	return nil, nil
}

func (ghm *GitHubManager) CreateStatus(ctx context.Context, identity *models.Identity, job *models.Job, status Status) error {
	client := getClientByIdentity(ctx, identity)
	repo, err := ghm.store.GetRepo(ctx, job.RepoID)
	if err != nil {
		return err
	}

	statusName, err := ghm.StatusName(status.State)
	if err != nil {
		return err
	}

	s := &github.RepoStatus{
		State:       github.String(statusName),
		TargetURL:   github.String(status.TargetURL),
		Context:     github.String(status.Context),
		Description: github.String(status.Description),
	}
	_, _, err = client.Repositories.CreateStatus(ctx, identity.Username, repo.Name, job.CommitSHA, s)

	return err
}

func defaultWebhook() *github.Hook {
	return &github.Hook{
		Active: github.Bool(true),
		Events: []string{"push", "pull_request"},
		Config: map[string]any{
			"url":          fmt.Sprintf("https://%s:%s%s", configs.Host, configs.Port, ghm.eventHandlerPath),
			"content_type": "json",
			"secret":       configs.WebhookSecret,
		},
	}
}

func getGitHubClient(ctx context.Context, token *oauth2.Token) *github.Client {
	client := GitHub.oauth2Config.Client(ctx, token)
	return github.NewClient(client)
}

func getClientByIdentity(ctx context.Context, identity *models.Identity) *github.Client {
	var t time.Time
	if identity.Token.Expiry != nil {
		t = *identity.Token.Expiry
	}

	token := oauth2.Token{
		AccessToken:  identity.Token.AccessToken,
		TokenType:    identity.Token.TokenType,
		RefreshToken: identity.Token.RefreshToken,
		Expiry:       t,
	}

	return getGitHubClient(ctx, &token)
}
