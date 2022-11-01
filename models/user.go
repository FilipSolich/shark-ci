package models

import (
	"errors"

	"github.com/FilipSolich/ci-server/db"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Identities []UserIdentity
	Repos      []Repository
}

func CreateUser(user *User) (*User, error) {
	result := db.DB.Create(user)
	return user, result.Error
}

func GetOrCreateUser(user *User) (*User, error) {
	var getUser User
	var err error
	result := db.DB.First(&getUser, user)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, result.Error
		}

		return CreateUser(user)
	}

	return &getUser, err
}

func (*User) TableName() string {
	return "user"
}
