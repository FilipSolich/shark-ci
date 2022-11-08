package handlers

import (
	"net/http"

	"github.com/FilipSolich/ci-server/configs"
	"github.com/FilipSolich/ci-server/db"
	"github.com/FilipSolich/ci-server/services"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TODO
func JobsTargetHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := mux.Vars(r)
	jobID, ok := params["id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	print(ctx, jobID)
}

func JobsReportStatusHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := mux.Vars(r)
	jobID, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	job, err := db.GetJobByID(ctx, jobID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	repo, err := db.GetRepoByID(ctx, job.Repo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	identity, err := repo.GetOwner(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	r.ParseForm()
	statusType := r.Form.Get("statusType")
	description := r.Form.Get("description")

	statusState, ok := services.StatusStateMap[statusType]
	if !ok {
		http.Error(w, "unknow status: "+statusType, http.StatusBadRequest)
		return
	}

	service, ok := services.Services[repo.ServiceName]
	if !ok {
		http.Error(w, "unknow service for repo: "+repo.FullName, http.StatusInternalServerError)
		return
	}

	status := services.Status{
		State:       statusState,
		TargetURL:   job.TargetURL,
		Context:     configs.CIServer,
		Description: description,
	}
	err = service.CreateStatus(ctx, identity, job, status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// TODO
func JobsPublishLogsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := mux.Vars(r)
	jobID, ok := params["id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	print(ctx, jobID)
}
