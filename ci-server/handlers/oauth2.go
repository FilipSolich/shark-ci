package handlers

import (
	"context"
	"net/http"

	"github.com/shark-ci/shark-ci/ci-server/db"
	"github.com/shark-ci/shark-ci/ci-server/services"
	"github.com/shark-ci/shark-ci/ci-server/sessions"
	"github.com/shark-ci/shark-ci/ci-server/store"
)

type OAuth2Handler struct {
	store store.Storer
}

func NewOAuth2Handler(s store.Storer) *OAuth2Handler {
	return &OAuth2Handler{
		store: s,
	}
}

func (h *OAuth2Handler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	serviceName := r.URL.Query().Get("service")

	service, ok := services.Services[serviceName]
	if !ok {
		http.Error(w, "unknown OAuth2 provider: "+serviceName, http.StatusBadRequest)
		return
	}

	ctx := context.TODO()
	oauth2State, err := h.store.GetOAuth2StateByState(ctx, state)
	if err != nil {
		http.Error(w, "incorrect state", http.StatusBadRequest)
		return
	}

	h.store.DeleteOAuth2State(ctx, oauth2State) // TODO: What to do if delete fails
	if !oauth2State.IsValid() {
		http.Error(w, "oauth2 state expired", http.StatusBadRequest)
		return
	}

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
