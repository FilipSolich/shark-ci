package message_queue

import (
	"context"

	"github.com/FilipSolich/shark-ci/model"
)

type MessageQueuer interface {
	Close(ctx context.Context) error
	SendJob(ctx context.Context, job *model.Job) error
	JobChannel() (jobChannel, error)
}

type jobChannel chan model.Job
