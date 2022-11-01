package models

import (
	"fmt"
	"net"
	"net/url"

	"github.com/FilipSolich/ci-server/configs"
	"github.com/FilipSolich/ci-server/db"
	"gorm.io/gorm"
)

type Job struct {
	gorm.Model
	OAuth2TokenID   uint
	CommitSHA       string `json:"commit_sha"`
	CloneURL        string `json:"clone_url"`
	TargetURL       string
	ReportStatusURL string
	PublishLogsURL  string
}

func CreateJob(job *Job) (*Job, error) {
	result := db.DB.Create(job)
	return job, result.Error
}

func (*Job) TableName() string {
	return "job"
}

func (j *Job) CreateJobURLs() error {
	baseURL := url.URL{
		Scheme: "https",
		Host:   net.JoinHostPort(configs.Host, configs.Port),
	}
	var err error

	targetURL := baseURL
	targetURL.Path, err = url.JoinPath(configs.JobsPath, fmt.Sprint(j.ID))
	if err != nil {
		return err
	}
	reportStatusURL := baseURL
	reportStatusURL.Path, err = url.JoinPath(configs.JobsPath, configs.JobsReportStatusHandlerPath, fmt.Sprint(j.ID))
	if err != nil {
		return err
	}
	publishLogsURL := baseURL
	publishLogsURL.Path, err = url.JoinPath(configs.JobsPath, configs.JobsPublishLogsHandlerPath, fmt.Sprint(j.ID))
	if err != nil {
		return err
	}

	j.TargetURL = targetURL.String()
	j.ReportStatusURL = reportStatusURL.String()
	j.PublishLogsURL = publishLogsURL.String()
	return db.DB.Save(j).Error
}
