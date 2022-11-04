package db

import (
	"context"
	"fmt"
	"net"
	"net/url"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/FilipSolich/ci-server/configs"
)

type Job struct {
	ID              primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Token           OAuth2Token        `json:"token,omitempty" bson:"token,omitempty"`
	CommitSHA       string             `json:"commmitSHA,omitempty" bson:"commmitSHA,omitempty"`
	CloneURL        string             `json:"cloneURL,omitempty" bson:"cloneURL,omitempty"`
	TargetURL       string             `json:"targetURL,omitempty" bson:"targetURL,omitempty"`
	ReportStatusURL string             `json:"reportStatusURL,omitempty" bson:"reportStatusURL,omitempty"`
	PublishLogsURL  string             `json:"publishLogsURL,omitempty" bson:"publishLogsURL,omitempty"`
}

func CreateJob(ctx context.Context, job *Job) (*Job, error) {
	err := job.createJobURLs()
	if err != nil {
		return nil, err
	}

	result, err := Jobs.InsertOne(ctx, job)
	if err != nil {
		return nil, err
	}

	job.ID = result.InsertedID.(primitive.ObjectID)
	return job, nil
}

func (j *Job) createJobURLs() error {
	baseURL := url.URL{
		Scheme: "https",
		Host:   net.JoinHostPort(configs.Host, configs.Port),
	}

	var err error
	targetURL := baseURL
	targetURL.Path, err = url.JoinPath(configs.JobsPath, fmt.Sprint(j.ID))
	if err != nil {
		return err
	}
	reportStatusURL := baseURL
	reportStatusURL.Path, err = url.JoinPath(configs.JobsPath, configs.JobsReportStatusHandlerPath, fmt.Sprint(j.ID))
	if err != nil {
		return err
	}
	publishLogsURL := baseURL
	publishLogsURL.Path, err = url.JoinPath(configs.JobsPath, configs.JobsPublishLogsHandlerPath, fmt.Sprint(j.ID))
	if err != nil {
		return err
	}

	j.TargetURL = targetURL.String()
	j.ReportStatusURL = reportStatusURL.String()
	j.PublishLogsURL = publishLogsURL.String()
	return nil
}
