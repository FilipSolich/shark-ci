package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/gorilla/mux"

	ciserver "github.com/FilipSolich/shark-ci/ci-server"
	"github.com/FilipSolich/shark-ci/ci-server/services"
	"github.com/FilipSolich/shark-ci/ci-server/store"
	"github.com/FilipSolich/shark-ci/message_queue"
)

type EventHandler struct {
	store      store.Storer
	mq         message_queue.MessageQueuer
	serviceMap services.ServiceMap
	serverName string
}

func NewEventHandler(store store.Storer, mq message_queue.MessageQueuer, serviceMap services.ServiceMap) *EventHandler {
	return &EventHandler{
		store:      store,
		mq:         mq,
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

	job, err := service.HandleEvent(r)
	if err != nil {
		if errors.Is(err, services.NoErrPingEvent) {
			w.WriteHeader(http.StatusNoContent)
			return
		} else if errors.Is(err, services.ErrEventNotSupported) {
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

	err = h.mq.SendJob(context.TODO(), job)
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

	status := services.NewStatus(services.StatusPending, job.TargetURL, ciserver.CIServer, "Job in progress")
	err = service.CreateStatus(ctx, identity, job, status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
