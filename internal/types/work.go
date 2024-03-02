package types

import (
	"golang.org/x/oauth2"

	"github.com/shark-ci/shark-ci/internal/server/models"
)

type Work struct {
	Pipeline models.Pipeline `json:"pipeline"`
	Token    oauth2.Token    `json:"token"`
}
