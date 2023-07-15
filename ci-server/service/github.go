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
	s            store.Storer
	oauth2Config *oauth2.Config
	config       config.CIServerConfig
}

var _ ServiceManager = &GitHubManager{}

func NewGitHubManager(clientID string, clientSecret string, s store.Storer, config config.CIServerConfig) *GitHubManager {
	return &GitHubManager{
		s: s,
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
func (*GitHubManager) Name() string {
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

func (m *GitHubManager) OAuth2Config() *oauth2.Config {
	return m.oauth2Config
}

func (m *GitHubManager) GetServiceUser(ctx context.Context, token *oauth2.Token) (*model.ServiceUser, error) {
	client := m.getGitHubClient(ctx, token)

	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return nil, err
	}

	serviceUser := model.NewServiceUser(user.GetLogin(), m.Name(), token)
	return serviceUser, nil
}

func (m *GitHubManager) GetUsersRepos(ctx context.Context, serviceUser *model.ServiceUser) ([]*model.Repo, error) {
	client := m.getClientByServiceUser(ctx, serviceUser)

	// TODO: Experimenting feature - get all repos, not just owned by user.
	// TODO: Add pagination.
	r, _, err := client.Repositories.List(ctx, "", &github.RepositoryListOptions{})

	repos := make([]*model.Repo, 0, len(r))
	for _, repo := range r {
		if repo.GetArchived() {
			continue
		}
		repo := model.NewRepo(serviceUser, repo.GetID(), m.Name(), repo.GetName(), repo.GetFullName())
		repos = append(repos, repo)
	}

	return repos, err
}

func (m *GitHubManager) CreateWebhook(ctx context.Context, serviceUser *model.ServiceUser, repo *model.Repo) (*model.Repo, error) {
	client := m.getClientByServiceUser(ctx, serviceUser)

	hook := m.defaultWebhook()
	hook, _, err := client.Repositories.CreateHook(ctx, serviceUser.Username, repo.Name, hook)
	if err != nil {
		return nil, err
	}

	repo.WebhookID = hook.GetID()
	repo.WebhookActive = true

	return repo, nil
}

func (m *GitHubManager) DeleteWebhook(ctx context.Context, serviceUser *model.ServiceUser, repo *model.Repo) error {
	client := m.getClientByServiceUser(ctx, serviceUser)

	_, err := client.Repositories.DeleteHook(ctx, serviceUser.Username, repo.Name, repo.WebhookID)
	repo.DeleteWebhook()
	return err
}

func (m *GitHubManager) ChangeWebhookState(ctx context.Context, serviceUser *model.ServiceUser, repo *model.Repo, active bool) (*model.Repo, error) {
	client := m.getClientByServiceUser(ctx, serviceUser)

	ghHook := m.defaultWebhook()
	ghHook.Active = github.Bool(active)
	_, _, err := client.Repositories.EditHook(ctx, serviceUser.Username, repo.Name, repo.WebhookID, ghHook)
	if err != nil {
		return nil, err
	}

	repo.WebhookActive = active
	return repo, nil
}

func (m *GitHubManager) HandleEvent(r *http.Request) (*model.Job, error) {
	//payload, err := github.ValidatePayload(r, []byte(m.config.SecretKey))
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
		return m.handlePush(r.Context(), event)
	case *github.PingEvent:
		return nil, NoErrPingEvent
	default:
		return nil, ErrEventNotSupported
	}
}

func (m *GitHubManager) handlePush(ctx context.Context, e *github.PushEvent) (*model.Job, error) {
	// TODO: Should this be commit which triggred webhook or last commit in repo?
	commit := e.Commits[len(e.Commits)-1]

	username := e.Repo.Owner.GetLogin()
	serviceUser, err := m.s.GetServiceUserByUniqueName(ctx, m.Name()+"/"+username)
	if err != nil {
		return nil, err
	}

	repo, err := m.s.GetRepoByUniqueName(ctx, m.Name()+"/"+e.Repo.GetFullName())
	if err != nil {
		return nil, err
	}

	job := model.NewJob(repo.ID, repo.UniqueName, commit.GetID(), e.Repo.GetCloneURL(), serviceUser.Token)
	return job, nil
}

func (m *GitHubManager) CreateStatus(ctx context.Context, serviceUser *model.ServiceUser, repo *model.Repo, job *model.Job, status Status) error {
	client := m.getClientByServiceUser(ctx, serviceUser)

	statusName, err := m.StatusName(status.State)
	if err != nil {
		return err
	}

	s := &github.RepoStatus{
		State:       github.String(statusName),
		TargetURL:   github.String(status.TargetURL),
		Context:     github.String(status.Context),
		Description: github.String(status.Description),
	}
	_, _, err = client.Repositories.CreateStatus(ctx, serviceUser.Username, repo.Name, job.CommitSHA, s)

	return err
}

func (m *GitHubManager) defaultWebhook() *github.Hook {
	return &github.Hook{
		Active: github.Bool(true),
		Events: []string{"push", "pull_request"},
		Config: map[string]any{
			"url":          fmt.Sprintf("https://%s:%s%s", m.config.Host, m.config.Port, ciserver.EventHandlerPath+"/"+m.Name()),
			"content_type": "json",
			"secret":       m.config.SecretKey,
		},
	}
}

func (m *GitHubManager) getGitHubClient(ctx context.Context, token *oauth2.Token) *github.Client {
	client := m.oauth2Config.Client(ctx, token)
	return github.NewClient(client)
}

func (m *GitHubManager) getClientByServiceUser(ctx context.Context, serviceUser *model.ServiceUser) *github.Client {
	return m.getGitHubClient(ctx, &serviceUser.Token)
}
