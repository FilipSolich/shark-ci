package handlers

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/shark-ci/shark-ci/ci-server/configs"
	"github.com/shark-ci/shark-ci/ci-server/db"
	"github.com/shark-ci/shark-ci/ci-server/middlewares"
	"github.com/shark-ci/shark-ci/ci-server/services"
	"github.com/shark-ci/shark-ci/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const logsFolder = "joblogs"

func JobsTargetHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := middlewares.UserFromContext(ctx, w)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	job, err := getJobFromRequest(r)
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

	if !user.IsUserIdentity(ctx, identity) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	logName := job.ID.Hex() + ".log"
	w.Header().Set("Content-Disposition", "attachment; filename="+logName)
	file, err := ioutil.ReadFile(filepath.Join(logsFolder, logName))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(file)
}

func JobsReportStatusHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	job, err := getJobFromRequest(r)
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

func JobsPublishLogsHandler(w http.ResponseWriter, r *http.Request) {
	job, err := getJobFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("log")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	err = os.Mkdir(logsFolder, os.ModePerm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	newLogFilename := filepath.Join(logsFolder, job.ID.Hex()+".log")
	newLog, err := os.OpenFile(newLogFilename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer newLog.Close()

	io.Copy(newLog, file)
}

func getJobFromRequest(r *http.Request) (*models.Job, error) {
	ctx := r.Context()
	params := mux.Vars(r)
	jobID, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		return nil, err
	}

	job, err := db.GetJobByID(ctx, jobID)
	return job, err
}
