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

func (j *Job) CreateTargetURL() error {
	URL, err := createJobURL(fmt.Sprint(j.ID), "")
	if err != nil {
		return err
	}

	j.TargetURL = URL.String()
	return db.DB.Save(j).Error
}

func (j *Job) CreateReportStatusURL() error {
	URL, err := createJobURL(fmt.Sprint(j.ID), configs.JobsReportStatusHandlerPath)
	if err != nil {
		return err
	}

	j.ReportStatusURL = URL.String()
	return db.DB.Save(j).Error
}

func (j *Job) CreatePublishLogsURL() error {
	URL, err := createJobURL(fmt.Sprint(j.ID), configs.JobsPublishLogsHandlerPath)
	if err != nil {
		return err
	}

	j.PublishLogsURL = URL.String()
	return db.DB.Save(j).Error
}

func createJobURL(jobID string, endpoint string) (url.URL, error) {
	path, err := url.JoinPath(configs.JobsPath, endpoint, jobID)
	if err != nil {
		return url.URL{}, err
	}

	URL := url.URL{
		Scheme: "https",
		Host:   net.JoinHostPort(configs.Host, configs.Port),
		Path:   path,
	}
	return URL, nil
}
