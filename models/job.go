package models

import (
	"github.com/FilipSolich/ci-server/db"
	"gorm.io/gorm"
)

type Job struct {
	gorm.Model
	OAuth2TokenID   uint
	CommitSHA       string `json:"commit_sha"`
	CloneURL        string `json:"clone_url"`
	ReportStatusURL string
	PublishLogsURL  string
}

func CreateJob(job *Job) (*Job, error) {
	result := db.DB.Create(job)
	return job, result.Error
}
