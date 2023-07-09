package model

import (
	"golang.org/x/oauth2"
)

type Job struct {
	ID         string       `json:"_id,omitempty" bson:"_id,omitempty"`
	RepoID     string       `json:"-" bson:"repo,omitempty"`
	UniqueName string       `json:"uniqueName,omitempty" bson:"uniqueName,omitempty"`
	CommitSHA  string       `json:"commmitSHA,omitempty" bson:"commmitSHA,omitempty"`
	CloneURL   string       `json:"cloneURL,omitempty" bson:"cloneURL,omitempty"`
	Token      oauth2.Token `json:"token,omitempty" bson:"token,omitempty"`
	TargetURL  string       `json:"targetURL,omitempty" bson:"targetURL,omitempty"`

	Ack  func() error `json:"-" bson:"-"`
	Nack func() error `json:"-" bson:"-"`
}

func NewJob(repoID string, uniqueName string, commitSHA string, cloneURL string, token oauth2.Token) *Job {
	return &Job{
		RepoID:     repoID,
		UniqueName: uniqueName,
		CommitSHA:  commitSHA,
		CloneURL:   cloneURL,
		Token:      token,
	}
}
