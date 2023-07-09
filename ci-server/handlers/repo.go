package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/csrf"
	"go.uber.org/zap"

	"github.com/FilipSolich/shark-ci/ci-server/middlewares"
	"github.com/FilipSolich/shark-ci/ci-server/service"
	"github.com/FilipSolich/shark-ci/ci-server/store"
	"github.com/FilipSolich/shark-ci/ci-server/template"
	"github.com/FilipSolich/shark-ci/models"
)

type RepoHandler struct {
	l          *zap.SugaredLogger
	s          store.Storer
	serviceMap service.ServiceMap
}

func NewRepoHandler(l *zap.SugaredLogger, s store.Storer, serviceMap service.ServiceMap) *RepoHandler {
	return &RepoHandler{
		l:          l,
		s:          s,
		serviceMap: serviceMap,
	}
}

func (h *RepoHandler) HandleRepos(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := middlewares.UserFromContext(ctx, w)
	if !ok {
		return
	}

	serviceRepos := map[string]map[string][]*models.Repo{}
	for serviceName, srv := range h.serviceMap {
		identity, err := h.s.GetIdentityByUser(ctx, user, serviceName)
		if err != nil {
			h.l.Error(err)
			continue
		}

		// TODO: Bundle into repo finder and updater
		repos, err := srv.GetUsersRepos(r.Context(), identity)
		if err != nil {
			h.l.Error(err)
			continue
		}

		for _, repo := range repos {
			h.s.CreateRepo(r.Context(), repo)
		}

		registered, notRegistered := splitRepos(repos)
		serviceRepos[serviceName] = map[string][]*models.Repo{}
		serviceRepos[serviceName]["registered"] = registered
		serviceRepos[serviceName]["not_registered"] = notRegistered
	}

	template.RenderTemplate(w, "repos.html", map[string]any{
		csrf.TemplateTag: csrf.TemplateField(r),
		"ServicesRepos":  serviceRepos,
	})
}

func (h *RepoHandler) HandleRegisterRepo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	identity, repo, service, err := h.getInfoFromRequest(ctx, w, r)
	if err != nil {
		h.l.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	repo, err = service.CreateWebhook(ctx, identity, repo)
	if err != nil {
		h.l.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = h.s.UpdateRepoWebhook(ctx, repo)
	if err != nil {
		h.l.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/repositories", http.StatusFound)
}

func (h *RepoHandler) HandleUnregisterRepo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	identity, repo, srv, err := h.getInfoFromRequest(ctx, w, r)
	if err != nil {
		h.l.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = srv.DeleteWebhook(ctx, identity, repo)
	if err != nil {
		h.l.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = h.s.UpdateRepoWebhook(ctx, repo)
	if err != nil {
		h.l.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
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
		h.l.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	repo, err = service.ChangeWebhookState(ctx, identity, repo, active)
	if err != nil {
		h.l.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = h.s.UpdateRepoWebhook(ctx, repo)
	if err != nil {
		h.l.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/repositories", http.StatusFound)
}

func (h *RepoHandler) getInfoFromRequest(ctx context.Context, w http.ResponseWriter, r *http.Request) (*models.Identity, *models.Repo, service.ServiceManager, error) {
	user, ok := middlewares.UserFromContext(ctx, w)
	if !ok {
		return nil, nil, nil, errors.New("unauthorized user")
	}

	r.ParseForm()
	repo, err := h.s.GetRepo(ctx, r.Form.Get("repo_id"))
	if err != nil {
		return nil, nil, nil, err
	}

	srv, ok := h.serviceMap[repo.ServiceName]
	if !ok {
		return nil, nil, nil, fmt.Errorf("unknown service: %s", repo.ServiceName)
	}

	identity, err := h.s.GetIdentityByUser(ctx, user, repo.ServiceName)
	if err != nil {
		return nil, nil, nil, err
	}

	return identity, repo, srv, nil
}

func splitRepos(repos []*models.Repo) ([]*models.Repo, []*models.Repo) {
	registered := []*models.Repo{}
	notRegistered := []*models.Repo{}
	for _, repo := range repos {
		if repo.WebhookID == 0 {
			notRegistered = append(notRegistered, repo)
		} else {
			registered = append(registered, repo)
		}
	}
	return registered, notRegistered
}
