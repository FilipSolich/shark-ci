package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/oauth2"

	"github.com/shark-ci/shark-ci/internal/server/db"
	"github.com/shark-ci/shark-ci/internal/server/models"
	"github.com/shark-ci/shark-ci/internal/server/types"
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

func (s *PostgresStore) GetUserIDByServiceUser(ctx context.Context, service string, username string) (int64, error) {
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

func (s *PostgresStore) GetServiceUserByUserID(ctx context.Context, service string, userID int64) (types.ServiceUser, error) {
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

func (s *PostgresStore) GetRepoIDByServiceRepoID(ctx context.Context, service string, serviceRepoID int64) (int64, error) {
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
		return nil, fmt.Errorf("cannot get user repos with userID=%d: %w", userID, err)
	}

	var result []types.Repo
	for _, repo := range repos {
		result = append(result, types.Repo{
			ID:            repo.ID,
			Service:       string(repo.Service),
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

func (s *PostgresStore) GetRepoWebhookChangeInfo(ctx context.Context, repoID int64,
) (*types.RepoWebhookChangeInfo, error) {
	var (
		info         = types.RepoWebhookChangeInfo{RepoID: repoID}
		refreshToken sql.NullString
		expire       sql.NullTime
	)
	err := s.conn.QueryRow(ctx, ``+
		`SELECT r.service, r.owner, r.name, r.webhook_id,su.access_token,`+
		` su.refresh_token, su.token_type, su.token_expire, su.user_id `+
		`FROM public.repo r JOIN public.service_user su ON r.service_user_id = su.id `+
		`WHERE r.id = $1`,
		repoID).
		Scan(&info.Service, &info.RepoOwner, &info.RepoName, &info.WebhookID, &info.Token.AccessToken,
			&refreshToken, &info.Token.TokenType, &expire, &info.UserID)
	if err != nil {
		return nil, err
	}

	if refreshToken.Valid {
		info.Token.RefreshToken = refreshToken.String
	}
	if expire.Valid {
		info.Token.Expiry = expire.Time
	}

	return &info, nil
}

// TODO: Create with existing webhook including update if webhook was manualy deleted
func (s *PostgresStore) CreateOrUpdateRepos(ctx context.Context, repos []models.Repo) error {
	query := `INSERT INTO public.repo (service, owner, name, repo_service_id, webhook_id, service_user_id) VALUES `
	values := []any{}
	for i, repo := range repos {
		if i > 0 {
			query += ","
		}

		query += fmt.Sprintf(`($%d, $%d, $%d, $%d, $%d, $%d) `, i*6+1, i*6+2, i*6+3, i*6+4, i*6+5, i*6+6)
		values = append(values, repo.Service, repo.Owner, repo.Name, repo.RepoServiceID, repo.WebhookID, repo.ServiceUserID)
	}
	query += `` +
		`ON CONFLICT (service, repo_service_id) DO UPDATE ` +
		`SET name = EXCLUDED.name, owner = EXCLUDED.owner, webhook_id = EXCLUDED.webhook_id`

	_, err := s.conn.Exec(ctx, query, values...)
	if err != nil {
		return err
	}

	return nil
}

//func (s *PostgresStore) GetPipeline(ctx context.Context, pipelineID int64) (*models.Pipeline, error) {
//	pipeline := &models.Pipeline{}
//	err := s.conn.QueryRowContext(ctx, ""+
//		"SELECT id, commit_sha, clone_url, status, url, started_at, finished_at, repo_id "+
//		"FROM pipeline "+
//		"WHERE id = $1",
//		pipelineID).
//		Scan(&pipeline.ID, &pipeline.CommitSHA, &pipeline.CloneURL, &pipeline.Status, &pipeline.URL, &pipeline.StartedAt, &pipeline.FinishedAt, &pipeline.RepoID)
//	if err != nil {
//		return nil, err
//	}
//
//	return pipeline, nil
//}

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
			RefreshToken: *ValueText(res.RefreshToken),
			TokenType:    res.TokenType,
			Expiry:       *ValueTime(res.TokenExpire),
		},
	}, nil
}

func (s *PostgresStore) CreatePipeline(ctx context.Context, pipeline *models.Pipeline) (int64, error) {
	pipelineID, err := s.queries.CreatePipeline(ctx, db.CreatePipelineParams{
		Status:    pipeline.Status,
		Context:   pipeline.Context,
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

func (s *PostgresStore) UpdatePipelineStatus(
	ctx context.Context, pipelineID int64, status string,
	started_at *time.Time, finished_at *time.Time,
) error {
	var t time.Time
	var set string
	if started_at != nil {
		t = *started_at
		set = `started_at`
	} else {
		t = *finished_at
		set = `finished_at`
	}

	_, err := s.conn.Exec(ctx, ``+
		`UPDATE public.pipeline `+
		`SET status = $1, `+set+` = $2 `+
		`WHERE id = $3`,
		status, t, pipelineID)
	return err
}

func (s *PostgresStore) GetPipelineStateChangeInfo(ctx context.Context, pipelineID int64,
) (*types.PipelineStateChangeInfo, error) {
	var (
		info         types.PipelineStateChangeInfo
		refreshToken sql.NullString
		tokenExpire  sql.NullTime
	)
	err := s.conn.QueryRow(ctx, ``+
		`SELECT p.url, p.commit_sha, p.context, p.started_at, r.service, r.owner,`+
		` r.name, su.access_token, su.refresh_token, su.token_type, su.token_expire `+
		`FROM (public.pipeline p JOIN public.repo r ON p.repo_id = r.id)`+
		` JOIN public.service_user su ON r.service_user_id = su.id `+
		`WHERE p.id = $1`,
		pipelineID).
		Scan(&info.URL, &info.CommitSHA, &info.Context, &info.StartedAt, &info.Service,
			&info.RepoOwner, &info.RepoName, &info.Token.AccessToken,
			&refreshToken, &info.Token.TokenType, &tokenExpire)
	if err != nil {
		return nil, err
	}

	if refreshToken.Valid {
		info.Token.RefreshToken = refreshToken.String
	}
	if tokenExpire.Valid {
		info.Token.Expiry = tokenExpire.Time
	}

	return &info, nil
}

func NullableInt8(ptr *int64) pgtype.Int8 {
	if ptr == nil {
		return pgtype.Int8{Valid: false}
	}
	return pgtype.Int8{Int64: *ptr, Valid: true}
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

func ValueInt8(value pgtype.Int8) *int64 {
	if !value.Valid {
		return nil
	}
	return &value.Int64
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
