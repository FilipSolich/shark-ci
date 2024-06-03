package service

import (
	"context"
	"net/http"

	"github.com/google/go-github/v62/github"
	"golang.org/x/oauth2"
	oauth2_github "golang.org/x/oauth2/github"

	"github.com/shark-ci/shark-ci/internal/config"
	"github.com/shark-ci/shark-ci/internal/server/store"
	"github.com/shark-ci/shark-ci/internal/types"
)

type GitHubManager struct {
	s            store.Storer
	oauth2Config *oauth2.Config
}

var _ ServiceManager = &GitHubManager{}

func NewGitHubManager(clientID string, clientSecret string, s store.Storer) *GitHubManager {
	return &GitHubManager{
		s: s,
		oauth2Config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Scopes:       []string{"repo"},
			Endpoint:     oauth2_github.Endpoint,
		},
	}
}

func (*GitHubManager) Name() types.Service {
	return types.ServiceGitHub
}

func (*GitHubManager) StatusName(status types.PipelineStatus) string {
	switch status {
	case types.Success:
		return "success"
	case types.Pending:
		return "pending"
	case types.Running:
		return "pending"
	case types.Error:
		return "error"
	default:
		return ""
	}
}

func (m *GitHubManager) OAuth2Config() *oauth2.Config {
	return m.oauth2Config
}

func (m *GitHubManager) GetServiceUser(ctx context.Context, token *oauth2.Token) (types.ServiceUser, error) {
	client := m.clientWithToken(ctx, token)

	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return types.ServiceUser{}, err
	}

	serviceUser := types.ServiceUser{
		Username:    user.GetLogin(),
		Email:       user.GetEmail(),
		Service:     m.Name(),
		AccessToken: token.AccessToken,
		TokenType:   token.TokenType,
	}
	if token.RefreshToken != "" {
		serviceUser.RefreshToken = &token.RefreshToken
	}
	if !token.Expiry.IsZero() {
		serviceUser.TokenExpire = &token.Expiry
	}
	return serviceUser, nil
}

func (m *GitHubManager) GetUserRepos(ctx context.Context, token *oauth2.Token, serviceUserID int64) ([]types.Repo, error) {
	client := m.clientWithToken(ctx, token)
	r, _, err := client.Repositories.ListByAuthenticatedUser(ctx, &github.RepositoryListByAuthenticatedUserOptions{ListOptions: github.ListOptions{PerPage: 50}})
	if err != nil {
		return nil, err
	}

	repos := make([]types.Repo, 0, len(r))
	for _, repo := range r {
		if repo.GetArchived() {
			continue
		}
		repo := types.Repo{
			RepoServiceID: repo.GetID(),
			Name:          repo.GetName(),
			Owner:         repo.GetOwner().GetLogin(),
			Service:       m.Name(),
			ServiceUserID: serviceUserID,
		}
		repos = append(repos, repo)
	}

	return repos, err
}

func (m *GitHubManager) CreateWebhook(ctx context.Context, token *oauth2.Token, owner string, repoName string) (int64, error) {
	client := m.clientWithToken(ctx, token)

	hook := &github.Hook{
		Active: github.Bool(true),
		Events: []string{"push", "pull_request"},
		Config: &github.HookConfig{
			ContentType: github.String("json"),
			URL:         github.String(config.ServerConf.Host + "/event_handler/" + string(m.Name())),
			Secret:      github.String(config.ServerConf.SecretKey),
		},
	}

	hook, _, err := client.Repositories.CreateHook(ctx, owner, repoName, hook)
	if err != nil {
		return 0, err
	}

	return hook.GetID(), nil
}

func (m *GitHubManager) DeleteWebhook(ctx context.Context, token *oauth2.Token, owner string, repoName string, webhookID int64) error {
	client := m.clientWithToken(ctx, token)

	_, err := client.Repositories.DeleteHook(ctx, owner, repoName, webhookID)
	return err
}

func (m *GitHubManager) HandleEvent(ctx context.Context, w http.ResponseWriter, r *http.Request) (*types.Pipeline, error) {
	payload, err := github.ValidatePayload(r, []byte(config.ServerConf.SecretKey))
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
		w.Write([]byte("pong"))
		return nil, nil
	default:
		return nil, ErrEventNotSupported
	}
}

func (m *GitHubManager) handlePush(ctx context.Context, e *github.PushEvent) (*types.Pipeline, error) {
	commit := e.HeadCommit.GetID()
	repoID, err := m.s.GetRepoIDByServiceRepoID(ctx, m.Name(), e.Repo.GetID())
	if err != nil {
		return nil, err
	}

	pipeline := &types.Pipeline{
		CommitSHA: commit,
		CloneURL:  e.Repo.GetCloneURL(),
		Status:    types.Pending,
		RepoID:    repoID,
	}

	return pipeline, nil
}

func (m *GitHubManager) CreateStatus(ctx context.Context, token *oauth2.Token, owner string, repoName string, commit string, status Status) error {
	client := m.clientWithToken(ctx, token)

	s := &github.RepoStatus{
		State:       github.String(m.StatusName(status.State)),
		TargetURL:   github.String(status.TargetURL),
		Context:     github.String(status.Context),
		Description: github.String(status.Description),
	}
	_, _, err := client.Repositories.CreateStatus(ctx, owner, repoName, commit, s)

	return err
}

func (m *GitHubManager) clientWithToken(ctx context.Context, token *oauth2.Token) *github.Client {
	client := m.oauth2Config.Client(ctx, token)
	return github.NewClient(client)
}
