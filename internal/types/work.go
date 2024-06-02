package types

import (
	"golang.org/x/oauth2"
)

type Work struct {
	Pipeline Pipeline     `json:"pipeline"`
	Token    oauth2.Token `json:"token"`
}
