package handlers

import (
	"context"
	"net/http"

	"github.com/FilipSolich/shark-ci/ci-server/service"
	"github.com/FilipSolich/shark-ci/ci-server/session"
	"github.com/FilipSolich/shark-ci/ci-server/store"
	"github.com/FilipSolich/shark-ci/model"
)

type OAuth2Handler struct {
	store      store.Storer
	serviceMap service.ServiceMap
}

func NewOAuth2Handler(s store.Storer, serviceMap service.ServiceMap) *OAuth2Handler {
	return &OAuth2Handler{
		store:      s,
		serviceMap: serviceMap,
	}
}

func (h *OAuth2Handler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	serviceName := r.URL.Query().Get("service")

	srv, ok := h.serviceMap[serviceName]
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
	config := srv.OAuth2Config()
	token, err := config.Exchange(ctx, code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get or create new UserIdentity and new User if needed.
	// TODO: Get user from request and pass it into function call.
	serviceIdentity, err := srv.GetUserIdentity(ctx, token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if identity exists
	identity, err := h.store.GetIdentityByUniqueName(ctx, serviceIdentity.UniqueName)
	if err != nil {
		identity = serviceIdentity
		err = h.store.CreateIdentity(ctx, identity)
	} else {
		err = h.store.UpdateIdentityToken(ctx, identity, serviceIdentity.Token)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user, err := h.store.GetUserByIdentity(ctx, identity)
	if err != nil {
		user := model.NewUser([]string{identity.ID})
		err = h.store.CreateUser(ctx, user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Store session.
	s, _ := session.Store.Get(r, "session")
	s.Values[session.SessionKey] = user.ID
	err = s.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}
