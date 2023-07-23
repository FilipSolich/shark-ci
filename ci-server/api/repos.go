package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"golang.org/x/exp/slog"

	"github.com/FilipSolich/shark-ci/ci-server/middleware"
	"github.com/FilipSolich/shark-ci/ci-server/service"
	"github.com/FilipSolich/shark-ci/ci-server/store"
	"github.com/FilipSolich/shark-ci/shared/model2"
	"github.com/gorilla/mux"
)

type RepoAPI struct {
	l        *slog.Logger
	s        store.Storer
	services service.Services
}

func NewRepoAPI(l *slog.Logger, s store.Storer, services service.Services) *RepoAPI {
	return &RepoAPI{
		l:        l,
		s:        s,
		services: services,
	}
}

func (api *RepoAPI) GetRepos(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := middleware.UserFromContext(ctx, w)
	if !ok {
		return
	}

	repos, err := api.s.GetReposByUser(ctx, user.ID)
	if err != nil {
		if repos == nil {
			api.l.Error("store: cannot get user repos", "err", err, "userID", user.ID)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		api.l.Warn("store: cannot get all user repos", "err", err, "userID", user.ID)
	}

	json.NewEncoder(w).Encode(repos)
}

func (api *RepoAPI) RefreshRepos(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := middleware.UserFromContext(ctx, w)
	if !ok {
		return
	}

	serviceUsers, err := api.s.GetServiceUsersByUser(ctx, user.ID)
	if err != nil {
		if serviceUsers == nil {
			api.l.Error("store: cannot get service users", "err", err, "userID", user.ID)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		api.l.Warn("store: cannot get all service users", "err", err, "userID", user.ID)
	}

	allRepos := []model2.Repo{}
	for _, serviceUser := range serviceUsers {
		srv, ok := api.services[serviceUser.Service]
		if !ok {
			api.l.Error("service: unknown service", "service", serviceUser.Service)
			continue
		}

		repos, err := srv.GetUsersRepos(ctx, &serviceUser)
		if err != nil {
			api.l.Error("service: cannot get user repositories from service", "err", err, "service", srv.Name())
			continue
		}

		allRepos = append(allRepos, repos...)
	}

	err = api.s.CreateOrUpdateRepos(ctx, allRepos)
	if err != nil {
		api.l.Error("store: cannot create or update repos", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (api *RepoAPI) CreateWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := middleware.UserFromContext(ctx, w)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	repoIDstring := vars["repoID"]
	repoID, err := strconv.ParseInt(repoIDstring, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	repo, err := api.s.GetRepo(ctx, repoID)
	if err != nil {
		api.l.Warn("store: cannot get repo", "err", err, "repoID", repoID)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	srv, ok := api.services[repo.Service]
	if !ok {
		api.l.Error("service: unknown service", "service", repo.Service)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	serviceUser, err := api.s.GetServiceUserByRepo(ctx, repo.ID)
	if err != nil {
		api.l.Error("store: cannot get service user by repo", "err", err, "repoID", repo.ID)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if serviceUser.UserID != user.ID {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	webhookID, err := srv.CreateWebhook(ctx, serviceUser, repo.Name)
	if err != nil {
		api.l.Error("service: cannot create webhook", "err", err, "repoID", repo.ID)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = api.s.UpdateRepoWebhook(ctx, repo.ID, webhookID)
	if err != nil {
		api.l.Error("store: cannot update repo webhook", "err", err, "repoID", repo.ID)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// TODO: Delete webhook.
