package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"golang.org/x/exp/slog"

	"github.com/FilipSolich/shark-ci/ci-server/middleware"
	"github.com/FilipSolich/shark-ci/ci-server/models"
	"github.com/FilipSolich/shark-ci/ci-server/service"
	"github.com/FilipSolich/shark-ci/ci-server/store"
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

	allRepos := []models.Repo{}
	for _, serviceUser := range serviceUsers {
		srv, ok := api.services[serviceUser.Service]
		if !ok {
			api.l.Error("service: unknown service", "service", serviceUser.Service)
			continue
		}

		repos, err := srv.GetUsersRepos(ctx, serviceUser.Token(), serviceUser.ID)
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
	repo, serviceUser, srv, ok := api.repositoryInfo(ctx, w, r)
	if !ok {
		return
	}

	webhookID, err := srv.CreateWebhook(ctx, serviceUser.Token(), serviceUser.Username, repo.Name)
	if err != nil {
		api.l.Error("service: cannot create webhook", "err", err, "repoID", repo.ID)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = api.s.UpdateRepoWebhook(ctx, repo.ID, &webhookID)
	if err != nil {
		api.l.Error("store: cannot update repo webhook", "err", err, "repoID", repo.ID)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (api *RepoAPI) DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	repo, serviceUser, srv, ok := api.repositoryInfo(ctx, w, r)
	if !ok {
		return
	}

	err := srv.DeleteWebhook(ctx, serviceUser.Token(), serviceUser.Username, repo.Name, *repo.WebhookID)
	if err != nil {
		api.l.Error("service: cannot delete webhook", "err", err, "repoID", repo.ID)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = api.s.UpdateRepoWebhook(ctx, repo.ID, nil)
	if err != nil {
		api.l.Error("store: cannot update repo webhook", "err", err, "repoID", repo.ID)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (api *RepoAPI) repositoryInfo(ctx context.Context, w http.ResponseWriter, r *http.Request) (*models.Repo, *models.ServiceUser, service.ServiceManager, bool) {
	user, ok := middleware.UserFromContext(ctx, w)
	if !ok {
		return nil, nil, nil, false
	}

	vars := mux.Vars(r)
	repoIDstring := vars["repoID"]
	repoID, err := strconv.ParseInt(repoIDstring, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil, nil, nil, false
	}

	repo, err := api.s.GetRepo(ctx, repoID)
	if err != nil {
		api.l.Warn("store: cannot get repo", "err", err, "repoID", repoID)
		w.WriteHeader(http.StatusNotFound)
		return nil, nil, nil, false
	}

	srv, ok := api.services[repo.Service]
	if !ok {
		api.l.Error("service: unknown service", "service", repo.Service)
		w.WriteHeader(http.StatusInternalServerError)
		return nil, nil, nil, false
	}

	serviceUser, err := api.s.GetServiceUserByUserAndService(ctx, user.ID, repo.Service)
	if err != nil {
		api.l.Error("store: cannot get service user by user and service", "err", err, "user", user.ID, "service", repo.Service)
		w.WriteHeader(http.StatusInternalServerError)
		return nil, nil, nil, false
	}

	return repo, serviceUser, srv, true
}
