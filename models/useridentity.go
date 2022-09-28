package models

import (
	"errors"

	"github.com/FilipSolich/ci-server/db"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

type UserIdentity struct {
	gorm.Model
	Token       OAuth2Token
	UserID      uint
	ServiceName string
	Username    string
}

func CreateUserIdentity(ui *UserIdentity) (*UserIdentity, error) {
	result := db.DB.Create(ui)
	return ui, result.Error
}

func GetOrCreateUserIdentity(ui *UserIdentity) (*UserIdentity, error) {
	var identity *UserIdentity
	var err error
	result := db.DB.First(identity, ui)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, result.Error
		}

		identity, err = CreateUserIdentity(ui)
	}

	user := &User{
		Identities: []UserIdentity{*identity},
	}
	user, err = GetOrCreateUser(user)
	identity.UserID = user.ID
	return identity, err
}

func (ui *UserIdentity) UpdateOAuth2Token(token *oauth2.Token) error {
	t, err := GetOrCreateOAuth2Token(&OAuth2Token{UserIdentityID: ui.ID})
	if err != nil {
		return err
	}

	t.Token = *token
	return db.DB.Save(t).Error
}
