package handlers

import (
	"context"
	"net/http"

	"github.com/FilipSolich/ci-server/configs"
	"github.com/FilipSolich/ci-server/models"
	"github.com/FilipSolich/ci-server/services"
	"github.com/google/go-github/v47/github"
)

type templateData struct {
	Username           string
	NotRegisteredRepos []*github.Repository
}

func IndexHandler(w http.ResponseWriter, r *http.Request, user *models.User) {
	ctx := context.Background()
	ghClient := services.GetGitHubClientByUser(ctx, user)
	repos, _, err := ghClient.Repositories.List(ctx, "", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	registeredWebhooks := user.Webhooks
	var notRegisteredRepos []*github.Repository
	for _, repo := range repos {
		registered := false
		for _, webHook := range registeredWebhooks {
			if repo.GetID() == int64(webHook.RepoID) {
				registered = true
				break
			}
		}
		if !registered {
			notRegisteredRepos = append(notRegisteredRepos, repo)
		}
	}

	configs.Templates.ExecuteTemplate(w, "index.html", templateData{
		Username:           user.Username,
		NotRegisteredRepos: notRegisteredRepos,
	})
}
