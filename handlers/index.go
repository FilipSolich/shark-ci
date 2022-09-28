package handlers

import (
	"net/http"

	"github.com/FilipSolich/ci-server/configs"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	configs.RenderTemplate(w, "index.html", nil)
}
