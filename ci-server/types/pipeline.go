package types

import (
	"time"

	"golang.org/x/oauth2"
)

type PipilineStateChangeInfo struct {
	CommitSHA string
	URL       string
	Service   string
	RepoOwner string
	RepoName  string
	Token     *oauth2.Token
	StartedAt *time.Time
}
