package db

import (
	"context"

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

func CreateRepo(repo *Repo) (*Repo, error) {
	result, err := OAuth2States.UpdateOne(context.Background(), repo, repo, options.Update().SetUpsert(true))
	repo.ID = result.UpsertedID.(primitive.ObjectID)
	return repo, err
}
