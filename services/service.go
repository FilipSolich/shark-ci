package services

import (
	"context"

	"github.com/FilipSolich/ci-server/models"
	"golang.org/x/oauth2"
)

var Services = map[string]ServiceManager{}

//type StatusState int

// GitHub: Success, Pending, Error, Failure
// GitLab: Success, Pending, Running, Failed, Canceled

//const (
//	StatusSuccess  StatusState = iota // GitHub -> Success, GitLab -> Success
//	StatusPending                     // GitHub -> Pendign, GitLab -> Pending
//	StatusRunning                     // GitHub -> Pending, GitLab -> Running
//	StatusCanceled                    // GitHub -> Error, GitLab -> Canceled
//	StatusError                       // GitHub -> Error, GitLab -> Failed
//)
//
//type RepoInfo struct {
//	ID       int64
//	Name     string
//	FullName string
//}
//
//type CommitInfo struct {
//	SHA string
//}
//
//type Status struct {
//	State       string
//	TargetUrl   string
//	Description string
//	Context     string
//}

type ServiceManager interface {
	GetServiceName() string          // Return service name.
	GetOAuth2Config() *oauth2.Config // Return OAuth2 config.

	GetOrCreateUserIdentity(ctx context.Context, token *oauth2.Token) (*models.UserIdentity, error) // Get or create user using OAuth2 token.

	GetUsersRepos(user *models.User) ([]*models.Repository, error) // Return user's repos on from service.

	// Legacy code
	//GetStatusName(status StatusState) string

	//CreatWebhook(ctx context.Context, user *models.User, repo RepoInfo) (*models.Webhook, error)
	//DeleteWebhook(ctx context.Context, user *models.User, hook *models.Webhook) error
	//ActivateWebhook(ctx context.Context, user *models.User, hook *models.Webhook) (*models.Webhook, error)
	//DeactivateWebhook(ctx context.Context, user *models.User, hook *models.Webhook) (*models.Webhook, error)

	//CreateStatus(ctx context.Context, user *models.User, repo RepoInfo, commit CommitInfo, status Status) error
}
