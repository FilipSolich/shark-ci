package services

import (
	"context"
	"net/http"

	"github.com/FilipSolich/ci-server/models"
	"golang.org/x/oauth2"
)

const (
	StatusSuccess StatusState = iota // GitHub -> Success, GitLab -> Success
	StatusPending                    // GitHub -> Pendign, GitLab -> Pending
	StatusRunning                    // GitHub -> Pending, GitLab -> Running
	StatusError                      // GitHub -> Error, GitLab -> Failed
)

var Services = map[string]ServiceManager{}

type StatusState int

//	type RepoInfo struct {
//		ID       int64
//		Name     string
//		FullName string
//	}
//
//	type CommitInfo struct {
//		SHA string
//	}

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
	GetOrCreateUserIdentity(ctx context.Context, token *oauth2.Token) (*models.UserIdentity, error)

	// Return user's repos on from service.
	GetUsersRepos(ctx context.Context, user *models.User) ([]*models.Repository, error)

	CreateWebhook(ctx context.Context, user *models.User, repo *models.Repository) (*models.Webhook, error)
	DeleteWebhook(ctx context.Context, user *models.User, repo *models.Repository, hook *models.Webhook) error
	ActivateWebhook(ctx context.Context, user *models.User, repo *models.Repository, hook *models.Webhook) (*models.Webhook, error)
	DeactivateWebhook(ctx context.Context, user *models.User, repo *models.Repository, hook *models.Webhook) (*models.Webhook, error)

	// Create new job from HTTP request.
	CreateJob(ctx context.Context, r *http.Request) (*models.Job, error)

	// Updates commit status.
	UpdateStatus(ctx context.Context, user *models.User, status Status, job *models.Job) error

	//GetStatusName(status StatusState) string
	//CreateStatus(ctx context.Context, user *models.User, repo RepoInfo, commit CommitInfo, status Status) error
}
