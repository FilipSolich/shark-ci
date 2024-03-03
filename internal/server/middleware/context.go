package middleware

import (
	"context"
	"net/http"

	"github.com/shark-ci/shark-ci/internal/server/types"
)

type contextKey int

const userKey contextKey = 1

func ContextWithUser(ctx context.Context, user types.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func UserFromContext(ctx context.Context, w http.ResponseWriter) (types.User, bool) {
	user, ok := ctx.Value(userKey).(types.User)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
	}
	return user, ok
}
