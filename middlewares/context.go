package middlewares

import (
	"context"
	"net/http"

	"github.com/shark-ci/shark-ci/db"
)

type contextKey int

const userKey contextKey = 1

func ContextWithUser(ctx context.Context, user *db.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func UserFromContext(ctx context.Context, w http.ResponseWriter) (*db.User, bool) {
	user, ok := ctx.Value(userKey).(*db.User)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
	}
	return user, ok
}
