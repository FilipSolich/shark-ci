package db

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OAuth2State struct {
	ID     primitive.ObjectID `bson:"_id,omitempty"`
	State  string             `bson:"state,omitempty"`
	Expiry time.Time          `bson:"expiry,omitempty"`
}

func NewOAuth2State(state *OAuth2State) (*OAuth2State, error) {
	state.Expiry = time.Now().Add(30 * time.Minute)
	result, err := OAuth2States.InsertOne(context.Background(), state)
	state.ID = result.InsertedID.(primitive.ObjectID)
	return state, err
}

func (state *OAuth2State) IsValid() bool {
	valid := time.Now().Before(state.Expiry)
	if !valid {
		_, err := OAuth2States.DeleteOne(context.Background(), state)
		if err != nil {
			log.Println(err)
		}
	}
	return valid
}
