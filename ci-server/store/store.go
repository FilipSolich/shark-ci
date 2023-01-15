package store

import (
	"context"

	"github.com/shark-ci/shark-ci/models"
	"golang.org/x/oauth2"
)

// TODO: Split on multiple storers
// All Create... methods should set ID to created item
type Storer interface {
	Ping(ctx context.Context) error
	Close(ctx context.Context) error
	Migrate(ctx context.Context) error

	GetUser(ctx context.Context, id string) (*models.User, error)
	GetUserByIdentity(ctx context.Context, i *models.Identity) (*models.User, error)
	CreateUser(ctx context.Context, u *models.User) error
	//UpdateUser(ctx context.Context, u *models.User) error // TODO: Delete if unused
	DeleteUser(ctx context.Context, u *models.User) error

	GetIdentity(ctx context.Context, id string) (*models.Identity, error)
	GetIdentityByUniqueName(ctx context.Context, uniqueName string) (*models.Identity, error)
	GetIdentityByRepo(ctx context.Context, r *models.Repo) (*models.Identity, error)
	CreateIdentity(ctx context.Context, i *models.Identity) error
	UpdateIdentityToken(ctx context.Context, i *models.Identity, token oauth2.Token) error
	DeleteIdentity(ctx context.Context, i *models.Identity) error

	GetRepo(ctx context.Context, id string) (*models.Repo, error)
	GetRepoByUniqueName(ctx context.Context, uniqueName string) (*models.Repo, error)
	CreateRepo(ctx context.Context, r *models.Repo) error
	//CreateRepoWebhook(ctx context.Context, r *models.Repo) error
	//UpdateRepo(ctx context.Context, r *models.Repo) error
	DeleteRepo(ctx context.Context, r *models.Repo) error
	//DeleteRepoWebhook(ctx context.Context, r *models.Repo) error

	//GetOAuth2State(ctx context.Context, id string) (*models.OAuth2State, error) // TODO: Delete if unused
	GetOAuth2StateByState(ctx context.Context, state string) (*models.OAuth2State, error)
	CreateOAuth2State(ctx context.Context, s *models.OAuth2State) error
	//UpdateOAuth2State(ctx context.Context, s *models.OAuth2State) error // TODO: Delete if unused
	DeleteOAuth2State(ctx context.Context, s *models.OAuth2State) error

	GetJob(ctx context.Context, id string) (*models.Job, error)
	CreateJob(ctx context.Context, j *models.Job) error
	// UpdateJob(ctx context.Context) // TODO: Delete if unused
	DeleteJob(ctx context.Context, j *models.Job) error
}
