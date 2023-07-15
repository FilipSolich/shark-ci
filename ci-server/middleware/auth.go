package middleware

import (
	"net/http"

	"github.com/FilipSolich/shark-ci/ci-server/session"
	"github.com/FilipSolich/shark-ci/ci-server/store"
	"github.com/gorilla/mux"
)

func AuthMiddleware(s store.Storer) mux.MiddlewareFunc {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sess, _ := session.Store.Get(r, "session")
			id, ok := sess.Values[session.SessionKey].(string)
			if !ok {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}

			ctx := r.Context()
			user, err := s.GetUser(ctx, id)
			if err != nil {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}

			ctx = ContextWithUser(ctx, user)
			r = r.WithContext(ctx)
			h.ServeHTTP(w, r)
		})
	}
}
