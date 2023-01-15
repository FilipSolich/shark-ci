package models

type Identity struct {
	ID          string      `bson:"_id,omitempty"`
	ServiceName string      `bson:"serviceName,omitempty"`
	Username    string      `bson:"username,omitempty"`
	Token       OAuth2Token `bson:"token,omitempty"`
	Repos       []string    `bson:"repos,omitempty"`
}
