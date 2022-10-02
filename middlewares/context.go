package middlewares

import (
	"context"
	"net/http"

	"github.com/FilipSolich/ci-server/models"
)

type contextKey int

const key contextKey = 1

func ContextWithUser(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, key, user)
}

func UserFromContext(ctx context.Context, w http.ResponseWriter) (*models.User, bool) {
	user, ok := ctx.Value(key).(*models.User)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
	}
	return user, ok
}
