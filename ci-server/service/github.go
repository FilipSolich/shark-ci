package service

import (
	"context"
	"net/http"

	"github.com/google/go-github/v53/github"
	"golang.org/x/oauth2"
	oauth2_github "golang.org/x/oauth2/github"

	"github.com/FilipSolich/shark-ci/ci-server/config"
	"github.com/FilipSolich/shark-ci/ci-server/store"
	"github.com/FilipSolich/shark-ci/shared/model2"
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

func (*GitHubManager) Name() string {
	return "GitHub"
}

func (*GitHubManager) StatusName(status StatusState) string {
	switch status {
	case StatusSuccess:
		return "success"
	case StatusPending:
		return "pending"
	case StatusRunning:
		return "pending"
	case StatusError:
		return "error"
	default:
		return ""
	}
}

func (m *GitHubManager) OAuth2Config() *oauth2.Config {
	return m.oauth2Config
}

func (m *GitHubManager) GetServiceUser(ctx context.Context, token *oauth2.Token) (*model2.ServiceUser, error) {
	client := m.clientWithToken(ctx, token)

	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return nil, err
	}

	serviceUser := &model2.ServiceUser{
		Username:     user.GetLogin(),
		Email:        user.GetEmail(),
		Service:      m.Name(),
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		TokenExpire:  token.Expiry,
	}
	return serviceUser, nil
}

func (m *GitHubManager) GetUsersRepos(ctx context.Context, serviceUser *model2.ServiceUser) ([]model2.Repo, error) {
	client := m.clientForServiceUser(ctx, serviceUser)

	// TODO: Experimenting feature - get all repos, not just owned by user.
	// TODO: Add pagination.
	r, _, err := client.Repositories.List(ctx, "", &github.RepositoryListOptions{})

	repos := make([]model2.Repo, 0, len(r))
	for _, repo := range r {
		if repo.GetArchived() {
			continue
		}
		repo := model2.Repo{
			RepoServiceID: repo.GetID(),
			Name:          repo.GetName(),
			Service:       m.Name(),
			ServiceUserID: serviceUser.ID,
		}
		repos = append(repos, repo)
	}

	return repos, err
}

func (m *GitHubManager) CreateWebhook(ctx context.Context, serviceUser *model2.ServiceUser, repoName string) (int64, error) {
	client := m.clientForServiceUser(ctx, serviceUser)

	hook := &github.Hook{
		Active: github.Bool(true),
		Events: []string{"push", "pull_request"},
		Config: map[string]any{
			"url":          m.config.WebhookEndpoint + "/" + m.Name(),
			"content_type": "json",
			"secret":       m.config.SecretKey,
		},
	}

	hook, _, err := client.Repositories.CreateHook(ctx, serviceUser.Username, repoName, hook)
	if err != nil {
		return 0, err
	}

	return hook.GetID(), nil
}

func (m *GitHubManager) DeleteWebhook(ctx context.Context, serviceUser *model2.ServiceUser, repoName string, webhookID int64) error {
	client := m.clientForServiceUser(ctx, serviceUser)

	_, err := client.Repositories.DeleteHook(ctx, serviceUser.Username, repoName, webhookID)
	return err
}

func (m *GitHubManager) HandleEvent(ctx context.Context, r *http.Request) (*model2.Pipeline, error) {
	payload, err := github.ValidatePayload(r, []byte(m.config.SecretKey))
	if err != nil {
		return nil, err
	}
	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		return nil, err
	}

	switch event := event.(type) {
	case *github.PushEvent:
		return m.handlePush(ctx, event)
	case *github.PingEvent:
		return nil, NoErrPingEvent
	default:
		return nil, ErrEventNotSupported
	}
}

func (m *GitHubManager) handlePush(ctx context.Context, e *github.PushEvent) (*model2.Pipeline, error) {
	commit := e.HeadCommit.GetSHA()

	repoID, err := m.s.GetRepoIDByServiceRepoID(ctx, m.Name(), e.Repo.GetID())
	if err != nil {
		return nil, err
	}

	pipeline := &model2.Pipeline{
		CommitSHA: commit,
		CloneURL:  e.Repo.GetCloneURL(),
		Status:    m.StatusName(StatusPending),
		RepoID:    repoID,
	}
	pipeline.CreateTargetURL()

	return pipeline, nil
}

func (m *GitHubManager) CreateStatus(ctx context.Context, serviceUser *model2.ServiceUser, repoName string, commit string, status Status) error {
	client := m.clientForServiceUser(ctx, serviceUser)

	s := &github.RepoStatus{
		State:       github.String(m.StatusName(status.State)),
		TargetURL:   github.String(status.TargetURL),
		Context:     github.String(status.Context),
		Description: github.String(status.Description),
	}
	_, _, err := client.Repositories.CreateStatus(ctx, serviceUser.Username, repoName, commit, s)

	return err
}

func (m *GitHubManager) clientWithToken(ctx context.Context, token *oauth2.Token) *github.Client {
	client := m.oauth2Config.Client(ctx, token)
	return github.NewClient(client)
}

func (m *GitHubManager) clientForServiceUser(ctx context.Context, serviceUser *model2.ServiceUser) *github.Client {
	return m.clientWithToken(ctx, &oauth2.Token{
		AccessToken:  serviceUser.AccessToken,
		RefreshToken: serviceUser.RefreshToken,
		TokenType:    serviceUser.TokenType,
		Expiry:       serviceUser.TokenExpire,
	})
}
