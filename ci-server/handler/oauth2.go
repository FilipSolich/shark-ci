package handler

import (
	"net/http"

	"github.com/FilipSolich/shark-ci/ci-server/service"
	"github.com/FilipSolich/shark-ci/ci-server/session"
	"github.com/FilipSolich/shark-ci/ci-server/store"
	"github.com/google/uuid"
	"golang.org/x/exp/slog"
)

type OAuth2Handler struct {
	l        *slog.Logger
	s        store.Storer
	services service.Services
}

func NewOAuth2Handler(l *slog.Logger, s store.Storer, services service.Services) *OAuth2Handler {
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

	ctx := r.Context()
	uuid_state, err := uuid.Parse(state)
	if err != nil {
		http.Error(w, "incorrect state", http.StatusBadRequest)
		return
	}

	oauth2State, err := h.s.GetAndDeleteOAuth2State(ctx, uuid_state)
	if err != nil {
		http.Error(w, "incorrect state", http.StatusBadRequest)
		return
	}

	if !oauth2State.IsValid() {
		http.Error(w, "oauth2 state expired", http.StatusBadRequest)
		return
	}

	// Get Oauth2 token from auth service.
	config := srv.OAuth2Config()
	token, err := config.Exchange(ctx, code)
	if err != nil {
		h.l.Error("cannot get OAuth2 token", "err", err, "service", srv.Name(), "code", code)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Get or create new ServiceUser and new User if needed.
	serviceUser, err := srv.GetServiceUser(ctx, token)
	if err != nil {
		h.l.Error("service: cannot get service user", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var uID int64
	su, err := h.s.GetServiceUserByUniqueName(ctx, serviceUser.Service, serviceUser.Username)
	if err != nil {
		uID, err = h.s.CreateUserAndServiceUser(ctx, serviceUser)
		if err != nil {
			h.l.Error("store: cannot create user and service user", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		uID = su.UserID
		err = h.s.UpdateServiceUserToken(ctx, serviceUser, token)
		if err != nil {
			h.l.Warn("store: cannot update user OAuth2 token", "err", err)
			// TODO: Is old token still usable? Or should handler return here?
		}
	}

	// Store session.
	s, _ := session.Store.Get(r, "session")
	s.Values[session.SessionKey] = uID
	err = s.Save(r, w)
	if err != nil {
		h.l.Error("cannot save session", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}
