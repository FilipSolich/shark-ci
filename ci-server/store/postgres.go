package store

import (
	"context"
	"database/sql"

	"github.com/FilipSolich/shark-ci/shared/model2"
	"github.com/google/uuid"

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

func (s *PostgresStore) GetUser(ctx context.Context, id int64) (*model2.User, error) {
	u := &model2.User{}
	err := s.db.QueryRowContext(ctx, "SELECT id, name, email FROM user WHERE id = $1", id).Scan(&u.ID, &u.Username, &u.Email)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (s *PostgresStore) GetServiceUserByUsername(ctx context.Context, service string, username string) (*model2.ServiceUser, error) {
	su := &model2.ServiceUser{}
	err := s.db.QueryRowContext(ctx, "SELECT id, service, username, email, access_token, refresh_token, token_type, token_expire, user_id FROM service_user WHERE username = $1 AND service = $2", username, service).Scan(&su.ID, &su.Service, &su.Username, &su.Email, &su.AccessToken, &su.RefreshToken, &su.TokenType, &su.TokenExpire, &su.UserID)
	if err != nil {
		return nil, err
	}

	return su, nil
}

func (s *PostgresStore) GetOAuth2State(ctx context.Context, state uuid.UUID) (*model2.OAuth2State, error) {
	oauth2State := &model2.OAuth2State{}
	err := s.db.QueryRowContext(ctx, "SELECT state, expire FROM oauth2_state WHERE state = $1", state).Scan(&oauth2State.State, &oauth2State.Expire)
	if err != nil {
		return nil, err
	}

	return oauth2State, nil
}

func (s *PostgresStore) CreateOAuth2State(ctx context.Context, state *model2.OAuth2State) error {
	_, err := s.db.ExecContext(ctx, "INSERT INTO oauth2_state (state, expire) VALUES ($1, $2)", state.State, state.Expire)
	return err
}
