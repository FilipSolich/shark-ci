package models

import (
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

type Job struct {
	gorm.Model
	oauth2.Token
	CommitSHA string `json:"commitSHA"`
	CloneURL  string `json:"cloneURL"`
}
