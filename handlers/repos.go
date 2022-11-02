package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/csrf"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/FilipSolich/ci-server/configs"
	"github.com/FilipSolich/ci-server/db"
	"github.com/FilipSolich/ci-server/middlewares"
	"github.com/FilipSolich/ci-server/models"
	"github.com/FilipSolich/ci-server/services"
)

func ReposHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := middlewares.UserFromContext(r.Context(), w)
	if !ok {
		return
	}

	serviceRepos := map[string]map[string][]*models.Repository{}
	for serviceName, service := range services.Services {
		repos, err := service.GetUsersRepos(r.Context(), user)
		if err != nil {
			log.Print(err)
			continue
		}

		registered, notRegistered := splitRepos(repos)
		serviceRepos[serviceName] = map[string][]*models.Repository{}
		serviceRepos[serviceName]["registered"] = registered
		serviceRepos[serviceName]["not_registered"] = notRegistered
	}

	configs.RenderTemplate(w, "repos.html", map[string]any{
		csrf.TemplateTag: csrf.TemplateField(r),
		"ServicesRepos":  serviceRepos,
	})
}

func ReposRegisterHandler(w http.ResponseWriter, r *http.Request) {
	user, repo, service, err := getUserRepoService(w, r)
	if err != nil {
		return
	}

	webhook, err := service.CreateWebhook(r.Context(), user, repo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = db.DB.Save(webhook).Error
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/repositories", http.StatusFound)
}

func ReposUnregisterHandler(w http.ResponseWriter, r *http.Request) {
	user, repo, service, err := getUserRepoService(w, r)
	if err != nil {
		return
	}

	err = service.DeleteWebhook(r.Context(), user, repo, &repo.Webhook)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = db.DB.Delete(&repo.Webhook).Error
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/repositories", http.StatusFound)
}

func ReposActivateHandler(w http.ResponseWriter, r *http.Request) {
	user, repo, service, err := getUserRepoService(w, r)
	if err != nil {
		return
	}

	hook, err := service.ActivateWebhook(r.Context(), user, repo, &repo.Webhook)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = db.DB.Save(hook).Error
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/repositories", http.StatusFound)
}

func ReposDeactivateHandler(w http.ResponseWriter, r *http.Request) {
	user, repo, service, err := getUserRepoService(w, r)
	if err != nil {
		return
	}

	hook, err := service.DeactivateWebhook(r.Context(), user, repo, &repo.Webhook)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = db.DB.Save(hook).Error
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/repositories", http.StatusFound)
}

func splitRepos(repos []*models.Repository) ([]*models.Repository, []*models.Repository) {
	registered := []*models.Repository{}
	notRegistered := []*models.Repository{}
	for _, repo := range repos {
		if repo.Webhook.ID == 0 {
			notRegistered = append(notRegistered, repo)
		} else {
			registered = append(registered, repo)
		}
	}
	return registered, notRegistered
}

func getUserRepoService(w http.ResponseWriter, r *http.Request) (*models.User, *models.Repository, services.ServiceManager, error) {
	user, ok := middlewares.UserFromContext(r.Context(), w)
	if !ok {
		return nil, nil, nil, errors.New("unauthorized user")
	}

	repo, err := getRepoFromRequest(w, r)
	if err != nil {
		return nil, nil, nil, err
	}

	service, ok := services.Services[repo.ServiceName]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return nil, nil, nil, fmt.Errorf("unknown service: %s", repo.ServiceName)
	}

	return user, repo, service, nil
}

func getRepoFromRequest(w http.ResponseWriter, r *http.Request) (*models.Repository, error) {
	r.ParseForm()
	repoID := r.Form.Get("repo_id")

	var repo models.Repository
	err := db.DB.Preload(clause.Associations).First(&repo, repoID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			w.WriteHeader(http.StatusBadRequest)
			return nil, fmt.Errorf("incorrect repository ID: %s", repoID)
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	return &repo, nil
}
