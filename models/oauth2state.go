package models

import (
	"time"
)

type OAuth2State struct {
	ID        string    `bson:"_id,omitempty"`
	State     string    `bson:"state,omitempty"`
	Expiry    time.Time `bson:"expiry,omitempty"`
	CreatedAt time.Time `bson:"createdAt,omitempty"`
}
