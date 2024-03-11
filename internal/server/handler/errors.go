package handler

import (
	"net/http"

	"github.com/shark-ci/shark-ci/templates"
	"golang.org/x/exp/slog"
)

func Error400(w http.ResponseWriter, msg string) {
	w.WriteHeader(http.StatusBadRequest)
	err := templates.Error400Tmpl.Execute(w, map[string]any{"Msg": msg})
	if err != nil {
		slog.Error("Cannot execute template.", "err", err)
	}
}

func Error404(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	if err := templates.Error404Tmpl.Execute(w, nil); err != nil {
		slog.Error("Cannot execute template.", "err", err)
	}
}

func Error5xx(w http.ResponseWriter, code int, msg string, err error) {
	slog.Error(msg, "err", err)
	w.WriteHeader(code)
	err = templates.Error5xxTmpl.Execute(w, nil)
	if err != nil {
		slog.Error("Cannot execute template.", "err", err)
	}
}
