package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/go-github/v49/github"
	"golang.org/x/oauth2"
	oauth2_github "golang.org/x/oauth2/github"

	"github.com/shark-ci/shark-ci/ci-server/configs"
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
	ghClient := ghm.getGitHubClient(ctx, token)

	ghUser, _, err := ghClient.Users.Get(ctx, "")
	if err != nil {
		return nil, err
	}

	identity := models.NewIdentity(ghUser.GetLogin(), ghm.name, token)
	return identity, nil
}

// Get repositories which aren't archived and are owned by user `identity`.
func (ghm *GitHubManager) GetUsersRepos(ctx context.Context, identity *models.Identity) ([]*models.Repo, error) {
	client := ghm.getClientByIdentity(ctx, identity)

	ghRepos, _, err := client.Repositories.List(ctx, "", &github.RepositoryListOptions{Type: "owner"})

	repos := make([]*models.Repo, 0, len(ghRepos))
	for _, repo := range ghRepos {
		if repo.GetArchived() {
			continue
		}
		r := models.NewRepo(identity, repo.GetID(), ghm.ServiceName(), repo.GetName(), repo.GetFullName())
		repos = append(repos, r)
	}

	return repos, err
}

func (ghm *GitHubManager) CreateWebhook(ctx context.Context, identity *models.Identity, repo *models.Repo) (*models.Repo, error) {
	client := ghm.getClientByIdentity(ctx, identity)

	hook := ghm.defaultWebhook()
	hook, _, err := client.Repositories.CreateHook(ctx, identity.Username, repo.Name, hook)
	if err != nil {
		return nil, err
	}

	repo.WebhookID = hook.GetID()
	repo.WebhookActive = true

	return repo, nil
}

func (ghm *GitHubManager) DeleteWebhook(ctx context.Context, identity *models.Identity, repo *models.Repo) error {
	client := ghm.getClientByIdentity(ctx, identity)

	_, err := client.Repositories.DeleteHook(ctx, identity.Username, repo.Name, repo.WebhookID)
	repo.DeleteWebhook()
	return err
}

func (ghm *GitHubManager) ChangeWebhookState(ctx context.Context, identity *models.Identity, repo *models.Repo, active bool) (*models.Repo, error) {
	client := ghm.getClientByIdentity(ctx, identity)

	ghHook := ghm.defaultWebhook()
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
		identity, err := ghm.store.GetIdentityByUniqueName(ctx, ghm.name+"/"+username)
		if err != nil {
			return nil, err
		}

		repo, err := ghm.store.GetRepoByUniqueName(ctx, ghm.name+"/"+event.Repo.GetFullName())
		if err != nil {
			return nil, err
		}

		job := models.NewJob(repo.ID, commit.GetID(), event.Repo.GetCloneURL(), identity.Token)
		return job, nil
	default:
		return nil, ErrEventNotSupported
	}
}

func (ghm *GitHubManager) CreateStatus(ctx context.Context, identity *models.Identity, job *models.Job, status Status) error {
	client := ghm.getClientByIdentity(ctx, identity)
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

func (ghm *GitHubManager) defaultWebhook() *github.Hook {
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

func (ghm *GitHubManager) getGitHubClient(ctx context.Context, token *oauth2.Token) *github.Client {
	client := ghm.oauth2Config.Client(ctx, token)
	return github.NewClient(client)
}

func (ghm *GitHubManager) getClientByIdentity(ctx context.Context, identity *models.Identity) *github.Client {
	return ghm.getGitHubClient(ctx, &identity.Token)
}
