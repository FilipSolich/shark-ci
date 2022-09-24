package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/FilipSolich/ci-server/configs"
	"github.com/FilipSolich/ci-server/db"
	"github.com/FilipSolich/ci-server/middlewares"
	"github.com/FilipSolich/ci-server/models"
	"github.com/FilipSolich/ci-server/services"
	"github.com/google/go-github/v47/github"
	"github.com/gorilla/csrf"
)

func getRepoInfoFromRequest(r *http.Request) (int64, string, string, error) {
	r.ParseForm()
	repoIDString := r.Form.Get("repo_id")
	repoName := r.Form.Get("repo_name")
	repoFullName := r.Form.Get("repo_full_name")

	repoID, err := strconv.ParseInt(repoIDString, 10, 64)
	if err != nil {
		return 0, "", "", err
	}

	return repoID, repoName, repoFullName, nil
}

func Repos(w http.ResponseWriter, r *http.Request) {
	user, ok := middlewares.UserFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

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
}

func ReposRegister(w http.ResponseWriter, r *http.Request) {
	user, ok := middlewares.UserFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	repoID, repoName, repoFullName, err := getRepoInfoFromRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	result := db.DB.Where(&models.Webhook{Service: "github", RepoID: repoID})
	if result.RowsAffected > 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	hook, err := services.CreateWebhook(ctx, user, repoName, repoID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	hook.RepoFullName = repoFullName
	result = db.DB.Create(hook)
	http.Redirect(w, r, "/repositories", http.StatusFound)
}

func ReposUnregister(w http.ResponseWriter, r *http.Request) {
	user, ok := middlewares.UserFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	repoID, _, _, err := getRepoInfoFromRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	hook := &models.Webhook{}
	result := db.DB.Where(&models.Webhook{Service: "github", RepoID: repoID}).First(hook)
	if result.Error != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	err = services.DeleteWebhook(ctx, user, hook)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	db.DB.Delete(hook)
	http.Redirect(w, r, "/repositories", http.StatusFound)
}

func ReposActivate(w http.ResponseWriter, r *http.Request) {
	user, ok := middlewares.UserFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	repoID, _, _, err := getRepoInfoFromRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	hook := &models.Webhook{}
	result := db.DB.Where(&models.Webhook{Service: "github", RepoID: repoID}).First(hook)
	if result.Error != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	err = services.ActivateWebhook(ctx, user, hook)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	hook.Active = true
	db.DB.Save(hook)
	http.Redirect(w, r, "/repositories", http.StatusFound)
}

func ReposDeactivate(w http.ResponseWriter, r *http.Request) {
	user, ok := middlewares.UserFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Println(user.Username)
}
