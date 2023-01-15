package db

import (
	"context"

	"github.com/shark-ci/shark-ci/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetOrCreateIdentity(ctx context.Context, identity *models.Identity, user *models.User) (*models.Identity, error) {
	filter := bson.D{
		{Key: "serviceName", Value: identity.ServiceName},
		{Key: "username", Value: identity.Username},
	}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	err := Identities.FindOneAndUpdate(ctx, filter, bson.D{{Key: "$set", Value: identity}}, opts).Decode(identity)
	if err != nil {
		return nil, err
	}

	if user == nil {
		user, err = CreateUser(ctx)
		// TODO: Maybe delete created identity if user cannot be created
		if err != nil {
			return nil, err
		}
	}

	filter = bson.D{
		{Key: "_id", Value: user.ID},
		{Key: "identities", Value: bson.D{
			{Key: "$ne", Value: identity.ID},
		}},
	}
	update := bson.D{
		{Key: "$push", Value: bson.D{
			{Key: "identities", Value: identity.ID},
		}},
	}
	_, err = Users.UpdateOne(ctx, filter, update)
	return identity, err
}

func GetIdentityByUser(ctx context.Context, user *models.User, service string) (*models.Identity, error) {
	var identity models.Identity
	filter := bson.D{
		{Key: "_id", Value: bson.D{
			{Key: "$in", Value: user.Identities},
		}},
		{Key: "serviceName", Value: service},
	}
	err := Identities.FindOne(ctx, filter).Decode(&identity)
	if err != nil {
		return nil, err
	}

	return &identity, nil
}

func GetIdentityByUsername(ctx context.Context, username string, service string) (*models.Identity, error) {
	var identity models.Identity
	filter := bson.D{
		{Key: "username", Value: username},
		{Key: "serviceName", Value: service},
	}
	err := Identities.FindOne(ctx, filter).Decode(&identity)
	if err != nil {
		return nil, err
	}

	return &identity, nil
}
