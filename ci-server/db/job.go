package db

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/shark-ci/shark-ci/models"
)

func CreateJob(ctx context.Context, job *models.Job) (*models.Job, error) {
	job.ID = primitive.NewObjectID()
	err := job.createJobURLs()
	if err != nil {
		return nil, err
	}

	_, err = Jobs.InsertOne(ctx, job)
	if err != nil {
		return nil, err
	}

	return job, nil
}

func GetJobByID(ctx context.Context, id primitive.ObjectID) (*models.Job, error) {
	job := &models.Job{}
	filter := bson.M{"_id": id}
	err := Jobs.FindOne(ctx, filter).Decode(job)

	return job, err
}
