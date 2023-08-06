package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/FilipSolich/shark-ci/ci-server/models"
	"github.com/FilipSolich/shark-ci/ci-server/types"
	"github.com/google/uuid"
	"golang.org/x/oauth2"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresStore struct {
	db *sql.DB
}

var _ Storer = &PostgresStore{}

func NewPostgresStore(postgresURI string) (*PostgresStore, error) {
	db, err := sql.Open("pgx", postgresURI)
	if err != nil {
		return nil, err
	}

	return &PostgresStore{
		db: db,
	}, nil
}

func (s *PostgresStore) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *PostgresStore) Close(_ context.Context) error {
	return s.db.Close()
}

func (s *PostgresStore) Clean(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM oauth2_state WHERE expire < NOW()")
	return err
}

func (s *PostgresStore) GetOAuth2State(ctx context.Context, state uuid.UUID) (*models.OAuth2State, error) {
	oauth2State := &models.OAuth2State{}
	err := s.db.QueryRowContext(ctx, "SELECT state, expire FROM oauth2_state WHERE state = $1",
		state.String()).
		Scan(&oauth2State.State, &oauth2State.Expire)
	if err != nil {
		return nil, err
	}

	return oauth2State, nil
}

func (s *PostgresStore) CreateOAuth2State(ctx context.Context, state *models.OAuth2State) error {
	_, err := s.db.ExecContext(ctx,
		"INSERT INTO oauth2_state (state, expire) VALUES ($1, $2)",
		state.State.String(), state.Expire)
	return err
}

func (s *PostgresStore) DeleteOAuth2State(ctx context.Context, state *models.OAuth2State) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM oauth2_state WHERE state = $1", state.State.String())
	return err
}

