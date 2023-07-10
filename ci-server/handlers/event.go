package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	ciserver "github.com/FilipSolich/shark-ci/ci-server"
	"github.com/FilipSolich/shark-ci/ci-server/service"
	"github.com/FilipSolich/shark-ci/ci-server/store"
	"github.com/FilipSolich/shark-ci/shared/message_queue"
)

type EventHandler struct {
	loger    *zap.SugaredLogger
	store    store.Storer
	mq       message_queue.MessageQueuer
	services service.Services
}

func NewEventHandler(l *zap.SugaredLogger, s store.Storer, mq message_queue.MessageQueuer, services service.Services) *EventHandler {
	return &EventHandler{
		loger:    l,
		store:    s,
		mq:       mq,
		services: services,
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

	srv, ok := h.services[serviceName]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	job, err := srv.HandleEvent(r)
	if err != nil {
		if errors.Is(err, service.NoErrPingEvent) {
			w.Write([]byte("pong"))
			return
		} else if errors.Is(err, service.ErrEventNotSupported) {
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

	serviceUser, err := h.store.GetServiceUserByRepo(ctx, repo)

	status := service.NewStatus(service.StatusPending, job.TargetURL, ciserver.CIServer, "Job in progress")
	err = srv.CreateStatus(ctx, serviceUser, job, status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
