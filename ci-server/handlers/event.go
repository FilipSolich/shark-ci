package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/shark-ci/shark-ci/ci-server/configs"
	"github.com/shark-ci/shark-ci/ci-server/services"
	"github.com/shark-ci/shark-ci/ci-server/store"
	"github.com/shark-ci/shark-ci/mq"
)

type EventHandler struct {
	store      store.Storer
	serviceMap services.ServiceMap
}

func NewEventHandler(store store.Storer, serviceMap services.ServiceMap) *EventHandler {
	return &EventHandler{
		store:      store,
		serviceMap: serviceMap,
	}
}

func (h *EventHandler) HandleEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := mux.Vars(r)
	serviceName, ok := params["service"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	service, ok := h.serviceMap[serviceName]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	job, err := service.CreateJob(ctx, r)
	if err != nil {
		if errors.Is(err, services.ErrEventNotSupported) {
			http.Error(w, "cannot handle this type of event", http.StatusNotImplemented)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	err = h.store.CreateJob(context.TODO(), job)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = mq.MQ.PublishJob(job)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	repo, err := h.store.GetRepo(ctx, job.RepoID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	identity, err := h.store.GetIdentityByRepo(ctx, repo)

	status := services.NewStatus(services.StatusPending, job.TargetURL, configs.CIServer, "Job in progress")
	err = service.CreateStatus(ctx, identity, job, status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
