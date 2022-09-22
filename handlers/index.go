package handlers

import (
	"net/http"

	"github.com/FilipSolich/ci-server/configs"
	"github.com/FilipSolich/ci-server/models"
)

func IndexHandler(w http.ResponseWriter, r *http.Request, user *models.User) {
	configs.RenderTemplate(w, "index.html", struct {
		Username string
	}{
		Username: user.Username,
	})
}
