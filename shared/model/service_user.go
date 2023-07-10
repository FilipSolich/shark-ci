package model

import "golang.org/x/oauth2"

type ServiceUser struct {
	ID          string       `bson:"_id,omitempty"`
	Username    string       `bson:"username,omitempty"`
	ServiceName string       `bson:"serviceName,omitempty"`
	UniqueName  string       `bson:"uniqueName,omitempty"`
	Token       oauth2.Token `bson:"token,omitempty"`
}

func NewServiceUser(username string, serviceName string, token *oauth2.Token) *ServiceUser {
	return &ServiceUser{
		Username:    username,
		ServiceName: serviceName,
		UniqueName:  serviceName + "/" + username,
		Token:       *token,
	}
}
