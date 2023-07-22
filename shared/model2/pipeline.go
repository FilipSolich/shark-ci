package model2

import "time"

type Pipeline struct {
	ID           int64     `json:"id"`
	CommitSHA    string    `json:"commit_sha"`
	CloneURL     string    `json:"clone_url"`
	Status       string    `json:"status"`
	StartedAt    time.Time `json:"started_at"`
	FinishedAt   time.Time `json:"finished_at"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	TokenExpire  time.Time `json:"token_expire"`
	RepoID       int64     `json:"repo_id"`
}
