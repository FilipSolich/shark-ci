package service

import (
	"context"
	"errors"
	"net/http"

	"github.com/FilipSolich/shark-ci/ci-server/config"
	"github.com/FilipSolich/shark-ci/ci-server/store"
	"github.com/FilipSolich/shark-ci/shared/model"
	"golang.org/x/oauth2"
)

// TODO: Change name to VCS

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

type ServiceMap map[string]ServiceManager

func InitServices(s store.Storer, config config.Config) ServiceMap {
	serviceMap := ServiceMap{}
	if config.GitHub.ClientID != "" {
		ghm := NewGitHubManager(config.GitHub.ClientID, config.GitHub.ClientSecret, s, config.CIServer)
		serviceMap[ghm.Name()] = ghm
	}
	return serviceMap
}

type ServiceManager interface {
	Name() string                 // Return service name.
	OAuth2Config() *oauth2.Config // Return OAuth2 config.

	// Get or create user with OAuth2 token.
	// Also creates new user profile if user does not exist.
	GetUserIdentity(ctx context.Context, token *oauth2.Token) (*model.Identity, error)

	// Return user's repos on from service.
	GetUsersRepos(ctx context.Context, identity *model.Identity) ([]*model.Repo, error)

	CreateWebhook(ctx context.Context, identity *model.Identity, repo *model.Repo) (*model.Repo, error)
	DeleteWebhook(ctx context.Context, identity *model.Identity, repo *model.Repo) error
	ChangeWebhookState(ctx context.Context, identity *model.Identity, repo *model.Repo, active bool) (*model.Repo, error)

	// Create new job from HTTP request.
	HandleEvent(r *http.Request) (*model.Job, error)

	StatusName(status StatusState) (string, error)
	CreateStatus(ctx context.Context, identity *model.Identity, job *model.Job, status Status) error
}
