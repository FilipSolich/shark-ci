package model2

import "time"

type Pipeline struct {
	ID         int64     `json:"id"`
	CommitSHA  string    `json:"commit_sha"`
	CloneURL   string    `json:"clone_url"`
	Status     string    `json:"status"`
	TargetURL  string    `json:"target_url"`
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at"`
	RepoID     int64     `json:"repo_id"`
}
