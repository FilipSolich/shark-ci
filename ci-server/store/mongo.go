package store

import (
	"context"

	"github.com/FilipSolich/shark-ci/ci-server/log"
	"github.com/FilipSolich/shark-ci/shared/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/oauth2"
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

var _ Storer = &MongoStore{}

func NewMongoStore(mongoURI string) (*MongoStore, error) {
	ms := &MongoStore{}
	var err error

	log.L.Info("Connecting to MongoDB")
	ctx := context.TODO()
	ms.client, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, err
	}

	err = ms.Ping(ctx)
	if err != nil {
		return nil, err
	}
	log.L.Info("MongoDB connected")

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

func (ms *MongoStore) GetUser(ctx context.Context, id string) (*model.User, error) {
	user := &model.User{}
	err := ms.users.FindOne(ctx, bson.M{"_id": id}).Decode(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (ms *MongoStore) GetUserByServiceUser(ctx context.Context, i *model.ServiceUser) (*model.User, error) {
	user := &model.User{}
	err := ms.users.FindOne(ctx, bson.M{"identities": i.ID}).Decode(&user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (ms *MongoStore) CreateUser(ctx context.Context, u *model.User) error {
	u.ID = primitive.NewObjectID().Hex()
	_, err := ms.users.InsertOne(ctx, u)
	return err
}

func (ms *MongoStore) DeleteUser(ctx context.Context, u *model.User) error {
	_, err := ms.users.DeleteOne(ctx, bson.M{"_id": u.ID})
	return err
}

func (ms *MongoStore) GetServiceUser(ctx context.Context, id string) (*model.ServiceUser, error) {
	serviceUser := &model.ServiceUser{}
	err := ms.identities.FindOne(ctx, bson.M{"_id": id}).Decode(serviceUser)
	if err != nil {
		return nil, err
	}

	return serviceUser, nil
}

func (ms *MongoStore) GetServiceUserByUniqueName(ctx context.Context, uniqueName string) (*model.ServiceUser, error) {
	serviceUser := &model.ServiceUser{}
	err := ms.identities.FindOne(ctx, bson.M{"uniqueName": uniqueName}).Decode(serviceUser)
	if err != nil {
		return nil, err
	}

	return serviceUser, nil
}

func (ms *MongoStore) GetServiceUserByRepo(ctx context.Context, r *model.Repo) (*model.ServiceUser, error) {
	repo, err := ms.GetRepo(ctx, r.ID)
	if err != nil {
		return nil, err
	}

	return ms.GetServiceUser(ctx, repo.ServiceUserID)
}

func (ms *MongoStore) GetServiceUserByUser(ctx context.Context, user *model.User, serviceName string) (*model.ServiceUser, error) {
	serviceUser := &model.ServiceUser{}
	filter := bson.D{
		{Key: "_id", Value: bson.D{
			{Key: "$in", Value: user.Identities},
		}},
		{Key: "serviceName", Value: serviceName},
	}
	err := ms.identities.FindOne(ctx, filter).Decode(&serviceUser)
	if err != nil {
		return nil, err
	}

	return serviceUser, nil
}

func (ms *MongoStore) CreateServiceUser(ctx context.Context, i *model.ServiceUser) error {
	i.ID = primitive.NewObjectID().Hex()
	_, err := ms.identities.InsertOne(ctx, i)
	return err
}

func (ms *MongoStore) UpdateServiceUserToken(ctx context.Context, i *model.ServiceUser, token oauth2.Token) error {
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "token", Value: token},
		}},
	}
	_, err := ms.identities.UpdateByID(ctx, i.ID, update)
	return err
}

func (ms *MongoStore) DeleteServiceUser(ctx context.Context, i *model.ServiceUser) error {
	_, err := ms.identities.DeleteOne(ctx, bson.M{"_id": i.ID})
	return err
}

func (ms *MongoStore) GetRepo(ctx context.Context, id string) (*model.Repo, error) {
	repo := &model.Repo{}
	err := ms.repos.FindOne(ctx, bson.M{"_id": id}).Decode(&repo)
	if err != nil {
		return nil, err
	}

	return repo, nil
}

func (ms *MongoStore) GetRepoByUniqueName(ctx context.Context, uniqueName string) (*model.Repo, error) {
	repo := &model.Repo{}
	err := ms.repos.FindOne(ctx, bson.M{"uniqueName": uniqueName}).Decode(&repo)
	if err != nil {
		return nil, err
	}

	return repo, nil
}

func (ms *MongoStore) CreateRepo(ctx context.Context, r *model.Repo) error {
	r.ID = primitive.NewObjectID().Hex()
	_, err := ms.repos.InsertOne(ctx, r)
	return err
}

func (ms *MongoStore) UpdateRepoWebhook(ctx context.Context, r *model.Repo) error {
	data := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "webhookID", Value: r.WebhookID},
			{Key: "webhookActive", Value: r.WebhookActive},
		}},
	}
	_, err := ms.repos.UpdateByID(ctx, r.ID, data)
	return err
}

func (ms *MongoStore) DeleteRepo(ctx context.Context, r *model.Repo) error {
	_, err := ms.repos.DeleteOne(ctx, bson.M{"_id": r.ID})
	return err
}

func (ms *MongoStore) GetOAuth2StateByState(ctx context.Context, state string) (*model.OAuth2State, error) {
	oauth2State := &model.OAuth2State{}
	err := ms.oauth2States.FindOne(ctx, bson.M{"state": state}).Decode(&oauth2State)
	if err != nil {
		return nil, err
	}

	return oauth2State, nil
}

func (ms *MongoStore) CreateOAuth2State(ctx context.Context, s *model.OAuth2State) error {
	s.ID = primitive.NewObjectID().Hex()
	_, err := ms.oauth2States.InsertOne(ctx, s)
	return err
}

func (ms *MongoStore) DeleteOAuth2State(ctx context.Context, s *model.OAuth2State) error {
	_, err := ms.oauth2States.DeleteOne(ctx, bson.M{"_id": s.ID})
	return err
}

func (ms *MongoStore) GetJob(ctx context.Context, id string) (*model.Job, error) {
	job := &model.Job{}
	err := ms.jobs.FindOne(ctx, bson.M{"_id": id}).Decode(&job)
	if err != nil {
		return nil, err
	}

	return job, nil
}

func (ms *MongoStore) CreateJob(ctx context.Context, j *model.Job) error {
	j.ID = primitive.NewObjectID().Hex()
	_, err := ms.jobs.InsertOne(ctx, j)
	return err
}

func (ms *MongoStore) DeleteJob(ctx context.Context, j *model.Job) error {
	_, err := ms.jobs.DeleteOne(ctx, bson.M{"_id": j.ID})
	return err
}
