package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"log/slog"

	"github.com/gorilla/mux"
	"github.com/shark-ci/shark-ci/ci-server/middleware"
	"github.com/shark-ci/shark-ci/ci-server/models"
	"github.com/shark-ci/shark-ci/ci-server/service"
	"github.com/shark-ci/shark-ci/ci-server/store"
	"github.com/shark-ci/shark-ci/ci-server/types"
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

	w.WriteHeader(http.StatusNoContent)
}

func (api *RepoAPI) CreateWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	info, srv, ok := api.repoWebhookChangeInfo(ctx, w, r)
	if !ok {
		return
	}

	webhookID, err := srv.CreateWebhook(ctx, &info.Token, info.RepoOwner, info.RepoName)
	if err != nil {
		api.l.Error("service: cannot create webhook", "err", err, "repoID", info.RepoID)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = api.s.UpdateRepoWebhook(ctx, info.RepoID, &webhookID)
	if err != nil {
		api.l.Error("store: cannot update repo webhook", "err", err, "repoID", info.RepoID)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (api *RepoAPI) DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	info, srv, ok := api.repoWebhookChangeInfo(ctx, w, r)
	if !ok {
		return
	}

	if info.WebhookID == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := srv.DeleteWebhook(ctx, &info.Token, info.RepoOwner, info.RepoName, *info.WebhookID)
	if err != nil {
		api.l.Error("service: cannot delete webhook", "err", err, "repoID", info.RepoID)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = api.s.UpdateRepoWebhook(ctx, info.RepoID, nil)
	if err != nil {
		api.l.Error("store: cannot update repo webhook", "err", err, "repoID", info.RepoID)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (api *RepoAPI) repoWebhookChangeInfo(ctx context.Context, w http.ResponseWriter, r *http.Request,
) (*types.RepoWebhookChangeInfo, service.ServiceManager, bool) {
	user, ok := middleware.UserFromContext(ctx, w)
	if !ok {
		return nil, nil, false
	}

	vars := mux.Vars(r)
	repoIDstring := vars["repoID"]
	repoID, err := strconv.ParseInt(repoIDstring, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil, nil, false
	}

	info, err := api.s.GetRepoWebhookChangeInfo(ctx, repoID)

	if info.UserID != user.ID {
		w.WriteHeader(http.StatusNotFound)
		return nil, nil, false
	}

	srv, ok := api.services[info.Service]
	if !ok {
		api.l.Error("service: unknown service", "service", info.Service)
		w.WriteHeader(http.StatusInternalServerError)
		return nil, nil, false
	}

	return info, srv, true
}
