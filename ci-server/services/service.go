package services

import (
	"context"
	"net/http"

	"github.com/shark-ci/shark-ci/models"
	"golang.org/x/oauth2"
)

// TODO: Change name to VCS

const (
	StatusSuccess StatusState = iota // GitHub -> Success, GitLab -> Success
	StatusPending                    // GitHub -> Pendign, GitLab -> Pending
	StatusRunning                    // GitHub -> Pending, GitLab -> Running
	StatusError                      // GitHub -> Error, GitLab -> Failed
)

type StatusState int

var StatusStateMap = map[string]StatusState{
	"success": StatusSuccess,
	"pending": StatusPending,
	"running": StatusRunning,
	"error":   StatusError,
}

var Services = map[string]ServiceManager{}

type Status struct {
	State       StatusState
	TargetURL   string
	Context     string
	Description string
}

type ServiceManager interface {
	GetServiceName() string          // Return service name.
	GetOAuth2Config() *oauth2.Config // Return OAuth2 config.

	// Get or create user with OAuth2 token.
	// Also creates new user profile if user does not exist.
	GetOrCreateUserIdentity(ctx context.Context, user *models.User, token *oauth2.Token) (*models.Identity, error)

	// Return user's repos on from service.
	GetUsersRepos(ctx context.Context, identity *models.Identity) ([]*models.Repo, error)

	CreateWebhook(ctx context.Context, identity *models.Identity, repo *models.Repo) (*models.Webhook, error)
	DeleteWebhook(ctx context.Context, identity *models.Identity, repo *models.Repo, hook *models.Webhook) error
	ChangeWebhookState(ctx context.Context, identity *models.Identity, repo *models.Repo, hook *models.Webhook, active bool) (*models.Webhook, error)

	// Create new job from HTTP request.
	CreateJob(ctx context.Context, r *http.Request) (*models.Job, error)

	GetStatusName(status StatusState) (string, bool)
	CreateStatus(ctx context.Context, identity *models.Identity, job *models.Job, status Status) error
}
