package middleware

import (
	"net/http"
	"os"

	"github.com/gorilla/handlers"
)

func LoggingMiddleware(h http.Handler) http.Handler {
	return handlers.LoggingHandler(os.Stdout, h)
}
