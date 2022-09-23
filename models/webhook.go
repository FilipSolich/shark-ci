package models

import "gorm.io/gorm"

type Webhook struct {
	gorm.Model
	ServiceWebhookID int64
	Service          string
	RepoID           int64
	RepoName         string
	RepoFullName     string
	Active           bool
}
