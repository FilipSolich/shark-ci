package middleware

import (
	"net/http"
	"os"

	"github.com/gorilla/handlers"
)

// TODO: Create logger with slog

func LoggingMiddleware(h http.Handler) http.Handler {
	return handlers.LoggingHandler(os.Stdout, h)
}
