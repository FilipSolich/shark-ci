package middlewares

import (
	"context"
	"net/http"

	"github.com/shark-ci/shark-ci/models"
)

type contextKey int

const userKey contextKey = 1

func ContextWithUser(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func UserFromContext(ctx context.Context, w http.ResponseWriter) (*models.User, bool) {
	user, ok := ctx.Value(userKey).(*models.User)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
	}
	return user, ok
}
