package handlers

import (
	"net/http"
)

func EventHandler(w http.ResponseWriter, r *http.Request) {
	//params := mux.Vars(r)
	//serviceName, ok := params["service"]
	//if !ok {
	//	w.WriteHeader(http.StatusBadRequest)
	//	return
	//}

	//service, ok := services.Services[serviceName]
	//if !ok {
	//	w.WriteHeader(http.StatusBadRequest)
	//	return
	//}

	//job, err := service.CreateJob(r.Context(), r)
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//	return
	//}
	//if job == nil && err == nil {
	//	http.Error(w, "cannot handle this type of event", http.StatusNotImplemented)
	//	return
	//}

	//err = db.DB.Save(job).Error
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//	return
	//}

	//err = job.CreateJobURLs()
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//	return
	//}

	//// TODO: Puiblish to message queue and update state
	////	err = mq.MQ.PublishJob(job)
	////	if err != nil {
	////		fmt.Println(err)
	////	}

	//status := services.Status{
	//	State:       services.StatusPending,
	//	TargetURL:   job.TargetURL,
	//	Context:     configs.CIServer,
	//	Description: "Job in progress",
	//}
	//// TODO: Change blank user on actual user
	//err = service.UpdateStatus(r.Context(), &models.User{}, status, job)
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//	return
	//}
}
