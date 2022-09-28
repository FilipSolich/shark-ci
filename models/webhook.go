package models

import (
	"github.com/FilipSolich/ci-server/db"
	"gorm.io/gorm"
)

type Webhook struct {
	gorm.Model
	RepositoryID     uint
	ServiceName      string
	ServiceWebhookID int64
	Active           bool
}

func CreateWebhook(hook *Webhook) (*Webhook, error) {
	result := db.DB.Create(hook)
	return hook, result.Error
}
