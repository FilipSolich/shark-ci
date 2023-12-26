package models

import (
	"time"

	"golang.org/x/oauth2"
)

type ServiceUser struct {
	ID           int64      `json:"id"`
	Service      string     `json:"service"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	AccessToken  string     `json:"-"`
	RefreshToken *string    `json:"-"`
	TokenType    string     `json:"-"`
	TokenExpire  *time.Time `json:"-"`
	UserID       int64      `json:"user_id"`
}

func (su ServiceUser) Token() *oauth2.Token {
	token := &oauth2.Token{
		AccessToken: su.AccessToken,
		TokenType:   su.TokenType,
	}
	if su.RefreshToken != nil {
		token.RefreshToken = *su.RefreshToken
	}
	if su.TokenExpire != nil {
		token.Expiry = *su.TokenExpire
	}
	return token
}
