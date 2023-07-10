package model

type User struct {
	ID         string   `bson:"_id,omitempty"`
	Identities []string `bson:"identities,omitempty"`
}

func NewUser(identities []string) *User {
	return &User{
		Identities: identities,
	}
}
