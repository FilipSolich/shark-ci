package store

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/shark-ci/shark-ci/internal/server/db"
	"github.com/shark-ci/shark-ci/internal/server/models"
	"github.com/shark-ci/shark-ci/internal/server/types"
)

type Storer interface {
	Ping(ctx context.Context) error
	Close(ctx context.Context) error
	Clean(ctx context.Context) error

	GetAndDeleteOAuth2State(ctx context.Context, state uuid.UUID) (types.OAuth2State, error)
	CreateOAuth2State(ctx context.Context, state types.OAuth2State) error

	GetUser(ctx context.Context, userID int64) (types.User, error)
	GetUserID(ctx context.Context, service string, username string) (int64, error)
	CreateUserAndServiceUser(ctx context.Context, serviceUser types.ServiceUser) (int64, int64, error)

	GetServiceUserByUserID(ctx context.Context, service string, userID int64) (types.ServiceUser, error)

	GetRepoIDByServiceRepoID(ctx context.Context, service string, serviceRepoID int64) (int64, error)
	GetUserRepos(ctx context.Context, userID int64) ([]types.Repo, error)

	GetRepoWebhookChangeInfo(ctx context.Context, repoID int64) (*types.RepoWebhookChangeInfo, error)
	GetRegisterWebhookInfoByRepo(ctx context.Context, repoID int64) (db.GetRegisterWebhookInfoRow, error)
	CreateOrUpdateRepos(ctx context.Context, repos []models.Repo) error
	UpdateRepoWebhook(ctx context.Context, repoID int64, webhookID *int64) error

	//GetPipeline(ctx context.Context, pipelineID int64) (*models.Pipeline, error)
	GetPipelineCreationInfo(ctx context.Context, repoID int64) (*types.PipelineCreationInfo, error)
	CreatePipeline(ctx context.Context, pipeline *models.Pipeline) (int64, error)
	UpdatePipelineStatus(ctx context.Context, pipelineID int64, status string, started_at *time.Time, finished_at *time.Time) error

	GetPipelineStateChangeInfo(ctx context.Context, pipelineID int64) (*types.PipelineStateChangeInfo, error)
}

func Cleaner(s Storer, d time.Duration) {
	ticker := time.NewTicker(d)
	go func() {
		for {
			<-ticker.C
			err := s.Clean(context.TODO())
			if err != nil {
				slog.Warn("Cannot clean DB", "err", err)
			}
		}
	}()
}
