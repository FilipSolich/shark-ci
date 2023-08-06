package service

import (
	"context"
	"errors"
	"net/http"

	"github.com/FilipSolich/shark-ci/ci-server/config"
	"github.com/FilipSolich/shark-ci/ci-server/models"
	"github.com/FilipSolich/shark-ci/ci-server/store"
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

func InitServices(s store.Storer, config config.Config) Services {
	services := Services{}
	if config.GitHub.ClientID != "" && config.GitHub.ClientSecret != "" {
		ghm := NewGitHubManager(config.GitHub.ClientID, config.GitHub.ClientSecret, s, config.CIServer)
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
	GetUsersRepos(ctx context.Context, serviceUser *models.ServiceUser) ([]models.Repo, error)

	CreateWebhook(ctx context.Context, serviceUser *models.ServiceUser, repoName string) (int64, error)
	DeleteWebhook(ctx context.Context, serviceUser *models.ServiceUser, repoName string, webhookID int64) error

	HandleEvent(ctx context.Context, r *http.Request) (*models.Pipeline, error)

	CreateStatus(ctx context.Context, token *oauth2.Token, owner string, repoName string, commit string, status Status) error
}
