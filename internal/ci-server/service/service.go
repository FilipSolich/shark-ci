package service

import (
	"context"
	"errors"
	"net/http"

	"github.com/shark-ci/shark-ci/internal/ci-server/models"
	"github.com/shark-ci/shark-ci/internal/ci-server/store"
	"github.com/shark-ci/shark-ci/internal/config"
	"golang.org/x/oauth2"
)

var ErrEventNotSupported = errors.New("event is not supported")
var NoErrPingEvent = errors.New("ping event")

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
	if config.CIServerConf.GitHub.ClientID != "" && config.CIServerConf.GitHub.ClientSecret != "" {
		ghm := NewGitHubManager(config.CIServerConf.GitHub.ClientID, config.CIServerConf.GitHub.ClientSecret, s)
		services[ghm.Name()] = ghm
	}
	// TODO: Add GitLab.
	return services
}

type ServiceManager interface {
	Name() string
	StatusName(status StatusState) string
	OAuth2Config() *oauth2.Config

	GetServiceUser(ctx context.Context, token *oauth2.Token) (*models.ServiceUser, error)
	GetUsersRepos(ctx context.Context, token *oauth2.Token, serviceUserID int64) ([]models.Repo, error)

	CreateWebhook(ctx context.Context, token *oauth2.Token, owner string, repoName string) (int64, error)
	DeleteWebhook(ctx context.Context, token *oauth2.Token, owner string, repoName string, webhookID int64) error

	HandleEvent(ctx context.Context, r *http.Request) (*models.Pipeline, error)

	CreateStatus(ctx context.Context, token *oauth2.Token, owner string, repoName string, commit string, status Status) error
}
