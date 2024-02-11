package handlers

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"golang.org/x/exp/slog"

	"github.com/shark-ci/shark-ci/internal/message_queue"
	"github.com/shark-ci/shark-ci/internal/server/service"
	"github.com/shark-ci/shark-ci/internal/server/store"
	"github.com/shark-ci/shark-ci/internal/types"
)

type EventHandler struct {
	s        store.Storer
	mq       message_queue.MessageQueuer
	services service.Services
}

func NewEventHandler(s store.Storer, mq message_queue.MessageQueuer, services service.Services) *EventHandler {
	return &EventHandler{
		s:        s,
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

	pipeline, err := srv.HandleEvent(ctx, r)
	if err != nil {
		if errors.Is(err, service.NoErrPingEvent) {
			w.Write([]byte("pong"))
		} else if errors.Is(err, service.ErrEventNotSupported) {
			http.Error(w, "cannot handle this type of event", http.StatusNotImplemented)
		} else {
			slog.Error("service: cannot hadle event", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	_, err = h.s.CreatePipeline(ctx, pipeline)
	if err != nil {
		slog.Error("store: cannot create pipeline", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	info, err := h.s.GetPipelineCreationInfo(ctx, pipeline.RepoID)
	if err != nil {
		slog.Error("store: cannot get service user", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	work := types.Work{
		Pipeline: *pipeline,
		Token:    info.Token,
	}
	err = h.mq.SendWork(ctx, work)
	if err != nil {
		slog.Error("message queue: cannot send work", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	status := service.Status{
		State:       service.StatusPending,
		TargetURL:   pipeline.URL,
		Context:     pipeline.Context, // TODO: Get context from somewhere
		Description: "Pipeline is pending",
	}
	err = srv.CreateStatus(ctx, &info.Token, info.Username, info.RepoName, pipeline.CommitSHA, status)
	if err != nil {
		slog.Error("cannot create status", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
