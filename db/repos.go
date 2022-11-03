package db

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Repo struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	RepoID      int64              `bson:"repoID,omitempty"`
	ServiceName string             `bson:"serviceName"`
	Name        string             `bson:"name,omitempty"`
	FullName    string             `bson:"fullName,omitempty"`
	Webhook     Webhook            `bson:"webhook,omitempty"`
}

type Webhook struct {
	WebhookID int64  `bson:"webhookID,omitempty"`
	Active    string `bson:"active,omitempty"`
}

func GetOrCreateRepo(ctx context.Context, repo *Repo) (*Repo, error) {
	filter := bson.D{
		{Key: "repoID", Value: repo.RepoID},
		{Key: "serviceName", Value: repo.ServiceName},
	}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	err := Identities.FindOneAndUpdate(ctx, filter, bson.D{{Key: "$set", Value: repo}}, opts).Decode(repo)
	if err != nil {
		return nil, err
	}

	return repo, nil
}
