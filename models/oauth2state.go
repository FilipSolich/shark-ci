package models

import (
	"fmt"
	"time"

	"github.com/FilipSolich/ci-server/db"
	"gorm.io/gorm"
)

type OAuth2State struct {
	gorm.Model
	State  string
	Expiry time.Time
}

func NewOAuth2State(state *OAuth2State) (*OAuth2State, error) {
	state.Expiry = time.Now().Add(30 * time.Minute)
	result := db.DB.Create(state)
	return state, result.Error
}

func (state *OAuth2State) IsValid() bool {
	valid := time.Now().Before(state.Expiry)
	if !valid {
		result := db.DB.Delete(state)
		if result.Error != nil {
			fmt.Println(result.Error) // TODO: log
		}
	}
	return valid
}
