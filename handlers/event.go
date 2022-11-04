package handlers

import (
	"net/http"

	"github.com/FilipSolich/ci-server/db"
	"github.com/FilipSolich/ci-server/services"
	"github.com/gorilla/mux"
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

	// TODO: Puiblish to message queue and update state
	//	err = mq.MQ.PublishJob(job)
	//	if err != nil {
	//		fmt.Println(err)
	//	}

	//status := services.Status{
	//	State:       services.StatusPending,
	//	TargetURL:   job.TargetURL,
	//	Context:     configs.CIServer,
	//	Description: "Job in progress",
	//}
	// TODO: Change blank user on actual user
	//err = service.UpdateStatus(r.Context(), &models.User{}, status, job)
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//	return
	//}
}
