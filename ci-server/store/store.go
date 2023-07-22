package store

import (
	"context"

	"github.com/FilipSolich/shark-ci/shared/model"
	"golang.org/x/oauth2"
)

// TODO: Split to multiple storers
// All Create... methods should set ID to created item
type Storer interface {
	Ping(ctx context.Context) error
	Close(ctx context.Context) error

	// Is called as goroutine from main and should clean up expired OAuth2States
	Clean(ctx context.Context) error

	GetUser(ctx context.Context, id string) (*model.User, error)
	GetUserByServiceUser(ctx context.Context, i *model.ServiceUser) (*model.User, error)
	CreateUser(ctx context.Context, u *model.User) error
	DeleteUser(ctx context.Context, u *model.User) error

	GetServiceUser(ctx context.Context, id string) (*model.ServiceUser, error)
	GetServiceUserByUniqueName(ctx context.Context, uniqueName string) (*model.ServiceUser, error)
	GetServiceUserByRepo(ctx context.Context, r *model.Repo) (*model.ServiceUser, error)
	GetServiceUserByUser(ctx context.Context, user *model.User, serviceName string) (*model.ServiceUser, error)
	CreateServiceUser(ctx context.Context, i *model.ServiceUser) error
	UpdateServiceUserToken(ctx context.Context, i *model.ServiceUser, token oauth2.Token) error
	DeleteServiceUser(ctx context.Context, i *model.ServiceUser) error

	GetRepo(ctx context.Context, id string) (*model.Repo, error)
	GetRepoByUniqueName(ctx context.Context, uniqueName string) (*model.Repo, error)
	CreateRepo(ctx context.Context, r *model.Repo) error
	UpdateRepoWebhook(ctx context.Context, r *model.Repo) error
	DeleteRepo(ctx context.Context, r *model.Repo) error

	GetOAuth2StateByState(ctx context.Context, state string) (*model.OAuth2State, error)
	CreateOAuth2State(ctx context.Context, s *model.OAuth2State) error
	DeleteOAuth2State(ctx context.Context, s *model.OAuth2State) error

	GetJob(ctx context.Context, id string) (*model.Job, error)
	CreateJob(ctx context.Context, j *model.Job) error
	DeleteJob(ctx context.Context, j *model.Job) error
}
