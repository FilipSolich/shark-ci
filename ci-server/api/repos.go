package api

import (
	"encoding/json"
	"net/http"

	"golang.org/x/exp/slog"

	"github.com/FilipSolich/shark-ci/ci-server/middleware"
	"github.com/FilipSolich/shark-ci/ci-server/service"
	"github.com/FilipSolich/shark-ci/ci-server/store"
	"github.com/FilipSolich/shark-ci/shared/model2"
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
