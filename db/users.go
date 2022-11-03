package db

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID         primitive.ObjectID   `bson:"_id,omitempty"`
	Identities []primitive.ObjectID `bson:"identities,omitempty"`
}

func CreateUser(ctx context.Context) (*User, error) {
	var user User
	result, err := Users.InsertOne(ctx, user)
	if err != nil {
		return nil, err
	}

	user.ID = result.InsertedID.(primitive.ObjectID)
	return &user, nil
}

func GetUserByID(ctx context.Context, id primitive.ObjectID) (*User, error) {
	var user User
	filter := bson.D{{Key: "_id", Value: id}}
	err := Users.FindOne(ctx, filter).Decode(&user)
	return &user, err
}

func GetUserByIdentity(ctx context.Context, identity *Identity) (*User, error) {
	var user User
	filter := bson.D{{Key: "identities", Value: identity.ID}}
	err := Users.FindOne(ctx, filter).Decode(&user)
	return &user, err
}