func (s *PostgresStore) GetUser(ctx context.Context, id int64) (*models.User, error) {
	u := &models.User{}
	err := s.db.QueryRowContext(ctx, `SELECT id, username, email FROM "user" WHERE id = $1`, id).
		Scan(&u.ID, &u.Username, &u.Email)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (s *PostgresStore) CreateUserAndServiceUser(ctx context.Context, serviceUser *models.ServiceUser) (int64, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	user := models.User{
		Username: serviceUser.Username,
		Email:    serviceUser.Email,
	}
	var userID int64
	err = tx.QueryRowContext(ctx, `INSERT INTO "user" (username, email) VALUES ($1, $2) RETURNING id`,
		user.Username, user.Email).Scan(&userID)
	if err != nil {
		return 0, err
	}

	serviceUser.UserID = userID
	_, err = tx.ExecContext(ctx, ""+
		"INSERT INTO service_user (service, username, email, access_token, refresh_token, token_type, token_expire, user_id) "+
		"VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		serviceUser.Service, serviceUser.Username, serviceUser.Email, serviceUser.AccessToken,
		serviceUser.RefreshToken, serviceUser.TokenType, serviceUser.TokenExpire, serviceUser.UserID)
	if err != nil {
		return 0, err
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func (s *PostgresStore) GetServiceUserByUniqueName(ctx context.Context, service string, username string) (*models.ServiceUser, error) {
	su := &models.ServiceUser{}
	err := s.db.QueryRowContext(ctx, ""+
		"SELECT id, service, username, email, access_token, refresh_token, token_type, token_expire, user_id "+
		"FROM service_user "+
		"WHERE username = $1 AND service = $2",
		username, service).
		Scan(&su.ID, &su.Service, &su.Username, &su.Email, &su.AccessToken,
			&su.RefreshToken, &su.TokenType, &su.TokenExpire, &su.UserID)
	if err != nil {
		return nil, err
	}

	return su, nil
}

func (s *PostgresStore) GetServiceUserByRepo(ctx context.Context, repoID int64) (*models.ServiceUser, error) {
	su := &models.ServiceUser{}
	err := s.db.QueryRowContext(ctx, ""+
		"SELECT id, service, username, email, access_token, refresh_token, token_type, token_expire, user_id "+
		"FROM service_user "+
		"WHERE id = (SELECT service_user_id FROM repo WHERE id = $1)",
		repoID).
		Scan(&su.ID, &su.Service, &su.Username, &su.Email, &su.AccessToken, &su.RefreshToken,
			&su.TokenType, &su.TokenExpire, &su.UserID)
	if err != nil {
		return nil, err
	}

	return su, nil
}

func (s *PostgresStore) GetServiceUserByUserAndService(ctx context.Context, userID int64, service string) (*models.ServiceUser, error) {
	su := &models.ServiceUser{}
	err := s.db.QueryRowContext(ctx, ""+
		"SELECT id, service, username, email, access_token, refresh_token, token_type, token_expire, user_id "+
		"FROM service_user "+
		"WHERE user_id = $1 AND service = $2",
		userID, service).
		Scan(&su.ID, &su.Service, &su.Username, &su.Email, &su.AccessToken,
			&su.RefreshToken, &su.TokenType, &su.TokenExpire, &su.UserID)
	if err != nil {
		return nil, err
	}

	return su, nil
}

func (s *PostgresStore) GetServiceUsersByUser(ctx context.Context, userID int64) ([]models.ServiceUser, error) {
	rows, err := s.db.QueryContext(ctx, ""+
		"SELECT id, service, username, email, access_token, refresh_token, token_type, token_expire, user_id "+
		"FROM service_user "+
		"WHERE user_id = $1",
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var serviceUsers []models.ServiceUser
	for rows.Next() {
		su := models.ServiceUser{}
		err := rows.Scan(&su.ID, &su.Service, &su.Username, &su.Email, &su.AccessToken,
			&su.RefreshToken, &su.TokenType, &su.TokenExpire, &su.UserID)
		if err != nil {
			return serviceUsers, err
		}
		serviceUsers = append(serviceUsers, su)
	}

	err = rows.Err()
	if err != nil {
		return serviceUsers, err
	}

	return serviceUsers, nil
}

func (s *PostgresStore) UpdateServiceUserToken(ctx context.Context, serviceUser *models.ServiceUser, token *oauth2.Token) error {
	_, err := s.db.ExecContext(ctx, ""+
		"UPDATE service_user "+
		"SET access_token = $1, refresh_token = $2, token_type = $3, token_expire = $4 "+
		"WHERE id = $5",
		token.AccessToken, token.RefreshToken, token.TokenType, token.Expiry, serviceUser.ID)
	return err
}

func (s *PostgresStore) GetRepo(ctx context.Context, repoID int64) (*models.Repo, error) {
	repo := &models.Repo{}
	err := s.db.QueryRowContext(ctx, ""+
		"SELECT id, service, repo_service_id, name, webhook_id, service_user_id "+
		"FROM repo "+
		"WHERE id = $1",
		repoID).
		Scan(&repo.ID, &repo.Service, &repo.RepoServiceID, &repo.Name, &repo.WebhookID, &repo.ServiceUserID)
	if err != nil {
		return nil, err
	}

	return repo, nil
}

func (s *PostgresStore) GetRepoName(ctx context.Context, repoID int64) (string, error) {
	var repoName string
	err := s.db.QueryRowContext(ctx, "SELECT name FROM repo WHERE id = $1", repoID).Scan(&repoName)
	if err != nil {
		return "", err
	}

	return repoName, nil
}

func (s *PostgresStore) GetRepoIDByServiceRepoID(ctx context.Context, service string, serviceRepoID int64) (int64, error) {
	var repoID int64
	err := s.db.QueryRowContext(ctx, "SELECT id FROM repo WHERE service = $1 AND repo_service_id = $2",
		service, serviceRepoID).
		Scan(&repoID)
	if err != nil {
		return 0, err
	}

	return repoID, nil
}

func (s *PostgresStore) GetReposByUser(ctx context.Context, userID int64) ([]models.Repo, error) {
	rows, err := s.db.QueryContext(ctx, ""+
		"SELECT id, service, repo_service_id, name, webhook_id, service_user_id "+
		"FROM repo "+
		"WHERE service_user_id in (SELECT id FROM service_user WHERE user_id = $1)",
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var repos []models.Repo
	for rows.Next() {
		repo := models.Repo{}
		err := rows.Scan(&repo.ID, &repo.Service, &repo.RepoServiceID, &repo.Name, &repo.WebhookID, &repo.ServiceUserID)
		if err != nil {
			return repos, err
		}
		repos = append(repos, repo)
	}

	err = rows.Err()
	if err != nil {
		return repos, err
	}

	return repos, nil
}

func (s *PostgresStore) CreateOrUpdateRepos(ctx context.Context, repos []models.Repo) error {
	query := "INSERT INTO repo (service, repo_service_id, name, service_user_id) VALUES"
	values := []any{}
	for i, repo := range repos {
		if i > 0 {
			query += ","
		}

		query += fmt.Sprintf("($%d, $%d, $%d, $%d) ", i*4+1, i*4+2, i*4+3, i*4+4)
		values = append(values, repo.Service, repo.RepoServiceID, repo.Name, repo.ServiceUserID)
	}
	query += "ON CONFLICT (service, repo_service_id) DO UPDATE SET name = EXCLUDED.name"

	_, err := s.db.ExecContext(ctx, query, values...)
	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresStore) UpdateRepoWebhook(ctx context.Context, repoID int64, webhookID *int64) error {
	_, err := s.db.ExecContext(ctx, "UPDATE repo SET webhook_id = $1 WHERE id = $2", webhookID, repoID)
	return err
}

func (s *PostgresStore) GetPipeline(ctx context.Context, pipelineID int64) (*models.Pipeline, error) {
	pipeline := &models.Pipeline{}
	err := s.db.QueryRowContext(ctx, ""+
		"SELECT id, commit_sha, clone_url, status, target_url, started_at, finished_at, repo_id "+
		"FROM pipeline "+
		"WHERE id = $1",
		pipelineID).
		Scan(&pipeline.ID, &pipeline.CommitSHA, &pipeline.CloneURL, &pipeline.Status, &pipeline.TargetURL, &pipeline.StartedAt, &pipeline.FinishedAt, &pipeline.RepoID)
	if err != nil {
		return nil, err
	}

	return pipeline, nil
}

func (s *PostgresStore) CreatePipeline(ctx context.Context, pipeline *models.Pipeline) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = tx.QueryRowContext(ctx, ""+
		"INSERT INTO pipeline (commit_sha, clone_url, status, repo_id) "+
		"VALUES ($1, $2, $3, $4) "+
		"RETURNING id",
		pipeline.CommitSHA, pipeline.CloneURL, pipeline.Status, pipeline.RepoID).Scan(&pipeline.ID)
	if err != nil {
		return err
	}

	pipeline.CreateTargetURL()
	_, err = tx.ExecContext(ctx, "UPDATE pipeline SET target_url = $1 WHERE id = $2", pipeline.TargetURL, pipeline.ID)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
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

	_, err := s.db.ExecContext(ctx, ``+
		`UPDATE public.pipeline `+
		`SET status = $1, `+set+` = $2 `+
		`WHERE id = $3`,
		status, t, pipelineID)
	return err
}

func (s *PostgresStore) GetPipelineStateChangeInfo(
	ctx context.Context, pipelineID int64,
) (*types.PipilineStateChangeInfo, error) {
	info := &types.PipilineStateChangeInfo{Token: &oauth2.Token{}}
	var (
		refreshToken sql.NullString
		tokenExpire  sql.NullTime
		startedAt    sql.NullTime
	)
	err := s.db.QueryRowContext(ctx, ``+
		`SELECT p.commit_sha, p.target_url, p.started_at, r.service, r.owner,`+
		` r.name, su.access_token, su.refresh_token, su.token_type, su.token_expire `+
		`FROM (public.pipeline p JOIN public.repo r ON p.repo_id = r.id)`+
		` JOIN public.service_user su ON r.service_user_id = su.id `+
		`WHERE p.id = $1`,
		pipelineID).
		Scan(&info.CommitSHA, &info.TargetURL, &startedAt, &info.Service,
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
	if startedAt.Valid {
		info.StartedAt = &startedAt.Time
	}

	return info, nil
}
