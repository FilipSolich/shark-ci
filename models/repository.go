package models

import (
	"github.com/FilipSolich/ci-server/db"
	"gorm.io/gorm"
)

type Repository struct {
	gorm.Model
	Webhook       Webhook
	UserID        uint
	ServiceName   string
	ServiceRepoID int64
	Name          string
	FullName      string
}

func CreateRepository(repo *Repository) (*Repository, error) {
	result := db.DB.Create(repo)
	return repo, result.Error
}
