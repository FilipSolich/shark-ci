package db

import (
	"context"
	"fmt"
	"net"
	"net/url"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/shark-ci/shark-ci/ci-server/configs"
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

func (j *models.Job) createJobURLs() error {
	baseURL := url.URL{
		Scheme: "https",
		Host:   net.JoinHostPort(configs.Host, configs.Port),
	}

	var err error
	targetURL := baseURL
	targetURL.Path, err = url.JoinPath(configs.JobsPath, fmt.Sprint(j.ID.Hex()))
	if err != nil {
		return err
	}
	reportStatusURL := baseURL
	reportStatusURL.Path, err = url.JoinPath(configs.JobsReportStatusHandlerPath, fmt.Sprint(j.ID.Hex()))
	if err != nil {
		return err
	}
	publishLogsURL := baseURL
	publishLogsURL.Path, err = url.JoinPath(configs.JobsPublishLogsHandlerPath, fmt.Sprint(j.ID.Hex()))
	if err != nil {
		return err
	}

	j.TargetURL = targetURL.String()
	j.ReportStatusURL = reportStatusURL.String()
	j.PublishLogsURL = publishLogsURL.String()
	return nil
}
