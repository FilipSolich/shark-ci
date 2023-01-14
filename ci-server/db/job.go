package db

import (
	"context"
	"fmt"
	"net"
	"net/url"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/shark-ci/shark-ci/ci-server/configs"
)

type Job struct {
	ID              primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Repo            primitive.ObjectID `json:"-" bson:"repo,omitempty"`
	Token           OAuth2Token        `json:"token,omitempty" bson:"token,omitempty"`
	CommitSHA       string             `json:"commmitSHA,omitempty" bson:"commmitSHA,omitempty"`
	CloneURL        string             `json:"cloneURL,omitempty" bson:"cloneURL,omitempty"`
	TargetURL       string             `json:"targetURL,omitempty" bson:"targetURL,omitempty"`
	ReportStatusURL string             `json:"reportStatusURL,omitempty" bson:"reportStatusURL,omitempty"`
	PublishLogsURL  string             `json:"publishLogsURL,omitempty" bson:"publishLogsURL,omitempty"`
}

func CreateJob(ctx context.Context, job *Job) (*Job, error) {
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

func GetJobByID(ctx context.Context, id primitive.ObjectID) (*Job, error) {
	job := &Job{}
	filter := bson.M{"_id": id}
	err := Jobs.FindOne(ctx, filter).Decode(job)

	return job, err
}

func (j *Job) createJobURLs() error {
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
