package message_queue

import (
	"context"

	"github.com/FilipSolich/shark-ci/shared/types"
)

type MessageQueuer interface {
	Close(ctx context.Context) error
	SendWork(ctx context.Context, work types.Work) error
	WorkChannel() (chan types.Work, error)
}
