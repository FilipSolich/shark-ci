package services

import (
	"context"
	"net/http"

	"golang.org/x/oauth2"

	"github.com/FilipSolich/ci-server/db"
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
	GetOrCreateUserIdentity(ctx context.Context, user *db.User, token *oauth2.Token) (*db.Identity, error)

	// Return user's repos on from service.
	GetUsersRepos(ctx context.Context, identity *db.Identity) ([]*db.Repo, error)

	CreateWebhook(ctx context.Context, identity *db.Identity, repo *db.Repo) (*db.Webhook, error)
	DeleteWebhook(ctx context.Context, identity *db.Identity, repo *db.Repo, hook *db.Webhook) error
	ChangeWebhookState(ctx context.Context, identity *db.Identity, repo *db.Repo, hook *db.Webhook, active bool) (*db.Webhook, error)

	// Create new job from HTTP request.
	CreateJob(ctx context.Context, r *http.Request) (*db.Job, error)

	//GetStatusName(status StatusState) string
	CreateStatus(ctx context.Context, identity *db.Identity, repo *db.Repo, job *db.Job, status Status) error
}
