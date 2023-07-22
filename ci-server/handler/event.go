package handler

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"golang.org/x/exp/slog"
	"golang.org/x/oauth2"

	ciserver "github.com/FilipSolich/shark-ci/ci-server"
	"github.com/FilipSolich/shark-ci/ci-server/service"
	"github.com/FilipSolich/shark-ci/ci-server/store"
	"github.com/FilipSolich/shark-ci/shared/message_queue"
	"github.com/FilipSolich/shark-ci/shared/types"
)

type EventHandler struct {
	l        *slog.Logger
	s        store.Storer
	mq       message_queue.MessageQueuer
	services service.Services
}

func NewEventHandler(l *slog.Logger, s store.Storer, mq message_queue.MessageQueuer, services service.Services) *EventHandler {
	return &EventHandler{
		l:        l,
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

	pipeline, err := srv.HandleEvent(r)
	if err != nil {
		if errors.Is(err, service.NoErrPingEvent) {
			w.Write([]byte("pong"))
		} else if errors.Is(err, service.ErrEventNotSupported) {
			http.Error(w, "cannot handle this type of event", http.StatusNotImplemented)
		} else {
			h.l.Error("service: cannot hadle event", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	err = h.s.CreatePipeline(ctx, pipeline)
	if err != nil {
		h.l.Error("store: cannot create pipeline", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	serviceUser, err := h.s.GetServiceUserByRepo(ctx, pipeline.RepoID)
	if err != nil {
		h.l.Error("store: cannot get service user", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	work := types.Work{
		Pipeline: *pipeline,
		Token: oauth2.Token{
			AccessToken:  serviceUser.AccessToken,
			TokenType:    serviceUser.TokenType,
			RefreshToken: serviceUser.RefreshToken,
			Expiry:       serviceUser.TokenExpire,
		},
	}
	err = h.mq.SendWork(ctx, work)
	if err != nil {
		h.l.Error("message queue: cannot send work", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	repoName, err := h.s.GetRepoName(ctx, pipeline.RepoID)
	if err != nil {
		h.l.Error("store: cannot get repo name", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	status := service.Status{
		State:       service.StatusPending,
		TargetURL:   pipeline.TargetURL,
		Context:     ciserver.CIServer,
		Description: "Job in progress",
	}
	err = srv.CreateStatus(ctx, serviceUser, repoName, pipeline.CommitSHA, status)
	if err != nil {
		h.l.Error("cannot create status", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
