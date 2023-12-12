package types

import "golang.org/x/oauth2"

type RepoWebhookChangeInfo struct {
	RepoID    int64
	Service   string
	RepoOwner string
	RepoName  string
	WebhookID *int64
	Token     oauth2.Token
	UserID    int64
}
