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
	WebhookID int64 `bson:"webhookID,omitempty"`
	Active    bool  `bson:"active,omitempty"`
}

func GetOrCreateRepo(ctx context.Context, repo *Repo, identity *Identity) (*Repo, error) {
	filter := bson.D{
		{Key: "repoID", Value: repo.RepoID},
		{Key: "serviceName", Value: repo.ServiceName},
	}
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "name", Value: repo.Name},
			{Key: "fullName", Value: repo.FullName},
		}},
	}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	err := Repos.FindOneAndUpdate(ctx, filter, update, opts).Decode(repo)
	if err != nil {
		return nil, err
	}

	filter = bson.D{
		{Key: "_id", Value: identity.ID},
		{Key: "repos", Value: bson.D{
			{Key: "$ne", Value: repo.ID},
		}},
	}
	update = bson.D{
		{Key: "$push", Value: bson.D{
			{Key: "repos", Value: repo.ID},
		}},
	}
	_, err = Identities.UpdateOne(ctx, filter, update)
	return repo, nil
}

func GetRepoByID(ctx context.Context, id primitive.ObjectID) (*Repo, error) {
	var repo Repo
	filter := bson.D{{Key: "_id", Value: id}}
	err := Repos.FindOne(ctx, filter).Decode(&repo)
	if err != nil {
		return nil, err
	}

	return &repo, nil
}

func GetRepoByFullName(ctx context.Context, fullName string, service string) (*Repo, error) {
	var repo Repo
	filter := bson.D{
		{Key: "serviceName", Value: service},
		{Key: "fullName", Value: fullName},
	}
	err := Repos.FindOne(ctx, filter).Decode(&repo)
	if err != nil {
		return nil, err
	}

	return &repo, nil
}

func (r *Repo) DeleteWebhook(ctx context.Context) error {
	data := bson.D{
		{Key: "$unset", Value: bson.D{
			{Key: "webhook", Value: ""},
		}},
	}
	_, err := Repos.UpdateByID(ctx, r.ID, data)
	return err
}

func (r *Repo) UpdateWebhook(ctx context.Context, webhook *Webhook) error {
	data := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "webhook", Value: bson.D{
				{Key: "webhookID", Value: webhook.WebhookID},
				{Key: "active", Value: webhook.Active},
			}},
		}},
	}
	_, err := Repos.UpdateByID(ctx, r.ID, data)
	return err
}

func (r *Repo) GetOwner(ctx context.Context) (*Identity, error) {
	var identity Identity
	filter := bson.D{
		{Key: "repos", Value: r.ID},
	}
	err := Identities.FindOne(ctx, filter).Decode(&identity)
	if err != nil {
		return nil, err
	}

	return &identity, nil
}
