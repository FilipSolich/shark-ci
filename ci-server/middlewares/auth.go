package middlewares

import (
	"net/http"

	"github.com/FilipSolich/shark-ci/ci-server/sessions"
	"github.com/FilipSolich/shark-ci/ci-server/store"
	"github.com/gorilla/mux"
)

func AuthMiddleware(store store.Storer) mux.MiddlewareFunc {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, _ := sessions.Store.Get(r, "session")
			id, ok := session.Values[sessions.SessionKey].(string)
			if !ok {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}

			ctx := r.Context()
			user, err := store.GetUser(ctx, id)
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
