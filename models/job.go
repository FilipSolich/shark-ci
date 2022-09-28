package models

import (
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

type Job struct {
	gorm.Model
	oauth2.Token
	CommitSHA string `json:"commit_sha"`
	CloneURL  string `json:"clone_url"`
}
