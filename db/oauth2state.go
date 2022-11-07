package db

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OAuth2State struct {
	ID     primitive.ObjectID `bson:"_id,omitempty"`
	State  string             `bson:"state,omitempty"`
	Expiry time.Time          `bson:"expiry,omitempty"`
}

func NewOAuth2State(state *OAuth2State) (*OAuth2State, error) {
	state.ID = primitive.NewObjectID()
	state.Expiry = time.Now().Add(30 * time.Minute)
	_, err := OAuth2States.InsertOne(context.Background(), state)
	return state, err
}

func GetOAuth2StateByState(ctx context.Context, state string) (*OAuth2State, error) {
	var oauth2State OAuth2State
	filter := bson.D{{Key: "state", Value: state}}
	err := OAuth2States.FindOne(ctx, filter).Decode(&oauth2State)
	if err != nil {
		return nil, err
	}

	return &oauth2State, nil
}

func (state *OAuth2State) Delete(ctx context.Context) error {
	_, err := OAuth2States.DeleteOne(ctx, state)
	return err
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
