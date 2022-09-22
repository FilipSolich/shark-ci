package models

import "gorm.io/gorm"

type Webhook struct {
	gorm.Model
	Service      string
	RepoID       int64
	RepoName     string
	RepoFullName string
}
