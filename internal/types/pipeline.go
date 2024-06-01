package types

import (
	"time"

	"golang.org/x/oauth2"
)

type PipelineStatus string

const (
	Success PipelineStatus = "success" // GitHub -> Success, GitLab -> Success
	Pending PipelineStatus = "pending" // GitHub -> Pendign, GitLab -> Pending
	Running PipelineStatus = "running" // GitHub -> Pending, GitLab -> Running
	Error   PipelineStatus = "error"   // GitHub -> Error, GitLab -> Failed
)

type PipelineCreationInfo struct {
	RepoName string
	Username string
	Token    oauth2.Token
}

type PipelineStateChangeInfo struct {
	CommitSHA string
	URL       string
	Context   string
	Service   string
	RepoOwner string
	RepoName  string
	Token     oauth2.Token
	StartedAt *time.Time
}
