package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
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

// TODO
func JobsReportStatusHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := mux.Vars(r)
	jobID, ok := params["id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	print(ctx, jobID)

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
