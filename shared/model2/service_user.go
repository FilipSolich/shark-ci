package model2

import "time"

type ServiceUser struct {
	ID           int64     `json:"id"`
	Service      string    `json:"service"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	TokenExpire  time.Time `json:"token_expire"`
	UserID       int64     `json:"user_id"`
}
