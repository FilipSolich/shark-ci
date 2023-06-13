package services

import (
	"context"
	"errors"
	"net/http"

	"github.com/shark-ci/shark-ci/models"
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

type ServiceManager interface {
	ServiceName() string          // Return service name.
	OAuth2Config() *oauth2.Config // Return OAuth2 config.

	// Get or create user with OAuth2 token.
	// Also creates new user profile if user does not exist.
	GetUserIdentity(ctx context.Context, token *oauth2.Token) (*models.Identity, error)

	// Return user's repos on from service.
	GetUsersRepos(ctx context.Context, identity *models.Identity) ([]*models.Repo, error)

	CreateWebhook(ctx context.Context, identity *models.Identity, repo *models.Repo) (*models.Repo, error)
	DeleteWebhook(ctx context.Context, identity *models.Identity, repo *models.Repo) error
	ChangeWebhookState(ctx context.Context, identity *models.Identity, repo *models.Repo, active bool) (*models.Repo, error)

	// Create new job from HTTP request.
	HandleEvent(r *http.Request) (*models.Job, error)

	StatusName(status StatusState) (string, error)
	CreateStatus(ctx context.Context, identity *models.Identity, job *models.Job, status Status) error
}
