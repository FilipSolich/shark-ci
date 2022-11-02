package models

import (
	"errors"

	"github.com/FilipSolich/ci-server/db"
	"golang.org/x/oauth2"
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

type Userm struct {
	Identities []Identity `bson:"identities"`
}

type Identity struct {
	ServiceName string       `bson:"serviceName"`
	Username    string       `bson:"username"`
	OAuth2Token OAuth2Tokenm `bson:"oauth2Token"`
	Repos       []Reposm     `bson:"repos"`
}

type OAuth2Tokenm struct {
	oauth2.Token
}

type Reposm struct {
	RepoID   string   `bson:"repoID"`
	Name     string   `bson:"name"`
	FullName string   `bson:"fullName"`
	Webhook  Webhookm `bson:"webhook"`
}

type Webhookm struct {
	WebhookID string `bson:"webhookID"`
	Active    bool   `bson:"active"`
}
