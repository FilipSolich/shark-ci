package handlers

import (
	"fmt"
	"net/http"

	"github.com/FilipSolich/ci-server/configs"
	"github.com/google/go-github/v47/github"
)

func EventHandler(w http.ResponseWriter, r *http.Request) {
	payload, err := github.ValidatePayload(r, []byte(configs.WebhookSecret))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	switch event := event.(type) {
	case *github.PushEvent:
		fmt.Println("a", event.Action)
	}
}
