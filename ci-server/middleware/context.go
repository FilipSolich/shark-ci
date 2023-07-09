package middleware

import (
	"context"
	"net/http"

	"github.com/FilipSolich/shark-ci/model"
)

type contextKey int

const userKey contextKey = 1

func ContextWithUser(ctx context.Context, user *model.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func UserFromContext(ctx context.Context, w http.ResponseWriter) (*model.User, bool) {
	user, ok := ctx.Value(userKey).(*model.User)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
	}
	return user, ok
}
