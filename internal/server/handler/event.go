package handler

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"golang.org/x/exp/slog"

	"github.com/shark-ci/shark-ci/internal/messagequeue"
	"github.com/shark-ci/shark-ci/internal/server/service"
	"github.com/shark-ci/shark-ci/internal/server/store"
	"github.com/shark-ci/shark-ci/internal/types"
)

type EventHandler struct {
	s        store.Storer
	mq       messagequeue.MessageQueuer
	services service.Services
}

func NewEventHandler(s store.Storer, mq messagequeue.MessageQueuer, services service.Services) *EventHandler {
	return &EventHandler{
		s:        s,
		mq:       mq,
		services: services,
	}
}

func (h *EventHandler) HandleEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	serviceName := mux.Vars(r)["service"]
	srv, ok := h.services[serviceName]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	pipeline, err := srv.HandleEvent(ctx, w, r)
	if err != nil {
		if errors.Is(err, service.ErrEventNotSupported) {
			http.Error(w, "cannot handle this type of event", http.StatusNotImplemented)
		} else {
			slog.Error("Cannot handle event.", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	if pipeline == nil {
		// TODO: handle rest of event somewhere else so HandleEvent should return pipeline but just error
		return
	}

	_, err = h.s.CreatePipeline(ctx, pipeline)
	if err != nil {
		slog.Error("Cannot create pipeline", "err", err)
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
		State:       types.Pending,
		TargetURL:   pipeline.URL,
		Context:     "Shark CI",
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
