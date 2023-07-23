package message_queue

import (
	"context"

	"github.com/FilipSolich/shark-ci/shared/model"
	"github.com/FilipSolich/shark-ci/shared/types"
)

type MessageQueuer interface {
	Close(ctx context.Context) error
	SendWork(ctx context.Context, work types.Work) error
	JobChannel() (jobChannel, error)
}

type jobChannel chan model.Job
