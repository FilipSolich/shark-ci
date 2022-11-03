package middlewares

import (
	"context"
	"net/http"

	"github.com/FilipSolich/ci-server/db"
)

type contextKey int

const key contextKey = 1

func ContextWithUser(ctx context.Context, user *db.User) context.Context {
	return context.WithValue(ctx, key, user)
}

func UserFromContext(ctx context.Context, w http.ResponseWriter) (*db.User, bool) {
	user, ok := ctx.Value(key).(*db.User)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
	}
	return user, ok
}
