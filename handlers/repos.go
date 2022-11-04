package handlers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/csrf"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/FilipSolich/ci-server/configs"
	"github.com/FilipSolich/ci-server/db"
	"github.com/FilipSolich/ci-server/middlewares"
	"github.com/FilipSolich/ci-server/services"
)

func ReposHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := middlewares.UserFromContext(ctx, w)
	if !ok {
		return
	}

	serviceRepos := map[string]map[string][]*db.Repo{}
	for serviceName, service := range services.Services {
		identity, err := db.GetIdentityByService(ctx, user, serviceName)
		if err != nil {
			log.Print(err)
			continue
		}

		repos, err := service.GetUsersRepos(r.Context(), identity)
		if err != nil {
			log.Print(err)
			continue
		}

		registered, notRegistered := splitRepos(repos)
		serviceRepos[serviceName] = map[string][]*db.Repo{}
		serviceRepos[serviceName]["registered"] = registered
		serviceRepos[serviceName]["not_registered"] = notRegistered
	}

	configs.RenderTemplate(w, "repos.html", map[string]any{
		csrf.TemplateTag: csrf.TemplateField(r),
		"ServicesRepos":  serviceRepos,
	})
}

func ReposRegisterHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	identity, repo, service, err := getInfoFromRequest(ctx, w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	webhook, err := service.CreateWebhook(ctx, identity, repo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = repo.UpdateWebhook(ctx, webhook)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/repositories", http.StatusFound)
}

func ReposUnregisterHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	identity, repo, service, err := getInfoFromRequest(ctx, w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = service.DeleteWebhook(ctx, identity, repo, &repo.Webhook)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = repo.DeleteWebhook(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/repositories", http.StatusFound)
}

func ReposActivateHandler(w http.ResponseWriter, r *http.Request) {
	changeRepoState(w, r, true)
}

func ReposDeactivateHandler(w http.ResponseWriter, r *http.Request) {
	changeRepoState(w, r, false)
}

func changeRepoState(w http.ResponseWriter, r *http.Request, active bool) {
	ctx := r.Context()
	identity, repo, service, err := getInfoFromRequest(ctx, w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hook, err := service.ChangeWebhookState(ctx, identity, repo, &repo.Webhook, active)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = repo.UpdateWebhook(ctx, hook)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/repositories", http.StatusFound)
}

func splitRepos(repos []*db.Repo) ([]*db.Repo, []*db.Repo) {
	registered := []*db.Repo{}
	notRegistered := []*db.Repo{}
	for _, repo := range repos {
		if repo.Webhook.WebhookID == 0 {
			notRegistered = append(notRegistered, repo)
		} else {
			registered = append(registered, repo)
		}
	}
	return registered, notRegistered
}

func getInfoFromRequest(ctx context.Context, w http.ResponseWriter, r *http.Request) (*db.Identity, *db.Repo, services.ServiceManager, error) {
	user, ok := middlewares.UserFromContext(ctx, w)
	if !ok {
		return nil, nil, nil, errors.New("unauthorized user")
	}

	r.ParseForm()
	repoID, err := primitive.ObjectIDFromHex(r.Form.Get("repo_id"))
	if err != nil {
		return nil, nil, nil, err
	}

	repo, err := db.GetRepoFromID(ctx, repoID)
	if err != nil {
		return nil, nil, nil, err
	}

	service, ok := services.Services[repo.ServiceName]
	if !ok {
		return nil, nil, nil, fmt.Errorf("unknown service: %s", repo.ServiceName)
	}

	identity, err := db.GetIdentityByService(ctx, user, repo.ServiceName)
	if err != nil {
		return nil, nil, nil, err
	}

	return identity, repo, service, nil
}
