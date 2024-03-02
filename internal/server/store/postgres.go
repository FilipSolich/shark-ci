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
	expire, err := s.queries.GetOAuth2StateExpiration(ctx, state)
	if err != nil {
		return types.OAuth2State{}, err
	}

	return types.OAuth2State{State: state, Expire: expire.Time}, s.queries.DeleteOAuth2State(ctx, state)
}

func (s *PostgresStore) CreateOAuth2State(ctx context.Context, state types.OAuth2State) error {
	return s.queries.CreateOAuth2State(ctx, db.CreateOAuth2StateParams{
		State:  state.State,
		Expire: pgtype.Timestamp{Time: state.Expire, Valid: true},
	})
}

func (s *PostgresStore) GetUser(ctx context.Context, userID int64) (*models.User, error) {
	u := &models.User{}
	err := s.conn.QueryRow(ctx, `SELECT id, username, email FROM public.user WHERE id = $1`, userID).
		Scan(&u.ID, &u.Username, &u.Email)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (s *PostgresStore) CreateUserAndServiceUser(ctx context.Context, serviceUser *models.ServiceUser) (int64, int64, error) {
	tx, err := s.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback(ctx)

	user := models.User{
		Username: serviceUser.Username,
		Email:    serviceUser.Email,
	}
	var userID int64
	err = tx.QueryRow(ctx, ``+
		`INSERT INTO public.user (username, email) VALUES ($1, $2) RETURNING id`,
		user.Username, user.Email).Scan(&userID)
	if err != nil {
		return 0, 0, err
	}

	serviceUser.UserID = userID

	var serviceUserID int64
	err = tx.QueryRow(ctx, ``+
		`INSERT INTO public.service_user`+
		` (service, username, email, access_token, refresh_token, token_type, token_expire, user_id) `+
		`VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
		serviceUser.Service, serviceUser.Username, serviceUser.Email, serviceUser.AccessToken,
		serviceUser.RefreshToken, serviceUser.TokenType, serviceUser.TokenExpire, serviceUser.UserID).
		Scan(&serviceUserID)
	if err != nil {
		return 0, 0, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return 0, 0, err
	}

	return userID, serviceUserID, nil
}

func (s *PostgresStore) GetServiceUserIDsByServiceUsername(ctx context.Context, service string, username string) (int64, int64, error) {
	res, err := s.queries.GetServiceUserIDsByServiceUsername(ctx, db.GetServiceUserIDsByServiceUsernameParams{
		Service:  db.Service(service),
		Username: username,
	})
	if err != nil {
		return 0, 0, err
	}

	return res.ID, res.UserID, nil
}

func (s *PostgresStore) GetServiceUsersRepoFetchInfo(ctx context.Context, userID int64) ([]*types.ServiceUserRepoFetchInfo, error) {
	rows, err := s.conn.Query(ctx, ``+
		`SELECT id, service, access_token, refresh_token, token_type, token_expire `+
		`FROM public.service_user `+
		`WHERE user_id = $1`,
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	serviceUsersInfo := []*types.ServiceUserRepoFetchInfo{}
	for rows.Next() {
		var (
			info         types.ServiceUserRepoFetchInfo
			refreshToken sql.NullString
			tokenExpire  sql.NullTime
		)
		err := rows.Scan(&info.ID, &info.Service, &info.Token.AccessToken, &refreshToken,
			&info.Token.TokenType, &tokenExpire)
		if err != nil {
			return serviceUsersInfo, err
		}
		if refreshToken.Valid {
			info.Token.RefreshToken = refreshToken.String
		}
		if tokenExpire.Valid {
			info.Token.Expiry = tokenExpire.Time
		}

		serviceUsersInfo = append(serviceUsersInfo, &info)
	}

	err = rows.Err()
	if err != nil {
		return serviceUsersInfo, err
	}

	return serviceUsersInfo, nil
}

func (s *PostgresStore) UpdateServiceUserToken(ctx context.Context, serviceUserID int64, token *oauth2.Token) error {
	_, err := s.conn.Exec(ctx, ``+
		`UPDATE public.service_user `+
		`SET access_token = $1, refresh_token = $2, token_type = $3, token_expire = $4 `+
		`WHERE id = $5`,
		token.AccessToken, token.RefreshToken, token.TokenType, token.Expiry, serviceUserID)
	return err
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

func (s *PostgresStore) GetReposByUser(ctx context.Context, userID int64) ([]db.Repo, error) {
	return s.queries.GetReposByUser(ctx, userID)
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

func (s *PostgresStore) GetRegisterWebhookInfoByRepo(ctx context.Context, repoID int64) (db.GetRegisterWebhookInfoByRepoRow, error) {
	return s.queries.GetRegisterWebhookInfoByRepo(ctx, repoID)
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

func (s *PostgresStore) UpdateRepoWebhook(ctx context.Context, repoID int64, webhookID *int64) error {
	webhook := pgtype.Int8{Valid: false}
	if webhookID != nil {
		webhook = pgtype.Int8{Int64: *webhookID, Valid: true}
	}
	return s.queries.SetRepoWebhook(ctx, db.SetRepoWebhookParams{
		ID:        repoID,
		WebhookID: webhook,
	})
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
	var (
		info         types.PipelineCreationInfo
		refreshToken sql.NullString
		tokenExpire  sql.NullTime
	)
	err := s.conn.QueryRow(ctx, ``+
		`SELECT su.username, su.access_token, su.refresh_token, su.token_type, su.token_expire, r.name `+
		`FROM public.service_user su JOIN repo r ON su.id = r.service_user_id `+
		`WHERE r.id = $1`,
		repoID).
		Scan(&info.Username, &info.Token.AccessToken, &refreshToken, &info.Token.TokenType, &tokenExpire)
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

func (s *PostgresStore) CreatePipeline(ctx context.Context, pipeline *models.Pipeline) (int64, error) {
	tx, err := s.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, ``+
		`INSERT INTO public.pipeline (status, context, clone_url, commit_sha, repo_id) `+
		`VALUES ($1, $2, $3, $4) `+
		`RETURNING id`,
		pipeline.Status, pipeline.Context, pipeline.CloneURL, pipeline.CommitSHA, pipeline.RepoID).
		Scan(&pipeline.ID)
	if err != nil {
		return 0, err
	}

	pipeline.CreateURL()
	_, err = tx.Exec(ctx, `UPDATE public.pipeline SET url = $1 WHERE id = $2`,
		pipeline.URL, pipeline.ID)
	if err != nil {
		return 0, err
	}

	err = tx.Commit(ctx)
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
