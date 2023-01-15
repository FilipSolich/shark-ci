package models

type User struct {
	ID         string   `bson:"_id,omitempty"`
	Identities []string `bson:"identities,omitempty"`
}

func NewUser() *User {
	return &User{
		Identities: []string{},
	}
}
