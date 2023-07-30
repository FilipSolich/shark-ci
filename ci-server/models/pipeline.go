package models

import (
	"fmt"
	"time"
)

type Pipeline struct {
	ID         int64      `json:"id"`
	CommitSHA  string     `json:"commit_sha"`
	CloneURL   string     `json:"clone_url"`
	Status     string     `json:"status"`
	TargetURL  string     `json:"target_url"`
	StartedAt  *time.Time `json:"started_at,omitempty"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	RepoID     int64      `json:"repo_id"`
}

func (p *Pipeline) CreateTargetURL(url string) {
	p.TargetURL = fmt.Sprintf("%s/%d", url, p.ID)
}
