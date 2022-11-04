package db

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	Client       *mongo.Client
	Users        *mongo.Collection
	Identities   *mongo.Collection
	Repos        *mongo.Collection
	Jobs         *mongo.Collection
	OAuth2States *mongo.Collection
)

type disconnect func(context.Context) error

func InitDatabase(mongoURI string) (disconnect, error) {
	ctx := context.Background()
	log.Println("Connection to database: " + mongoURI)
	Client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, err
	}

	err = Client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}

	log.Println("Database connected")

	Users = Client.Database("CIServerDB").Collection("users")
	Identities = Client.Database("CIServerDB").Collection("identities")
	Repos = Client.Database("CIServerDB").Collection("repos")
	Jobs = Client.Database("CIServerDB").Collection("jobs")
	OAuth2States = Client.Database("CIServerDB").Collection("oauth2States")

	return Client.Disconnect, nil
}
