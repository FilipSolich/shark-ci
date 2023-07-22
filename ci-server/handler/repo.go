package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/csrf"
	"golang.org/x/exp/slog"

	"github.com/FilipSolich/shark-ci/ci-server/middleware"
	"github.com/FilipSolich/shark-ci/ci-server/service"
	"github.com/FilipSolich/shark-ci/ci-server/store"
	"github.com/FilipSolich/shark-ci/ci-server/template"
	"github.com/FilipSolich/shark-ci/shared/model"
)

type RepoHandler struct {
	l        *slog.Logger
	s        store.Storer
	services service.Services
}

func NewRepoHandler(l *slog.Logger, s store.Storer, services service.Services) *RepoHandler {
	return &RepoHandler{
		l:        l,
		s:        s,
		services: services,
	}
}

func (h *RepoHandler) HandleRepos(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := middleware.UserFromContext(ctx, w)
	if !ok {
		return
	}

	serviceRepos := map[string]map[string][]*model.Repo{}
	for serviceName, srv := range h.services {
		serviceUser, err := h.s.GetServiceUserByUser(ctx, user, serviceName)
		if err != nil {
			h.l.Error("store: cannot get service user", "err", err)
			continue
		}

		// TODO: Bundle into repo finder and updater
		repos, err := srv.GetUsersRepos(r.Context(), serviceUser)
		if err != nil {
			h.l.Error("service: cannot get user repositorues from service", "err", err)
			continue
		}

		for _, repo := range repos {
			h.s.CreateRepo(r.Context(), repo)
		}

		registered, notRegistered := splitRepos(repos)
		serviceRepos[serviceName] = map[string][]*model.Repo{}
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
	serviceUser, repo, srv, err := h.getInfoFromRequest(ctx, w, r)
	if err != nil {
		h.l.Error("cannot get info from request", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	repo, err = srv.CreateWebhook(ctx, serviceUser, repo)
	if err != nil {
		h.l.Error("service: cannot create a webhook", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = h.s.UpdateRepoWebhook(ctx, repo)
	if err != nil {
		h.l.Error("store: cannot update a webhook", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/repositories", http.StatusFound)
}

func (h *RepoHandler) HandleUnregisterRepo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	serviceUser, repo, srv, err := h.getInfoFromRequest(ctx, w, r)
	if err != nil {
		h.l.Error("cannot get info from request", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = srv.DeleteWebhook(ctx, serviceUser, repo)
	if err != nil {
		h.l.Error("service: cannot delete webhook", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = h.s.UpdateRepoWebhook(ctx, repo)
	if err != nil {
		h.l.Error("store: cannot update webhook", "err", err)
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
	serviceUser, repo, srv, err := h.getInfoFromRequest(ctx, w, r)
	if err != nil {
		h.l.Error("cannot get infor from request", "err", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	repo, err = srv.ChangeWebhookState(ctx, serviceUser, repo, active)
	if err != nil {
		h.l.Error("service: cannot change a webhook state", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = h.s.UpdateRepoWebhook(ctx, repo)
	if err != nil {
		h.l.Error("store: cannot update a webhook", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/repositories", http.StatusFound)
}

func (h *RepoHandler) getInfoFromRequest(ctx context.Context, w http.ResponseWriter, r *http.Request) (*model.ServiceUser, *model.Repo, service.ServiceManager, error) {
	user, ok := middleware.UserFromContext(ctx, w)
	if !ok {
		return nil, nil, nil, errors.New("unauthorized user")
	}

	r.ParseForm()
	repo, err := h.s.GetRepo(ctx, r.Form.Get("repo_id"))
	if err != nil {
		return nil, nil, nil, err
	}

	srv, ok := h.services[repo.ServiceName]
	if !ok {
		return nil, nil, nil, fmt.Errorf("unknown service: %s", repo.ServiceName)
	}

	serviceUser, err := h.s.GetServiceUserByUser(ctx, user, repo.ServiceName)
	if err != nil {
		return nil, nil, nil, err
	}

	return serviceUser, repo, srv, nil
}

func splitRepos(repos []*model.Repo) ([]*model.Repo, []*model.Repo) {
	registered := []*model.Repo{}
	notRegistered := []*model.Repo{}
	for _, repo := range repos {
		if repo.WebhookID == 0 {
			notRegistered = append(notRegistered, repo)
		} else {
			registered = append(registered, repo)
		}
	}
	return registered, notRegistered
}
