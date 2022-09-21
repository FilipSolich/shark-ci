package models

import (
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

type OAuth2Token struct {
	gorm.Model
	oauth2.Token
	UserID uint
}
