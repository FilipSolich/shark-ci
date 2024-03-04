package handlers

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/shark-ci/shark-ci/internal/server/service"
	"github.com/shark-ci/shark-ci/internal/server/session"
	"github.com/shark-ci/shark-ci/internal/server/store"
)

type OAuth2Handler struct {
	s        store.Storer
	services service.Services
}

func NewOAuth2Handler(s store.Storer, services service.Services) *OAuth2Handler {
	return &OAuth2Handler{
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
	if oauth2State.Expire.Before(time.Now()) {
		http.Error(w, "oauth2 state expired", http.StatusBadRequest)
		return
	}

	config := srv.OAuth2Config()
	token, err := config.Exchange(ctx, code)
	if err != nil {
		slog.Error("Cannot get OAuth2 token.", "service", srv.Name(), "code", code, "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	serviceUser, err := srv.GetServiceUser(ctx, token)
	if err != nil {
		slog.Error("Cannot get service user.", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	userID, err := h.s.GetUserID(ctx, serviceUser.Service, serviceUser.Username)
	if err != nil {
		userID, _, err = h.s.CreateUserAndServiceUser(ctx, serviceUser)
		if err != nil {
			slog.Error("Cannot create user and service user.", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	s, _ := session.Store.Get(r, "session")
	s.Values[session.SessionKey] = userID
	err = s.Save(r, w)
	if err != nil {
		slog.Error("Cannot save session.", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}
