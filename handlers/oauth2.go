package handlers

import (
	"context"
	"net/http"

	"github.com/FilipSolich/ci-server/models"
	"github.com/FilipSolich/ci-server/services"
	"github.com/FilipSolich/ci-server/sessions"
	"github.com/google/go-github/v47/github"
)

func OAuth2CallbackHandler(w http.ResponseWriter, r *http.Request) {
	codes, ok := r.URL.Query()["code"]
	if !ok || len(codes) != 1 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	code := codes[0]

	servicesNames, ok := r.URL.Query()["service"]
	if !ok || len(servicesNames) != 1 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	serviceName := servicesNames[0]

	ctx := context.Background()
	token, err := services.GitHubOAut2Config.Exchange(ctx, code)
	if err != nil {
		panic(err)
	}

	client := services.GitHubOAut2Config.Client(ctx, token)
	ghClient := github.NewClient(client)

	userInfo, _, err := ghClient.Users.Get(ctx, "")
	if err != nil {
		panic(err)
	}
	userName := userInfo.GetLogin()

	u := &models.User{
		Username: userName,
		Service:  serviceName,
	}
	t := &models.OAuth2Token{
		Token: *token,
	}
	user, err := models.GetOrCreateUser(u, t)
	if err != nil {
		panic(err)
	}

	session, _ := sessions.Store.Get(r, "session")
	session.Values["id"] = user.ID
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}
