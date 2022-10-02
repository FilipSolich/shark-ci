package handlers

import (
	"net/http"

	"github.com/FilipSolich/ci-server/configs"
	"github.com/FilipSolich/ci-server/db"
	"github.com/FilipSolich/ci-server/services"
	"github.com/gorilla/mux"
)

func EventHandler(w http.ResponseWriter, r *http.Request) {
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

	job, err := service.CreateJob(r.Context(), r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if job == nil && err == nil {
		http.Error(w, "cannot handle this type of event", http.StatusNotImplemented)
		return
	}

	err = db.DB.Save(job).Error
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	job.ReportStatusURL = "" // TODO: Generate URL for status reporting
	job.PublishLogsURL = ""  // TODO: Generate URL for logs publishinng
	err = db.DB.Save(job).Error
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: Puiblish to message queue and update state
	//	err = mq.MQ.PublishJob(job)
	//	if err != nil {
	//		fmt.Println(err)
	//	}

	status := services.Status{
		State:       services.StatusPending,
		TargetURL:   "", // TODO: Generate target URL
		Context:     configs.CIServer,
		Description: "", // TODO: Add description
	}
	err = service.UpdateStatus(r.Context(), status, job)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
