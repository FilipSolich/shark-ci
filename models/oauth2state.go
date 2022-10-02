package models

import (
	"log"
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

func (*OAuth2State) TableName() string {
	return "oauth2_state"
}

func (state *OAuth2State) IsValid() bool {
	valid := time.Now().Before(state.Expiry)
	if !valid {
		result := db.DB.Delete(state)
		if result.Error != nil {
			log.Println(result.Error)
		}
	}
	return valid
}
