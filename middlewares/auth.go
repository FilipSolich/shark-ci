package middlewares

import (
	"net/http"

	"github.com/FilipSolich/ci-server/db"
	"github.com/FilipSolich/ci-server/models"
	"github.com/FilipSolich/ci-server/sessions"
	"gorm.io/gorm/clause"
)

func AuthMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var user models.User
		session, _ := sessions.Store.Get(r, "session")
		id, ok := session.Values["id"].(uint)
		result := db.DB.Preload(clause.Associations).First(&user, id)
		if !ok || result.Error != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		ctx := r.Context()
		ctx = ContextWithUser(ctx, &user)
		r = r.WithContext(ctx)
		h.ServeHTTP(w, r)
	})
}
