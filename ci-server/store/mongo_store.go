package store

import (
	"context"
	"log"

	"github.com/shark-ci/shark-ci/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoStore struct {
	client       *mongo.Client
	db           *mongo.Database
	users        *mongo.Collection
	identities   *mongo.Collection
	repos        *mongo.Collection
	jobs         *mongo.Collection
	oauth2States *mongo.Collection
}

func NewMongoStore(mongoURI string) (*MongoStore, error) {
	ms := &MongoStore{}
	var err error

	// TODO: Change message. mongoURI may contain password.
	log.Println("Connecting to database: " + mongoURI)
	ctx := context.TODO()
	ms.client, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, err
	}

	err = ms.Ping(ctx)
	if err != nil {
		return nil, err
	}
	log.Println("Database connected")

	ms.db = ms.client.Database("CIServer")
	ms.users = ms.db.Collection("users")
	ms.identities = ms.db.Collection("identities")
	ms.repos = ms.db.Collection("repos")
	ms.jobs = ms.db.Collection("jobs")
	ms.oauth2States = ms.db.Collection("oauth2states")

	// Add TTL 30m to OAuth2States collection.
	opts := options.Index().SetExpireAfterSeconds(30 * 60)
	model := mongo.IndexModel{
		Keys:    bson.D{{Key: "createdAt", Value: 1}},
		Options: opts,
	}
	_, err = ms.oauth2States.Indexes().CreateOne(ctx, model)
	return ms, err
}

func (ms *MongoStore) Ping(ctx context.Context) error {
	return ms.client.Ping(ctx, nil)
}

func (ms *MongoStore) Close(ctx context.Context) error {
	return ms.client.Disconnect(ctx)
}

func (ms *MongoStore) Migrate(ctx context.Context) error {
	return nil
}

func (ms *MongoStore) GetUser(ctx context.Context, id string) (*models.User, error) {
	user := &models.User{}
	err := ms.users.FindOne(ctx, bson.M{"_id": id}).Decode(user)
	return user, err
}

func (ms *MongoStore) GetUserByIdentity(ctx context.Context, i *models.Identity) (*models.User, error) {
	user := &models.User{}
	err := ms.users.FindOne(ctx, bson.M{"identities": i.ID}).Decode(&user)
	return user, err
}

func (ms *MongoStore) CreateUser(ctx context.Context, u *models.User) error {
	u.ID = primitive.NewObjectID().Hex()
	_, err := ms.users.InsertOne(ctx, u)
	return err
}

func (ms *MongoStore) UpdateUser(ctx context.Context, u *models.User) error {
	update := bson.D{{
		Key:   "$set",
		Value: u,
	}}
	_, err := ms.users.UpdateByID(ctx, u.ID, update)
	return err
}

func (ms *MongoStore) DeleteUser(ctx context.Context, u *models.User) error {
	_, err := ms.users.DeleteOne(ctx, bson.M{"_id": u.ID})
	return err
}

func (ms *MongoStore) CreateOAuth2State(ctx context.Context, s *models.OAuth2State) error {
	return nil
}
