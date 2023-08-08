package types

import (
	"time"

	"golang.org/x/oauth2"
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
