package types

import (
	"github.com/FilipSolich/shark-ci/shared/model2"
	"golang.org/x/oauth2"
)

type Work struct {
	Pipeline model2.Pipeline `json:"pipeline"`
	Token    oauth2.Token    `json:"token"`
}
