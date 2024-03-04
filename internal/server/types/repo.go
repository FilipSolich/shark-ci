package types

import "golang.org/x/oauth2"

type Repo struct {
	ID            int64
	Service       string
	Owner         string
	Name          string
	RepoServiceID int64
	WebhookID     *int64
	ServiceUserID int64
}

type RepoWebhookChangeInfo struct {
	RepoID    int64
	Service   string
	RepoOwner string
	RepoName  string
	WebhookID *int64
	Token     oauth2.Token
	UserID    int64
}
