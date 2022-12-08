package handlers

import (
	"context"
	"net/http"

	"github.com/shark-ci/shark-ci/db"
	"github.com/shark-ci/shark-ci/services"
	"github.com/shark-ci/shark-ci/sessions"
)

func OAuth2CallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	serviceName := r.URL.Query().Get("service")

	service, ok := services.Services[serviceName]
	if !ok {
		http.Error(w, "unknown OAuth2 provider: "+serviceName, http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	oauth2State, err := db.GetOAuth2StateByState(ctx, state)
	if err != nil || !oauth2State.IsValid() {
		http.Error(w, "incorrect state", http.StatusBadRequest)
		return
	}
	oauth2State.Delete(ctx) // TODO: What to do if delete fails?

	// Get Oauth2 token from auth service.
	config := service.GetOAuth2Config()
	token, err := config.Exchange(ctx, code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get or create new UserIdentity and new User if needed.
	// TODO: Get user from request and pass it into function call.
	identity, err := service.GetOrCreateUserIdentity(ctx, nil, token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user, err := db.GetUserByIdentity(ctx, identity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Store session.
	session, _ := sessions.Store.Get(r, "session")
	session.Values[sessions.SessionKey] = user.ID.Hex()
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}
