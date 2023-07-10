package service

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/google/go-github/v53/github"
	"golang.org/x/oauth2"
	oauth2_github "golang.org/x/oauth2/github"

	ciserver "github.com/FilipSolich/shark-ci/ci-server"
	"github.com/FilipSolich/shark-ci/ci-server/config"
	"github.com/FilipSolich/shark-ci/ci-server/store"
	"github.com/FilipSolich/shark-ci/shared/model"
)

type GitHubManager struct {
	store        store.Storer
	oauth2Config *oauth2.Config
	config       config.CIServerConfig
}

var _ ServiceManager = &GitHubManager{}

func NewGitHubManager(clientID string, clientSecret string, store store.Storer, config config.CIServerConfig) *GitHubManager {
	return &GitHubManager{
		store: store,
		oauth2Config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Scopes:       []string{"repo"},
			Endpoint:     oauth2_github.Endpoint,
		},
		config: config,
	}
}

// Return service name.
func (ghm *GitHubManager) Name() string {
	return "GitHub"
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
	return "", fmt.Errorf("invalid state: %d", status)
}

func (ghm *GitHubManager) OAuth2Config() *oauth2.Config {
	return ghm.oauth2Config
}

func (ghm *GitHubManager) GetUserIdentity(ctx context.Context, token *oauth2.Token) (*model.Identity, error) {
	ghClient := ghm.getGitHubClient(ctx, token)

	ghUser, _, err := ghClient.Users.Get(ctx, "")
	if err != nil {
		return nil, err
	}

	identity := model.NewIdentity(ghUser.GetLogin(), ghm.Name(), token)
	return identity, nil
}

// Get repositories which aren't archived and are owned by user `identity`.
func (ghm *GitHubManager) GetUsersRepos(ctx context.Context, identity *model.Identity) ([]*model.Repo, error) {
	client := ghm.getClientByIdentity(ctx, identity)

	// TODO: Experimenting feature - get all repos, not just owned by user.
	// TODO: Add pagination.
	ghRepos, _, err := client.Repositories.List(ctx, "", &github.RepositoryListOptions{})

	repos := make([]*model.Repo, 0, len(ghRepos))
	for _, repo := range ghRepos {
		if repo.GetArchived() {
			continue
		}
		r := model.NewRepo(identity, repo.GetID(), ghm.Name(), repo.GetName(), repo.GetFullName())
		repos = append(repos, r)
	}

	return repos, err
}

func (ghm *GitHubManager) CreateWebhook(ctx context.Context, identity *model.Identity, repo *model.Repo) (*model.Repo, error) {
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

func (ghm *GitHubManager) DeleteWebhook(ctx context.Context, identity *model.Identity, repo *model.Repo) error {
	client := ghm.getClientByIdentity(ctx, identity)

	_, err := client.Repositories.DeleteHook(ctx, identity.Username, repo.Name, repo.WebhookID)
	repo.DeleteWebhook()
	return err
}

func (ghm *GitHubManager) ChangeWebhookState(ctx context.Context, identity *model.Identity, repo *model.Repo, active bool) (*model.Repo, error) {
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

func (ghm *GitHubManager) HandleEvent(r *http.Request) (*model.Job, error) {
	//payload, err := github.ValidatePayload(r, []byte(ghm.config.SecretKey))
	//if err != nil {
	//	return nil, err
	//}
	data, _ := ioutil.ReadAll(r.Body)
	event, err := github.ParseWebHook(github.WebHookType(r), data)
	if err != nil {
		return nil, err
	}

	switch event := event.(type) {
	case *github.PushEvent:
		return ghm.handlePush(r.Context(), event)
	case *github.PingEvent:
		return nil, NoErrPingEvent
	default:
		return nil, ErrEventNotSupported
	}
}

func (ghm *GitHubManager) handlePush(ctx context.Context, e *github.PushEvent) (*model.Job, error) {
	// TODO: Should this be commit which triggred webhook or last commit in repo?
	commit := e.Commits[len(e.Commits)-1]

	username := e.Repo.Owner.GetLogin()
	identity, err := ghm.store.GetIdentityByUniqueName(ctx, ghm.Name()+"/"+username)
	if err != nil {
		return nil, err
	}

	repo, err := ghm.store.GetRepoByUniqueName(ctx, ghm.Name()+"/"+e.Repo.GetFullName())
	if err != nil {
		return nil, err
	}

	job := model.NewJob(repo.ID, repo.UniqueName, commit.GetID(), e.Repo.GetCloneURL(), identity.Token)
	return job, nil
}

func (ghm *GitHubManager) CreateStatus(ctx context.Context, identity *model.Identity, job *model.Job, status Status) error {
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
			"url":          fmt.Sprintf("https://%s:%s%s", ghm.config.Host, ghm.config.Port, ciserver.EventHandlerPath+"/"+ghm.Name()),
			"content_type": "json",
			"secret":       ghm.config.SecretKey,
		},
	}
}

func (ghm *GitHubManager) getGitHubClient(ctx context.Context, token *oauth2.Token) *github.Client {
	client := ghm.oauth2Config.Client(ctx, token)
	return github.NewClient(client)
}

func (ghm *GitHubManager) getClientByIdentity(ctx context.Context, identity *model.Identity) *github.Client {
	return ghm.getGitHubClient(ctx, &identity.Token)
}
