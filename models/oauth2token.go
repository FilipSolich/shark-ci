package models

import (
	"errors"

	"github.com/FilipSolich/ci-server/db"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

type OAuth2Token struct {
	gorm.Model
	oauth2.Token
	UserIdentityID uint
	Jobs           []Job `json:"-"`
}

func CreateOAuth2Token(token *OAuth2Token) (*OAuth2Token, error) {
	result := db.DB.Create(token)
	return token, result.Error
}

func GetOrCreateOAuth2Token(token *OAuth2Token) (*OAuth2Token, error) {
	var getToken OAuth2Token
	var err error
	result := db.DB.First(&getToken, token)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, result.Error
		}

		return CreateOAuth2Token(token)
	}

	return &getToken, err
}

func (*OAuth2Token) TableName() string {
	return "oauth2_token"
}
