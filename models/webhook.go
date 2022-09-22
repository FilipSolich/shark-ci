package models

import "gorm.io/gorm"

type Webhook struct {
	gorm.Model
	Service      string
	RepoFullName string
	RepoID       int64
}
