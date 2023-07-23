package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/FilipSolich/shark-ci/shared/model2"
	"github.com/google/uuid"
	"golang.org/x/oauth2"

	_ "github.com/lib/pq"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(postgresURI string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", postgresURI)
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

func (s *PostgresStore) GetOAuth2State(ctx context.Context, state uuid.UUID) (*model2.OAuth2State, error) {
	oauth2State := &model2.OAuth2State{}
	err := s.db.QueryRowContext(ctx, "SELECT state, expire FROM oauth2_state WHERE state = $1",
		state.String()).
		Scan(&oauth2State.State, &oauth2State.Expire)
	if err != nil {
		return nil, err
	}

	return oauth2State, nil
}

func (s *PostgresStore) CreateOAuth2State(ctx context.Context, state *model2.OAuth2State) error {
	_, err := s.db.ExecContext(ctx,
		"INSERT INTO oauth2_state (state, expire) VALUES ($1, $2)",
		state.State.String(), state.Expire)
	return err
}

func (s *PostgresStore) DeleteOAuth2State(ctx context.Context, state *model2.OAuth2State) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM oauth2_state WHERE state = $1", state.State.String())
	return err
}

func (s *PostgresStore) GetUser(ctx context.Context, id int64) (*model2.User, error) {
	u := &model2.User{}
	err := s.db.QueryRowContext(ctx, "SELECT id, name, email FROM user WHERE id = $1", id).
		Scan(&u.ID, &u.Username, &u.Email)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (s *PostgresStore) CreateUserAndServiceUser(ctx context.Context, serviceUser *model2.ServiceUser) (int64, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	user := model2.User{
		Username: serviceUser.Username,
		Email:    serviceUser.Email,
	}
	res, err := tx.ExecContext(ctx, "INSERT INTO user (name, email) VALUES ($1, $2)",
		user.Username, user.Email)
	if err != nil {
		return 0, err
	}

	userID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	serviceUser.UserID = userID
	_, err = tx.ExecContext(ctx, ""+
		"INSERT INTO service_user (service, username, email, access_token, refresh_token, token_type, token_expire, user_id)"+
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

func (s *PostgresStore) GetServiceUserByUniqueName(ctx context.Context, service string, username string) (*model2.ServiceUser, error) {
	su := &model2.ServiceUser{}
	err := s.db.QueryRowContext(ctx, ""+
		"SELECT id, service, username, email, access_token, refresh_token, token_type, token_expire, user_id"+
		"FROM service_user"+
		"WHERE username = $1 AND service = $2",
		username, service).
		Scan(&su.ID, &su.Service, &su.Username, &su.Email, &su.AccessToken,
			&su.RefreshToken, &su.TokenType, &su.TokenExpire, &su.UserID)
	if err != nil {
		return nil, err
	}

	return su, nil
}

func (s *PostgresStore) GetServiceUserByRepo(ctx context.Context, repoID int64) (*model2.ServiceUser, error) {
	su := &model2.ServiceUser{}
	err := s.db.QueryRowContext(ctx, ""+
		"SELECT id, service, username, email, access_token, refresh_token, token_type, token_expire, user_id"+
		"FROM service_user"+
		"WHERE id = (SELECT service_user_id FROM repo WHERE id = $1)",
		repoID).
		Scan(&su.ID, &su.Service, &su.Username, &su.Email, &su.AccessToken, &su.RefreshToken,
			&su.TokenType, &su.TokenExpire, &su.UserID)
	if err != nil {
		return nil, err
	}

	return su, nil
}

func (s *PostgresStore) GetServiceUsersByUser(ctx context.Context, userID int64) ([]model2.ServiceUser, error) {
	rows, err := s.db.QueryContext(ctx, ""+
		"SELECT id, service, username, email, access_token, refresh_token, token_type, token_expire, user_id"+
		"FROM service_user"+
		"WHERE user_id = $1",
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var serviceUsers []model2.ServiceUser
	for rows.Next() {
		su := model2.ServiceUser{}
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

func (s *PostgresStore) UpdateServiceUserToken(ctx context.Context, serviceUser *model2.ServiceUser, token *oauth2.Token) error {
	_, err := s.db.ExecContext(ctx, ""+
		"UPDATE service_user"+
		"SET access_token = $1, refresh_token = $2, token_type = $3, token_expire = $4"+
		"WHERE id = $5",
		token.AccessToken, token.RefreshToken, token.TokenType, token.Expiry, serviceUser.ID)
	return err
}

func (s *PostgresStore) GetRepo(ctx context.Context, repoID int64) (*model2.Repo, error) {
	repo := &model2.Repo{}
	err := s.db.QueryRowContext(ctx, ""+
		"SELECT id, service, repo_service_id, name, webhook_id, service_user_id"+
		"FROM repo"+
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
	err := s.db.QueryRowContext(ctx, "SELECT id FROM repo WHERE service = $1 AND service_repo_id = $2",
		service, serviceRepoID).
		Scan(&repoID)
	if err != nil {
		return 0, err
	}

	return repoID, nil
}

func (s *PostgresStore) GetReposByUser(ctx context.Context, userID int64) ([]model2.Repo, error) {
	rows, err := s.db.QueryContext(ctx, ""+
		"SELECT id, service, repo_service_id, name, webhook_id, service_user_id"+
		"FROM repo"+
		"WHERE service_user_id in (SELECT id FROM service_user WHERE user_id = $1)",
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var repos []model2.Repo
	for rows.Next() {
		repo := model2.Repo{}
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

func (s *PostgresStore) CreateOrUpdateRepos(ctx context.Context, repos []model2.Repo) error {
	query := "INSERT INTO (service, repo_service_id, name, service_user_id) VALUES"
	values := []interface{}{}
	for i, repo := range repos {
		if i > 1 {
			query += ","
		}

		query += fmt.Sprintf(" ($%d, $%d, $%d, $%d)", i*4+1, i*4+2, i*4+3, i*4+4)
		values = append(values, repo.Service, repo.RepoServiceID, repo.Name, repo.ServiceUserID)
	}
	query += " ON CONFLICT (service, repo_service_id) DO UPDATE SET name = EXCLUDED.name"

	_, err := s.db.ExecContext(ctx, query, values...)
	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresStore) UpdateRepoWebhook(ctx context.Context, repoID int64, webhookID int64) error {
	var value any = webhookID
	if webhookID == 0 {
		value = nil
	}

	_, err := s.db.ExecContext(ctx, "UPDATE repo SET webhook_id = $1 WHERE id = $2", value, repoID)
	return err
}

func (s *PostgresStore) CreatePipeline(ctx context.Context, pipeline *model2.Pipeline) error {
	_, err := s.db.ExecContext(ctx, ""+
		"INSERT INTO pipeline (commit_sha, clone_url, status, repo_id)"+
		"VALUES ($1, $2, $3, $4)",
		pipeline.CommitSHA, pipeline.CloneURL, pipeline.Status, pipeline.RepoID)
	return err
}
