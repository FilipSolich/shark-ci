package message_queue

import (
	"context"

	"github.com/shark-ci/shark-ci/models"
)

type MessageQueuer interface {
	Close(ctx context.Context) error
	SendJob(ctx context.Context, job *models.Job) error
	JobChannel() (jobChannel, error)
}

type jobChannel chan models.Job
