package handler

import (
	"context"
	"net/http"

	"github.com/FilipSolich/shark-ci/ci-server/service"
	"github.com/FilipSolich/shark-ci/ci-server/session"
	"github.com/FilipSolich/shark-ci/ci-server/store"
	"github.com/FilipSolich/shark-ci/shared/model"
	"go.uber.org/zap"
)

type OAuth2Handler struct {
	l        *zap.SugaredLogger
	s        store.Storer
	services service.Services
}

func NewOAuth2Handler(l *zap.SugaredLogger, s store.Storer, services service.Services) *OAuth2Handler {
	return &OAuth2Handler{
		l:        l,
		s:        s,
		services: services,
	}
}

func (h *OAuth2Handler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	serviceName := r.URL.Query().Get("service")

	srv, ok := h.services[serviceName]
	if !ok {
		http.Error(w, "unknown OAuth2 provider: "+serviceName, http.StatusBadRequest)
		return
	}

	ctx := context.TODO()
	oauth2State, err := h.s.GetOAuth2StateByState(ctx, state)
	if err != nil {
		http.Error(w, "incorrect state", http.StatusBadRequest)
		return
	}

	h.s.DeleteOAuth2State(ctx, oauth2State) // TODO: What to do if delete fails
	if !oauth2State.IsValid() {
		http.Error(w, "oauth2 state expired", http.StatusBadRequest)
		return
	}

	// Get Oauth2 token from auth service.
	config := srv.OAuth2Config()
	token, err := config.Exchange(ctx, code)
	if err != nil {
		h.l.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Get or create new ServiceUser and new User if needed.
	// TODO: Get user from request and pass it into function call.
	serviceUser, err := srv.GetServiceUser(ctx, token)
	if err != nil {
		h.l.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Check if ServiceUser exists
	u, err := h.s.GetServiceUserByUniqueName(ctx, serviceUser.UniqueName)
	if err != nil {
		u = serviceUser
		err = h.s.CreateServiceUser(ctx, u)
	} else {
		err = h.s.UpdateServiceUserToken(ctx, u, serviceUser.Token)
	}
	if err != nil {
		h.l.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	user, err := h.s.GetUserByServiceUser(ctx, u)
	if err != nil {
		user := model.NewUser([]string{u.ID})
		err = h.s.CreateUser(ctx, user)
		if err != nil {
			h.l.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	// Store session.
	s, _ := session.Store.Get(r, "session")
	s.Values[session.SessionKey] = user.ID
	err = s.Save(r, w)
	if err != nil {
		h.l.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}
