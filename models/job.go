package models

import (
	"fmt"
	"net"

	"github.com/shark-ci/shark-ci/ci-server/configs"
	"golang.org/x/oauth2"
)

type Job struct {
	ID        string       `json:"_id,omitempty" bson:"_id,omitempty"`
	RepoID    string       `json:"-" bson:"repo,omitempty"`
	CommitSHA string       `json:"commmitSHA,omitempty" bson:"commmitSHA,omitempty"`
	CloneURL  string       `json:"cloneURL,omitempty" bson:"cloneURL,omitempty"`
	Token     oauth2.Token `json:"token,omitempty" bson:"token,omitempty"`
	TargetURL string       `json:"targetURL,omitempty" bson:"targetURL,omitempty"`

	Ack  func() error
	Nack func() error
}

func NewJob(repoID string, commitSHA string, cloneURL string, token oauth2.Token) *Job {
	return &Job{
		RepoID:    repoID,
		CommitSHA: commitSHA,
		CloneURL:  cloneURL,
		Token:     token,
	}
}

func (j *Job) createJobURL() {
	var host string
	if configs.Port == "443" {
		host = configs.Host
	} else {
		host = net.JoinHostPort(configs.Host, configs.Port)
	}

	j.TargetURL = fmt.Sprintf("https://%s/%s/%s", host, configs.JobsPath, j.ID)
}
