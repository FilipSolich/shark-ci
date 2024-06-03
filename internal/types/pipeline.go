package types

import (
	"fmt"
	"time"

	"github.com/shark-ci/shark-ci/internal/config"
	"golang.org/x/oauth2"
)

type PipelineStatus string

const (
	Success PipelineStatus = "success" // GitHub -> Success, GitLab -> Success
	Pending PipelineStatus = "pending" // GitHub -> Pendign, GitLab -> Pending
	Running PipelineStatus = "running" // GitHub -> Pending, GitLab -> Running
	Error   PipelineStatus = "error"   // GitHub -> Error, GitLab -> Failed
)

type Pipeline struct {
	ID         int64
	URL        string
	Status     PipelineStatus
	CloneURL   string
	CommitSHA  string
	StartedAt  *time.Time
	FinishedAt *time.Time
	RepoID     int64
}

func (p *Pipeline) CreateURL() {
	p.URL = fmt.Sprintf("%s/repos/%d/pipelines/%d", config.ServerConf.Host, p.RepoID, p.ID)
}

type PipelineCreationInfo struct {
	RepoName string
	Username string
	Token    oauth2.Token
}

type PipelineStateChangeInfo struct {
	CommitSHA string
	URL       string
	Service   Service
	RepoOwner string
	RepoName  string
	Token     oauth2.Token
	StartedAt *time.Time
}

type PipelineLog struct {
	Order      int
	Cmd        string
	Output     string
	ExitCode   int
	PipelineID int64
}
