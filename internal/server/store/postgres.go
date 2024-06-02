package store

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/oauth2"

	"github.com/shark-ci/shark-ci/internal/server/db"
	"github.com/shark-ci/shark-ci/internal/types"
)

type PostgresStore struct {
	conn    *pgx.Conn
	queries *db.Queries
}

var _ Storer = &PostgresStore{}

func NewPostgresStore(ctx context.Context, postgresURI string) (*PostgresStore, error) {
	conn, err := pgx.Connect(ctx, postgresURI)
	if err != nil {
		return nil, err
	}

	return &PostgresStore{
		conn:    conn,
		queries: db.New(conn),
	}, nil
}

func (s *PostgresStore) Ping(ctx context.Context) error {
	return s.conn.Ping(ctx)
}

func (s *PostgresStore) Close(ctx context.Context) error {
	return s.conn.Close(ctx)
}

func (s *PostgresStore) Clean(ctx context.Context) error {
	return s.queries.CleanOAuth2State(ctx)
}

func (s *PostgresStore) GetAndDeleteOAuth2State(ctx context.Context, state uuid.UUID) (types.OAuth2State, error) {
	expire, err := s.queries.GetAndDeleteOAuth2State(ctx, state)
	if err != nil {
		return types.OAuth2State{}, fmt.Errorf("cannot delete OAuth2State with state=%s: %w", state, err)
	}

	return types.OAuth2State{State: state, Expire: expire.Time}, nil
}

func (s *PostgresStore) CreateOAuth2State(ctx context.Context, state types.OAuth2State) error {
	return s.queries.CreateOAuth2State(ctx, db.CreateOAuth2StateParams{
		State:  state.State,
		Expire: pgtype.Timestamp{Time: state.Expire, Valid: true},
	})
}

func (s *PostgresStore) GetUser(ctx context.Context, userID int64) (types.User, error) {
	user, err := s.queries.GetUser(ctx, userID)
	if err != nil {
		return types.User{}, fmt.Errorf("cannot get user with id=%d: %w", userID, err)
	}

	return types.User{ID: user.ID, Username: user.Username, Email: user.Email}, nil
}

func (s *PostgresStore) GetUserIDByServiceUser(ctx context.Context, service types.Service, username string) (int64, error) {
	userID, err := s.queries.GetUserIDByServiceUser(ctx, db.GetUserIDByServiceUserParams{
		Service:  db.Service(service),
		Username: username,
	})
	if err != nil {
		return 0, fmt.Errorf("cannot get user ID with service=%s and username=%s: %w", service, username, err)
	}

	return userID, nil
}

