package handlers

import (
	"net/http"

	"github.com/FilipSolich/ci-server/configs"
	"github.com/FilipSolich/ci-server/models"
)

type templateData struct {
	Title    string
	Username string
}

func IndexHandler(w http.ResponseWriter, r *http.Request, user *models.User) {
	configs.Templates.ExecuteTemplate(w, "index.html", templateData{
		Title:    "New Title",
		Username: user.Username,
	})
}
