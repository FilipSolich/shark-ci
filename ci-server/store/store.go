package store

import (
	"context"

	"github.com/FilipSolich/shark-ci/shared/model"
	"github.com/FilipSolich/shark-ci/shared/model2"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

// TODO: Split to multiple storers
type Storer interface {
	Ping(ctx context.Context) error
	Close(ctx context.Context) error

	// Cleanup expired OAuth2 states.
	Clean(ctx context.Context) error

	GetOAuth2State(ctx context.Context, state uuid.UUID) (*model2.OAuth2State, error)
	CreateOAuth2State(ctx context.Context, state *model2.OAuth2State) error
	DeleteOAuth2State(ctx context.Context, state *model2.OAuth2State) error

	GetUser(ctx context.Context, id int64) (*model2.User, error)
	CreateUserAndServiceUser(ctx context.Context, serviceUser *model2.ServiceUser) (int64, error)

	GetServiceUserByUniqueName(ctx context.Context, service string, username string) (*model2.ServiceUser, error)
	//GetServiceUserTokenByUniqueName(ctx context.Context, service string, username string) (*oauth2.Token, error) // TODO: Unused
	GetServiceUserByRepo(ctx context.Context, repoID int64) (*model2.ServiceUser, error)
	UpdateServiceUserToken(ctx context.Context, serviceUser *model2.ServiceUser, token *oauth2.Token) error

	GetRepoIDByServiceRepoID(ctx context.Context, service string, serviceRepoID int64) (int64, error)
	GetRepoName(ctx context.Context, repoID int64) (string, error)
	//GetTokenByRepo(ctx context.Context, repoID int64) (*oauth2.Token, error)

	CreatePipeline(ctx context.Context, pipeline *model2.Pipeline) error

	// -- TODO: Old API --

	//GetUser(ctx context.Context, id string) (*model.User, error)
	GetUserByServiceUser(ctx context.Context, i *model.ServiceUser) (*model.User, error)
	CreateUser(ctx context.Context, u *model.User) error
	DeleteUser(ctx context.Context, u *model.User) error

	GetServiceUser(ctx context.Context, id string) (*model.ServiceUser, error)
	//GetServiceUserByUniqueName(ctx context.Context, uniqueName string) (*model.ServiceUser, error)
	//GetServiceUserByRepo(ctx context.Context, r *model.Repo) (*model.ServiceUser, error)
	GetServiceUserByUser(ctx context.Context, user *model.User, serviceName string) (*model.ServiceUser, error)
	CreateServiceUser(ctx context.Context, i *model.ServiceUser) error
	//UpdateServiceUserToken(ctx context.Context, i *model.ServiceUser, token oauth2.Token) error
	DeleteServiceUser(ctx context.Context, i *model.ServiceUser) error

	GetRepo(ctx context.Context, id string) (*model.Repo, error)
	GetRepoByUniqueName(ctx context.Context, uniqueName string) (*model.Repo, error)
	CreateRepo(ctx context.Context, r *model.Repo) error
	UpdateRepoWebhook(ctx context.Context, r *model.Repo) error
	DeleteRepo(ctx context.Context, r *model.Repo) error

	GetOAuth2StateByState(ctx context.Context, state string) (*model.OAuth2State, error)

	GetJob(ctx context.Context, id string) (*model.Job, error)
	CreateJob(ctx context.Context, j *model.Job) error
	DeleteJob(ctx context.Context, j *model.Job) error
}
