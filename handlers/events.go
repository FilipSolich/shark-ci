package handlers

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/shark-ci/shark-ci/configs"
	"github.com/shark-ci/shark-ci/db"
	"github.com/shark-ci/shark-ci/mq"
	"github.com/shark-ci/shark-ci/services"
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

	job, err = db.CreateJob(ctx, job)
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
