package store

import (
	"context"

	"github.com/shark-ci/shark-ci/models"
)

type Storer interface {
	Ping(ctx context.Context) error
	Close(ctx context.Context) error
	Migrate(ctx context.Context) error

	GetUser(ctx context.Context, id string) (*models.User, error)
	GetUserByIdentity(ctx context.Context, i *models.Identity) (*models.User, error)
	CreateUser(ctx context.Context, u *models.User) error
	UpdateUser(ctx context.Context, u *models.User) error
	DeleteUser(ctx context.Context, u *models.User) error

	//GetRepo(ctx context.Context, id string) (*models.Repo, error)
	//GetRepoByFullName(ctx context.Context, fullName string) (*models.Repo, error)
	//CreateRepo(ctx context.Context, r *models.Repo) error
	//UpdateRepo(ctx context.Context, r *models.Repo) error
	//DeleteRepo(ctx context.Context, r *models.Repo) error

	CreateOAuth2State(ctx context.Context, s *models.OAuth2State) error
}