func (s *PostgresStore) CreateUserAndServiceUser(ctx context.Context, serviceUser types.ServiceUser) (int64, int64, error) {
	tx, err := s.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, 0, fmt.Errorf("cannot begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	qtx := s.queries.WithTx(tx)

	userID, err := qtx.CreateUser(ctx, db.CreateUserParams{
		Username: serviceUser.Username,
		Email:    serviceUser.Email,
	})
	if err != nil {
		return 0, 0, fmt.Errorf("cannot create user: %w", err)
	}

	serviceUserID, err := qtx.CreateServiceUser(ctx, db.CreateServiceUserParams{
		Service:      db.Service(serviceUser.Service),
		Username:     serviceUser.Username,
		Email:        serviceUser.Email,
		AccessToken:  serviceUser.AccessToken,
		RefreshToken: NullableText(serviceUser.RefreshToken),
		TokenType:    serviceUser.TokenType,
		TokenExpire:  NullableTimestamp(serviceUser.TokenExpire),
		UserID:       userID,
	})
	if err != nil {
		return 0, 0, fmt.Errorf("cannot create service user: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("cannot commit transaction: %w", err)
	}

	return userID, serviceUserID, nil
}

func (s *PostgresStore) GetServiceUserByUserID(ctx context.Context, service types.Service, userID int64) (types.ServiceUser, error) {
	serviceUser, err := s.queries.GetServiceUserByUserID(ctx, db.GetServiceUserByUserIDParams{
		Service: db.Service(service),
		UserID:  userID,
	})
	if err != nil {
		return types.ServiceUser{}, fmt.Errorf("cannot get service user with service=%s and userID=%d: %w", service, userID, err)
	}

	return types.ServiceUser{
		ID:           serviceUser.ID,
		Service:      service,
		Username:     serviceUser.Username,
		Email:        serviceUser.Email,
		AccessToken:  serviceUser.AccessToken,
		RefreshToken: ValueText(serviceUser.RefreshToken),
		TokenType:    serviceUser.TokenType,
		TokenExpire:  ValueTime(serviceUser.TokenExpire),
		UserID:       userID,
	}, nil
}

func (s *PostgresStore) GetRepoIDByServiceRepoID(ctx context.Context, service types.Service, serviceRepoID int64) (int64, error) {
	var repoID int64
	err := s.conn.QueryRow(ctx, `SELECT id FROM public.repo WHERE service = $1 AND repo_service_id = $2`,
		service, serviceRepoID).
		Scan(&repoID)
	if err != nil {
		return 0, err
	}

	return repoID, nil
}

func (s *PostgresStore) GetUserRepos(ctx context.Context, userID int64) ([]types.Repo, error) {
	repos, err := s.queries.GetUserRepos(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user repos from DB: %w", err)
	}

	var result []types.Repo
	for _, repo := range repos {
		result = append(result, types.Repo{
			ID:            repo.ID,
			Service:       types.Service(repo.Service),
			Owner:         repo.Owner,
			Name:          repo.Name,
			RepoServiceID: repo.RepoServiceID,
			WebhookID:     repo.WebhookID,
			ServiceUserID: repo.ServiceUserID,
		})
	}

	return result, nil
}

func (s *PostgresStore) CreateRepo(ctx context.Context, repo types.Repo) (int64, error) {
	repoID, err := s.queries.CreateRepo(ctx, db.CreateRepoParams{
		Service:       db.Service(repo.Service),
		Owner:         repo.Owner,
		Name:          repo.Name,
		RepoServiceID: repo.RepoServiceID,
		WebhookID:     repo.WebhookID,
		ServiceUserID: repo.ServiceUserID,
	})
	if err != nil {
		return 0, fmt.Errorf("cannot create repo: %w", err)
	}

	return repoID, nil
}

func (s *PostgresStore) DeleteRepo(ctx context.Context, repoID int64) error {
	return s.queries.DeleteRepo(ctx, repoID)
}

func (s *PostgresStore) GetPipelineCreationInfo(ctx context.Context, repoID int64) (*types.PipelineCreationInfo, error) {
	res, err := s.queries.GetPipelineCreationInfo(ctx, repoID)
	if err != nil {
		return nil, err
	}

	return &types.PipelineCreationInfo{
		Username: res.Username,
		RepoName: res.Name,
		Token: oauth2.Token{
			AccessToken:  res.AccessToken,
			RefreshToken: res.RefreshToken.String,
			TokenType:    res.TokenType,
			Expiry:       res.TokenExpire.Time,
		},
	}, nil
}

func (s *PostgresStore) CreatePipeline(ctx context.Context, pipeline *types.Pipeline) (int64, error) {
	pipelineID, err := s.queries.CreatePipeline(ctx, db.CreatePipelineParams{
		Status:    db.PipelineStatus(pipeline.Status),
		CloneUrl:  pipeline.CloneURL,
		CommitSha: pipeline.CommitSHA,
		RepoID:    pipeline.RepoID,
	})
	if err != nil {
		return 0, err
	}

	pipeline.ID = pipelineID
	pipeline.CreateURL()
	err = s.queries.SetPipelineUrl(ctx, db.SetPipelineUrlParams{
		ID:  pipelineID,
		Url: NullableText(&pipeline.URL),
	})
	if err != nil {
		return 0, err
	}

	return pipeline.ID, nil
}

func (s *PostgresStore) PipelineStarted(ctx context.Context, pipelineID int64, status types.PipelineStatus, startedAt time.Time) error {
	return s.queries.PipelineStarted(ctx, db.PipelineStartedParams{
		ID:        pipelineID,
		Status:    db.PipelineStatus(status),
		StartedAt: pgtype.Timestamp{Time: startedAt, Valid: true},
	})
}

func (s *PostgresStore) PipelineFinnished(ctx context.Context, pipelineID int64, status types.PipelineStatus, finnisedAt time.Time) error {
	return s.queries.PipelineFinished(ctx, db.PipelineFinishedParams{
		ID:         pipelineID,
		Status:     db.PipelineStatus(status),
		FinishedAt: pgtype.Timestamp{Time: finnisedAt, Valid: true},
	})
}

func (s *PostgresStore) GetPipelineStateChangeInfo(ctx context.Context, pipelineID int64,
) (*types.PipelineStateChangeInfo, error) {
	res, err := s.queries.GetPipelineStateChangeInfo(ctx, pipelineID)
	if err != nil {
		return nil, err
	}

	return &types.PipelineStateChangeInfo{
		CommitSHA: res.CommitSha,
		URL:       res.Url.String,
		Service:   types.Service(res.Service),
		RepoOwner: res.Owner,
		RepoName:  res.Name,
		Token: oauth2.Token{
			AccessToken:  res.AccessToken,
			RefreshToken: res.RefreshToken.String,
			TokenType:    res.TokenType,
			Expiry:       res.TokenExpire.Time,
		},
		StartedAt: ValueTime(res.StartedAt),
	}, nil
}

func NullableText(ptr *string) pgtype.Text {
	if ptr == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *ptr, Valid: true}
}

func NullableTimestamp(ptr *time.Time) pgtype.Timestamp {
	if ptr == nil {
		return pgtype.Timestamp{Valid: false}
	}
	return pgtype.Timestamp{Time: *ptr, Valid: true}
}

func ValueText(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func ValueTime(value pgtype.Timestamp) *time.Time {
	if !value.Valid {
		return nil
	}
	return &value.Time
}
