package model2

import (
	"time"

	"github.com/google/uuid"
)

type OAuth2State struct {
	State  uuid.UUID `json:"state"`
	Expire time.Time `json:"expire"`
}

func (s OAuth2State) IsValid() bool {
	return s.Expire.After(time.Now())
}
