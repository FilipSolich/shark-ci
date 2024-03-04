package middleware

import (
	"context"
	"log/slog"
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
		slog.Error("User not found in context.")
		w.WriteHeader(http.StatusUnauthorized) // TODO: render unauthorized page or redirect to login page
	}
	return user, ok
}
