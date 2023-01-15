package models

import (
	"time"
)

type OAuth2State struct {
	ID        string    `bson:"_id,omitempty"`
	State     string    `bson:"state,omitempty"`
	CreatedAt time.Time `bson:"createdAt,omitempty"`
	Expire    time.Time `bson:"expire,omitempty"`
}

func NewOAuth2Satate(state string, expireAfter time.Duration) *OAuth2State {
	createdAt := time.Now()
	return &OAuth2State{
		State:     state,
		CreatedAt: createdAt,
		Expire:    createdAt.Add(expireAfter),
	}
}

func (s *OAuth2State) IsValid() bool {
	valid := time.Now().Before(s.Expire)
	return valid
}
