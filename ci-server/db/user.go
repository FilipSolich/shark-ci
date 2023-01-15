package db

import (
	"context"

	"github.com/shark-ci/shark-ci/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func CreateUser(ctx context.Context) (*models.User, error) {
	var user models.User
	result, err := Users.InsertOne(ctx, user)
	if err != nil {
		return nil, err
	}

	user.ID = result.InsertedID.(primitive.ObjectID)
	return &user, nil
}

func GetUserByID(ctx context.Context, id primitive.ObjectID) (*models.User, error) {
	var user models.User
	filter := bson.D{{Key: "_id", Value: id}}
	err := Users.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func GetUserByIdentity(ctx context.Context, identity *models.Identity) (*models.User, error) {
	var user models.User
	filter := bson.D{{Key: "identities", Value: identity.ID}}
	err := Users.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (u *models.User) IsUserIdentity(ctx context.Context, identity *models.Identity) bool {
	filter := bson.D{
		{Key: "_id", Value: u.ID},
		{Key: "identities", Value: identity.ID},
	}
	err := Users.FindOne(ctx, filter)
	if err != nil {
		return false
	}

	return true
}
