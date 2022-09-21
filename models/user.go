package models

import (
	"github.com/FilipSolich/ci-server/db"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type User struct {
	gorm.Model
	Username string
	Service  string
	Token    OAuth2Token
	Webhooks []Webhook
}

func GetOrCreateUser(user *User, token *OAuth2Token) (*User, error) {
	var u User
	result := db.DB.Preload(clause.Associations).First(&u, user)
	if result.Error != nil {
		user.Token = *token
		result = db.DB.Create(user)
		if result.Error != nil {
			return nil, result.Error
		}
		return user, nil
	}
	db.DB.Delete(&u.Token)
	u.Token = *token
	a := db.DB.Save(&u)
	if a.Error != nil {
		panic(a.Error)
	}
	return &u, nil
}
