package middlewares

import (
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/FilipSolich/ci-server/db"
	"github.com/FilipSolich/ci-server/sessions"
)

func AuthMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := sessions.Store.Get(r, "session")
		id, ok := session.Values[sessions.SessionKey].(string)
		objectID, err := primitive.ObjectIDFromHex(id)
		if err != nil || !ok {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		ctx := r.Context()
		user, err := db.GetUserByID(ctx, objectID)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		ctx = ContextWithUser(ctx, user)
		r = r.WithContext(ctx)
		h.ServeHTTP(w, r)
	})
}
