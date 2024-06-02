package store

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/shark-ci/shark-ci/internal/types"
)

type Storer interface {
	Ping(ctx context.Context) error
	Close(ctx context.Context) error
	Clean(ctx context.Context) error

	GetAndDeleteOAuth2State(ctx context.Context, state uuid.UUID) (types.OAuth2State, error)
	CreateOAuth2State(ctx context.Context, state types.OAuth2State) error

	GetUser(ctx context.Context, userID int64) (types.User, error)
	GetUserIDByServiceUser(ctx context.Context, service types.Service, username string) (int64, error)
	CreateUserAndServiceUser(ctx context.Context, serviceUser types.ServiceUser) (int64, int64, error)
	GetServiceUserByUserID(ctx context.Context, service types.Service, userID int64) (types.ServiceUser, error)

	GetRepoIDByServiceRepoID(ctx context.Context, service types.Service, serviceRepoID int64) (int64, error)
	GetUserRepos(ctx context.Context, userID int64) ([]types.Repo, error)
	CreateRepo(ctx context.Context, repo types.Repo) (int64, error)
	DeleteRepo(ctx context.Context, repoID int64) error

	GetPipelineCreationInfo(ctx context.Context, repoID int64) (*types.PipelineCreationInfo, error)
	GetPipelineStateChangeInfo(ctx context.Context, pipelineID int64) (*types.PipelineStateChangeInfo, error)
	CreatePipeline(ctx context.Context, pipeline *types.Pipeline) (int64, error)
	PipelineStarted(ctx context.Context, pipelineID int64, status types.PipelineStatus, startedAt time.Time) error
	PipelineFinnished(ctx context.Context, pipelineID int64, status types.PipelineStatus, finnisedAt time.Time) error

	CreatePipelineLog(ctx context.Context, log types.PipelineLog) (int64, error)
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
