package service

import (
	"context"
	"errors"
	"net/http"

	"golang.org/x/oauth2"

	"github.com/shark-ci/shark-ci/internal/config"
	"github.com/shark-ci/shark-ci/internal/server/models"
	"github.com/shark-ci/shark-ci/internal/server/store"
	"github.com/shark-ci/shark-ci/internal/server/types"
)

var ErrEventNotSupported = errors.New("event is not supported")

type StatusState int

const (
	StatusSuccess StatusState = iota // GitHub -> Success, GitLab -> Success
	StatusPending                    // GitHub -> Pendign, GitLab -> Pending
	StatusRunning                    // GitHub -> Pending, GitLab -> Running
	StatusError                      // GitHub -> Error, GitLab -> Failed
)

type Status struct {
	State       StatusState
	TargetURL   string
	Context     string
	Description string
}

type Services map[string]ServiceManager

func InitServices(s store.Storer) Services {
	services := Services{}
	if config.ServerConf.GitHub.ClientID != "" && config.ServerConf.GitHub.ClientSecret != "" {
		ghm := NewGitHubManager(config.ServerConf.GitHub.ClientID, config.ServerConf.GitHub.ClientSecret, s)
		services[ghm.Name()] = ghm
	}
	// TODO: Add GitLab.
	return services
}

type ServiceManager interface {
	Name() string
	StatusName(status StatusState) string
	OAuth2Config() *oauth2.Config

	GetServiceUser(ctx context.Context, token *oauth2.Token) (types.ServiceUser, error)
	GetUserRepos(ctx context.Context, token *oauth2.Token, serviceUserID int64) ([]types.Repo, error)

	CreateWebhook(ctx context.Context, token *oauth2.Token, owner string, repoName string) (int64, error)
	DeleteWebhook(ctx context.Context, token *oauth2.Token, owner string, repoName string, webhookID int64) error

	HandleEvent(ctx context.Context, w http.ResponseWriter, r *http.Request) (*models.Pipeline, error)

	CreateStatus(ctx context.Context, token *oauth2.Token, owner string, repoName string, commit string, status Status) error
}
