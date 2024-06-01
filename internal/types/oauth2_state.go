package types

import (
	"time"

	"github.com/google/uuid"
)

type OAuth2State struct {
	State  uuid.UUID
	Expire time.Time
}
