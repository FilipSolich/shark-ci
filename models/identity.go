package models

import "golang.org/x/oauth2"

type Identity struct {
	ID          string       `bson:"_id,omitempty"`
	Username    string       `bson:"username,omitempty"`
	ServiceName string       `bson:"serviceName,omitempty"`
	UniqueName  string       `bson:"uniqueName,omitempty"`
	Token       oauth2.Token `bson:"token,omitempty"`
	Repos       []string     `bson:"repos,omitempty"`
}

func NewIdentity(username string, serviceName string, token *oauth2.Token) *Identity {
	return &Identity{
		Username:    username,
		ServiceName: serviceName,
		UniqueName:  serviceName + "/" + username,
		Token:       *token,
		Repos:       []string{},
	}
}
