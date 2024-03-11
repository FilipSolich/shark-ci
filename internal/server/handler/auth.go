package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/shark-ci/shark-ci/internal/server/service"
	"github.com/shark-ci/shark-ci/internal/server/session"
	"github.com/shark-ci/shark-ci/internal/server/store"
	"github.com/shark-ci/shark-ci/internal/server/types"
	"github.com/shark-ci/shark-ci/templates"
	"golang.org/x/oauth2"
)

type AuthHandler struct {
	s        store.Storer
	services service.Services
}

func NewAuthHandler(s store.Storer, services service.Services) AuthHandler {
	return AuthHandler{
		s:        s,
		services: services,
	}
}

func (h AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	state, err := uuid.NewRandom()
	if err != nil {
		Error5xx(w, http.StatusInternalServerError, "Cannot generate UUID.", err)
		return
	}

	oauth2State := types.OAuth2State{
		State:  state,
		Expire: time.Now().Add(30 * time.Minute),
	}
	err = h.s.CreateOAuth2State(r.Context(), oauth2State)
	if err != nil {
		Error5xx(w, http.StatusInternalServerError, "Cannot create OAuth2 state.", err)
		return
	}

	data := map[string]string{}
	for _, s := range h.services {
		config := s.OAuth2Config()
		url := config.AuthCodeURL(oauth2State.State.String(), oauth2.AccessTypeOffline)
		data[s.Name()+"URL"] = url
	}

	err = templates.LoginTmpl.Execute(w, map[string]any{"URLs": data})
	if err != nil {
		Error5xx(w, http.StatusInternalServerError, "Cannot execute template.", err)
		return
	}
}

func (h AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := session.Store.Get(r, "session")
	session.Options.MaxAge = -1
	err := session.Save(r, w)
	if err != nil {
		Error5xx(w, http.StatusInternalServerError, "Cannot remove users session.", err)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (h AuthHandler) OAuth2Callback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	serviceName := r.URL.Query().Get("service")

	srv, ok := h.services[serviceName]
	if !ok {
		Error400(w, "Unknown OAuth2 provider"+serviceName)
		return
	}

	ctx := r.Context()
	uuid_state, err := uuid.Parse(state)
	if err != nil {
		Error400(w, "Invalid UUID")
		return
	}

	oauth2State, err := h.s.GetAndDeleteOAuth2State(ctx, uuid_state)
	if err != nil {
		Error400(w, "Invalid state")
		return
	}
	if oauth2State.Expire.Before(time.Now()) {
		Error400(w, "State expired")
		return
	}

	config := srv.OAuth2Config()
	token, err := config.Exchange(ctx, code)
	if err != nil {
		Error5xx(w, http.StatusInternalServerError, "Cannot get OAuth2 token.", fmt.Errorf("error exchanging OAuth2 code for token: %w", err))
		return
	}

	serviceUser, err := srv.GetServiceUser(ctx, token)
	if err != nil {
		Error5xx(w, http.StatusInternalServerError, "Cannot get service user.", err)
		return
	}

	userID, err := h.s.GetUserIDByServiceUser(ctx, serviceUser.Service, serviceUser.Username)
	if err != nil {
		userID, _, err = h.s.CreateUserAndServiceUser(ctx, serviceUser)
		if err != nil {
			Error5xx(w, http.StatusInternalServerError, "Cannot create user and service user.", err)
			return
		}
	}

	s, _ := session.Store.Get(r, "session")
	s.Values[session.SessionKey] = userID
	err = s.Save(r, w)
	if err != nil {
		Error5xx(w, http.StatusInternalServerError, "Cannot save users session.", err)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}
