package service

import (
	"context"
	"errors"
	"net/http"

	"github.com/FilipSolich/shark-ci/ci-server/config"
	"github.com/FilipSolich/shark-ci/ci-server/store"
	"github.com/FilipSolich/shark-ci/shared/model"
	"github.com/FilipSolich/shark-ci/shared/model2"
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

var StatusStateMap = map[string]StatusState{
	"success": StatusSuccess,
	"pending": StatusPending,
	"running": StatusRunning,
	"error":   StatusError,
}

type Status struct {
	State       StatusState
	TargetURL   string
	Context     string
	Description string
}

func NewStatus(state StatusState, targetURL string, ctx string, description string) Status {
	return Status{
		State:       state,
		TargetURL:   targetURL,
		Context:     ctx,
		Description: description,
	}
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
	StatusName(status StatusState) (string, error)
	OAuth2Config() *oauth2.Config

	GetServiceUser(ctx context.Context, token *oauth2.Token) (*model2.ServiceUser, error)
	GetUsersRepos(ctx context.Context, serviceUser *model.ServiceUser) ([]*model.Repo, error)

	CreateWebhook(ctx context.Context, serviceUser *model.ServiceUser, repo *model.Repo) (*model.Repo, error)
	DeleteWebhook(ctx context.Context, serviceUser *model.ServiceUser, repo *model.Repo) error
	ChangeWebhookState(ctx context.Context, serviceUser *model.ServiceUser, repo *model.Repo, active bool) (*model.Repo, error)

	HandleEvent(r *http.Request) (*model.Job, error)

	CreateStatus(ctx context.Context, serviceUser *model.ServiceUser, repo *model.Repo, job *model.Job, status Status) error
}
