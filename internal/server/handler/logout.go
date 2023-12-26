package handler

import (
	"net/http"

	"github.com/shark-ci/shark-ci/internal/server/session"
)

// TODO: Move under login handler with register.
type LogoutHandler struct{}

func NewLogoutHandler() *LogoutHandler {
	return &LogoutHandler{}
}

func (h *LogoutHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	session, _ := session.Store.Get(r, "session")
	session.Options.MaxAge = -1
	err := session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}
