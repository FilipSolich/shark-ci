package handlers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/csrf"

	"github.com/shark-ci/shark-ci/ci-server/configs"
	"github.com/shark-ci/shark-ci/ci-server/middlewares"
	"github.com/shark-ci/shark-ci/ci-server/services"
	"github.com/shark-ci/shark-ci/ci-server/store"
)

type RepoHandler struct {
	store store.Storer
}

func NewRepoHandler(store store.Storer) *RepoHandler {
	return &RepoHandler{
		store: store,
	}
}

func (h *RepoHandler) HandleRepos(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := middlewares.UserFromContext(ctx, w)
	if !ok {
		return
	}

	serviceRepos := map[string]map[string][]*models.Repo{}
	for serviceName, service := range services.Services {
		identity, err := models.GetIdentityByUser(ctx, user, serviceName)
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
		serviceRepos[serviceName] = map[string][]*models.Repo{}
		serviceRepos[serviceName]["registered"] = registered
		serviceRepos[serviceName]["not_registered"] = notRegistered
	}

	configs.RenderTemplate(w, "repos.html", map[string]any{
		csrf.TemplateTag: csrf.TemplateField(r),
		"ServicesRepos":  serviceRepos,
	})
}

func (h *RepoHandler) HandleRegisterRepo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	identity, repo, service, err := h.getInfoFromRequest(ctx, w, r)
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

func (h *RepoHandler) HandleUnregisterRepo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	identity, repo, service, err := h.getInfoFromRequest(ctx, w, r)
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

func (h *RepoHandler) HandleActivateRepo(w http.ResponseWriter, r *http.Request) {
	h.changeRepoState(w, r, true)
}

func (h *RepoHandler) HandleDeactivateRepo(w http.ResponseWriter, r *http.Request) {
	h.changeRepoState(w, r, false)
}

func (h *RepoHandler) changeRepoState(w http.ResponseWriter, r *http.Request, active bool) {
	ctx := r.Context()
	identity, repo, service, err := h.getInfoFromRequest(ctx, w, r)
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

func (h *RepoHandler) getInfoFromRequest(ctx context.Context, w http.ResponseWriter, r *http.Request) (*models.Identity, *models.Repo, services.ServiceManager, error) {
	user, ok := middlewares.UserFromContext(ctx, w)
	if !ok {
		return nil, nil, nil, errors.New("unauthorized user")
	}

	r.ParseForm()
	repo, err := h.store.GetRepo(ctx, r.Form.Get("repo_id"))
	if err != nil {
		return nil, nil, nil, err
	}

	service, ok := services.Services[repo.ServiceName]
	if !ok {
		return nil, nil, nil, fmt.Errorf("unknown service: %s", repo.ServiceName)
	}

	identity, err := models.GetIdentityByUser(ctx, user, repo.ServiceName)
	if err != nil {
		return nil, nil, nil, err
	}

	return identity, repo, service, nil
}

func splitRepos(repos []*models.Repo) ([]*models.Repo, []*models.Repo) {
	registered := []*models.Repo{}
	notRegistered := []*models.Repo{}
	for _, repo := range repos {
		if repo.Webhook.WebhookID == 0 {
			notRegistered = append(notRegistered, repo)
		} else {
			registered = append(registered, repo)
		}
	}
	return registered, notRegistered
}
