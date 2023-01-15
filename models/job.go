package models

import (
	"net"
	"net/url"

	"github.com/shark-ci/shark-ci/ci-server/configs"
	"golang.org/x/oauth2"
)

type Job struct {
	ID              string       `json:"_id,omitempty" bson:"_id,omitempty"`
	RepoID          string       `json:"-" bson:"repo,omitempty"`
	CommitSHA       string       `json:"commmitSHA,omitempty" bson:"commmitSHA,omitempty"`
	CloneURL        string       `json:"cloneURL,omitempty" bson:"cloneURL,omitempty"`
	Token           oauth2.Token `json:"token,omitempty" bson:"token,omitempty"`
	TargetURL       string       `json:"targetURL,omitempty" bson:"targetURL,omitempty"`
	ReportStatusURL string       `json:"reportStatusURL,omitempty" bson:"reportStatusURL,omitempty"`
	PublishLogsURL  string       `json:"publishLogsURL,omitempty" bson:"publishLogsURL,omitempty"`
}

func NewJob(repoID string, commitSHA string, cloneURL string, token oauth2.Token) *Job {
	return &Job{
		RepoID:    repoID,
		CommitSHA: commitSHA,
		CloneURL:  cloneURL,
		Token:     token,
	}
}

func (j *Job) createJobURLs() error {
	baseURL := url.URL{
		Scheme: "https",
		Host:   net.JoinHostPort(configs.Host, configs.Port),
	}

	var err error
	targetURL := baseURL
	targetURL.Path, err = url.JoinPath(configs.JobsPath, j.ID)
	if err != nil {
		return err
	}
	reportStatusURL := baseURL
	reportStatusURL.Path, err = url.JoinPath(configs.JobsReportStatusHandlerPath, j.ID)
	if err != nil {
		return err
	}
	publishLogsURL := baseURL
	publishLogsURL.Path, err = url.JoinPath(configs.JobsPublishLogsHandlerPath, j.ID)
	if err != nil {
		return err
	}

	j.TargetURL = targetURL.String()
	j.ReportStatusURL = reportStatusURL.String()
	j.PublishLogsURL = publishLogsURL.String()
	return nil
}
