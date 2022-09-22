package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/FilipSolich/ci-server/configs"
	"github.com/FilipSolich/ci-server/db"
	"github.com/FilipSolich/ci-server/models"
	"github.com/FilipSolich/ci-server/services"
	"github.com/google/go-github/v47/github"
	"github.com/gorilla/csrf"
)

func ReposGetHandler(w http.ResponseWriter, r *http.Request, user *models.User) {
	ctx := context.Background()
	ghClient := services.GetGitHubClientByUser(ctx, user)
	repos, _, err := ghClient.Repositories.List(ctx, "", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	registeredWebhooks := []*models.Webhook{}
	notRegisteredRepos := []*github.Repository{}

	reposIDs := []int64{}
	for _, repo := range repos {
		reposIDs = append(reposIDs, repo.GetID())
	}
	db.DB.Where("repo_id IN ? AND service = ?", reposIDs, user.Service).Find(&registeredWebhooks)

	for _, repo := range repos {
		registered := false
		for _, webHook := range registeredWebhooks {
			if repo.GetID() == webHook.RepoID {
				registered = true
				break
			}
		}
		if !registered {
			notRegisteredRepos = append(notRegisteredRepos, repo)
		}
	}

	configs.RenderTemplate(w, "repos.html", map[string]any{
		csrf.TemplateTag:     csrf.TemplateField(r),
		"RegisteredWebhooks": registeredWebhooks,
		"NotRegisteredRepos": notRegisteredRepos,
	})

	//configs.RenderTemplate(w, "repos.html", struct {
	//	RegisteredWebhooks []*models.Webhook
	//	NotRegisteredRepos []*github.Repository
	//	CSRFToken          string
	//}{
	//	RegisteredWebhooks: registeredWebhooks,
	//	NotRegisteredRepos: notRegisteredRepos,
	//	CSRFToken:          "token",
	//})
}

func ReposPostHandler(w http.ResponseWriter, r *http.Request, user *models.User) {
	fmt.Println("Here")
}
