package store

import (
	"context"

	"github.com/FilipSolich/shark-ci/model"
	"golang.org/x/oauth2"
)

// TODO: Split to multiple storers
// All Create... methods should set ID to created item
type Storer interface {
	Ping(ctx context.Context) error
	Close(ctx context.Context) error

	GetUser(ctx context.Context, id string) (*model.User, error)
	GetUserByIdentity(ctx context.Context, i *model.Identity) (*model.User, error)
	CreateUser(ctx context.Context, u *model.User) error
	DeleteUser(ctx context.Context, u *model.User) error

	GetIdentity(ctx context.Context, id string) (*model.Identity, error)
	GetIdentityByUniqueName(ctx context.Context, uniqueName string) (*model.Identity, error)
	GetIdentityByRepo(ctx context.Context, r *model.Repo) (*model.Identity, error)
	GetIdentityByUser(ctx context.Context, user *model.User, serviceName string) (*model.Identity, error)
	CreateIdentity(ctx context.Context, i *model.Identity) error
	UpdateIdentityToken(ctx context.Context, i *model.Identity, token oauth2.Token) error
	DeleteIdentity(ctx context.Context, i *model.Identity) error

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
