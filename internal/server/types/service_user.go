package types

import (
	"time"

	"golang.org/x/oauth2"
)

type ServiceUserRepoFetchInfo struct {
	ID      int64
	Service string
	Token   oauth2.Token
}

type ServiceUser struct {
	ID           int64
	Service      string
	Username     string
	Email        string
	AccessToken  string
	RefreshToken *string
	TokenType    string
	TokenExpire  *time.Time
	UserID       int64
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
