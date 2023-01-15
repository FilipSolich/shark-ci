package models

type User struct {
	ID         string   `bson:"_id,omitempty"`
	Identities []string `bson:"identities,omitempty"`
}
