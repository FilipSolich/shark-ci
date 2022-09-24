package middlewares

import (
	"context"

	"github.com/FilipSolich/ci-server/models"
)

type userKey int

const key userKey = 1

func ContextWithUser(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, key, user)
}

func UserFromContext(ctx context.Context) (*models.User, bool) {
	user, ok := ctx.Value(key).(*models.User)
	return user, ok
}
