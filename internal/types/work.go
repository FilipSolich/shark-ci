package types

import (
	"github.com/shark-ci/shark-ci/internal/server/models"
	"golang.org/x/oauth2"
)

type Work struct {
	Pipeline models.Pipeline `json:"pipeline"`
	Token    oauth2.Token    `json:"token"`
}
