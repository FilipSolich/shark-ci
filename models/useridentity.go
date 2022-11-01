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
	var identity UserIdentity
	var err error
	result := db.DB.First(&identity, ui)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, result.Error
		}

		_, err = CreateUserIdentity(ui)
		if err != nil {
			return nil, err
		}

		user := &User{
			Identities: []UserIdentity{*ui},
		}
		_, err = GetOrCreateUser(user)
		ui.UserID = user.ID
		return ui, err
	}

	return &identity, nil
}

func (*UserIdentity) TableName() string {
	return "user_identity"
}

func (ui *UserIdentity) UpdateOAuth2Token(token *oauth2.Token) error {
	db.DB.Where("user_identity_id = ?", ui.ID).Delete(&ui.Token)
	t := OAuth2Token{
		Token: *token,
	}
	ui.Token = t
	return db.DB.Save(ui).Error
}
