package handlers

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/shark-ci/shark-ci/ci-server/configs"
	"github.com/shark-ci/shark-ci/ci-server/db"
	"github.com/shark-ci/shark-ci/ci-server/services"
	"github.com/shark-ci/shark-ci/mq"
)

func EventHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := mux.Vars(r)
	serviceName, ok := params["service"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	service, ok := services.Services[serviceName]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	job, err := service.CreateJob(ctx, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if job == nil && err == nil {
		http.Error(w, "cannot handle this type of event", http.StatusNotImplemented)
		return
	}

	job, err = db.CreateJob(context.TODO(), job)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = mq.MQ.PublishJob(job)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	repo, err := db.GetRepoByID(ctx, job.Repo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	identity, err := repo.GetOwner(ctx)

	status := services.Status{
		State:       services.StatusPending,
		TargetURL:   job.TargetURL,
		Context:     configs.CIServer,
		Description: "Job in progress",
	}
	err = service.CreateStatus(ctx, identity, job, status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
