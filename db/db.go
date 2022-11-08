package db

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const databaseName = "CIServerDB"

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

	Users = Client.Database(databaseName).Collection("users")
	Identities = Client.Database(databaseName).Collection("identities")
	Repos = Client.Database(databaseName).Collection("repos")
	Jobs = Client.Database(databaseName).Collection("jobs")
	OAuth2States = Client.Database(databaseName).Collection("oauth2States")

	// Add TTL 30m to OAuth2States collection.
	opts := options.Index().SetExpireAfterSeconds(30 * 60)
	model := mongo.IndexModel{
		Keys:    bson.D{{Key: "createdAt", Value: 1}},
		Options: opts,
	}
	_, err = OAuth2States.Indexes().CreateOne(ctx, model)

	return Client.Disconnect, err
}
