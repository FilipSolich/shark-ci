package middleware

import (
	"context"
	"net/http"

	"github.com/shark-ci/shark-ci/internal/types"
)

type contextKey int

const userKey contextKey = 1

func ContextWithUser(ctx context.Context, user types.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func UserFromContext(ctx context.Context, w http.ResponseWriter) types.User {
	user, ok := ctx.Value(userKey).(types.User)
	if !ok {
		panic("User not found in context. Probably unused AuthMiddleware when login is required.")
	}
	return user
}
