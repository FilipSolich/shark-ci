package messagequeue

import (
	"context"

	"github.com/shark-ci/shark-ci/internal/types"
)

type MessageQueuer interface {
	Close(ctx context.Context) error
	SendWork(ctx context.Context, work types.Work) error
	WorkChannel() (chan types.Work, error)
}
