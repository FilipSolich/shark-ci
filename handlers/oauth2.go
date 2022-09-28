package handlers

import (
	"context"
	"net/http"

	"github.com/FilipSolich/ci-server/db"
	"github.com/FilipSolich/ci-server/models"
	"github.com/FilipSolich/ci-server/services"
	"github.com/FilipSolich/ci-server/sessions"
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

	var oauth2State models.OAuth2State
	result := db.DB.First(&oauth2State, models.OAuth2State{State: state})
	if result.Error != nil || !oauth2State.IsValid() {
		http.Error(w, "incorrect state", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	config := service.GetOAuth2Config()
	token, err := config.Exchange(ctx, code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userIdentity, err := service.GetOrCreateUserIdentity(ctx, token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = userIdentity.UpdateOAuth2Token(token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session, _ := sessions.Store.Get(r, "session")
	session.Values[sessions.SessionKey] = userIdentity.UserID
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}
