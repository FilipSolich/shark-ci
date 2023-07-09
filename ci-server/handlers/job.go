package handlers

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	ciserver "github.com/FilipSolich/shark-ci/ci-server"
	"github.com/FilipSolich/shark-ci/ci-server/middleware"
	"github.com/FilipSolich/shark-ci/ci-server/service"
	"github.com/FilipSolich/shark-ci/ci-server/store"
	"github.com/FilipSolich/shark-ci/models"
	"github.com/gorilla/mux"
)

const logsFolder = "joblogs"

type JobHandler struct {
	store      store.Storer
	serviceMap service.ServiceMap
}

func NewJobHandler(store store.Storer, serviceMap service.ServiceMap) *JobHandler {
	return &JobHandler{
		store:      store,
		serviceMap: serviceMap,
	}
}

func (h *JobHandler) HandleJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, ok := middleware.UserFromContext(ctx, w)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	job, err := h.getJobFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: All this err.Error() may leak sensitive information to unauthorized user CHANGE IT!!!
	repo, err := h.store.GetRepo(ctx, job.RepoID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	identity, err := h.store.GetIdentityByRepo(ctx, repo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	repoOwner, err := h.store.GetUserByIdentity(ctx, identity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if user.ID != repoOwner.ID {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	logName := job.ID + ".log"
	w.Header().Set("Content-Disposition", "attachment; filename="+logName)
	file, err := ioutil.ReadFile(filepath.Join(logsFolder, logName))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(file)
}

func (h *JobHandler) HandleStatusReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	job, err := h.getJobFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	repo, err := h.store.GetRepo(ctx, job.RepoID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	identity, err := h.store.GetIdentityByRepo(ctx, repo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	r.ParseForm()
	statusType := r.Form.Get("statusType")
	description := r.Form.Get("description")

	statusState, ok := service.StatusStateMap[statusType]
	if !ok {
		http.Error(w, "unknow status: "+statusType, http.StatusBadRequest)
		return
	}

	srv, ok := h.serviceMap[repo.ServiceName]
	if !ok {
		http.Error(w, "unknow service for repo: "+repo.FullName, http.StatusInternalServerError)
		return
	}

	status := service.NewStatus(statusState, job.TargetURL, ciserver.CIServer, description)
	err = srv.CreateStatus(ctx, identity, job, status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *JobHandler) HandleLogReport(w http.ResponseWriter, r *http.Request) {
	job, err := h.getJobFromRequest(r)
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
	newLogFilename := filepath.Join(logsFolder, job.ID+".log")
	newLog, err := os.OpenFile(newLogFilename, os.O_WRONLY|os.O_CREATE, 0o666)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer newLog.Close()

	io.Copy(newLog, file)
}

func (h *JobHandler) getJobFromRequest(r *http.Request) (*models.Job, error) {
	ctx := r.Context()
	params := mux.Vars(r)
	jobID, ok := params["id"]
	if !ok {
		return nil, errors.New("invalid job ID")
	}

	job, err := h.store.GetJob(ctx, jobID)
	return job, err
}
