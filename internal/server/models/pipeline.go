package models

import (
	"fmt"
	"time"

	"github.com/shark-ci/shark-ci/internal/config"
)

type Pipeline struct {
	ID         int64      `json:"id"`
	URL        string     `json:"url"`
	Status     string     `json:"status"`
	CloneURL   string     `json:"clone_url"`
	CommitSHA  string     `json:"commit_sha"`
	StartedAt  *time.Time `json:"started_at,omitempty"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	RepoID     int64      `json:"repo_id"`
}

func (p *Pipeline) CreateURL() {
	p.URL = fmt.Sprintf("%s/repos/%d/pipelines/%d", config.ServerConf.Host, p.RepoID, p.ID)
}
