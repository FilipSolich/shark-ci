package middleware

import (
	"context"
	"net/http"

	"github.com/FilipSolich/shark-ci/shared/model2"
)

type contextKey int

const userKey contextKey = 1

func ContextWithUser(ctx context.Context, user *model2.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func UserFromContext(ctx context.Context, w http.ResponseWriter) (*model2.User, bool) {
	user, ok := ctx.Value(userKey).(*model2.User)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
	}
	return user, ok
}
