package models

import "gorm.io/gorm"

type Webhook struct {
	gorm.Model
	RepoFullName string
	RepoID       int
	UserID       uint
}
