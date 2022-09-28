package handlers

import (
	"net/http"
)

func EventHandler(w http.ResponseWriter, r *http.Request) {
	//payload, err := github.ValidatePayload(r, []byte(configs.WebhookSecret))
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//	return
	//}

	//event, err := github.ParseWebHook(github.WebHookType(r), payload)
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//	return
	//}

	//switch event := event.(type) {
	//case *github.PushEvent:
	//	commit := event.Commits[len(event.Commits)-1]

	//	var user models.User
	//	username := event.Repo.Owner.GetLogin()
	//	result := db.DB.Where(&models.User{Service: "github", Username: username}).First(&user)
	//	if result.Error != nil {
	//		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
	//		return
	//	}

	//	var token models.OAuth2Token
	//	result = db.DB.Where(&models.OAuth2Token{UserID: user.ID}).First(&token)
	//	if result.Error != nil {
	//		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
	//		return
	//	}

	//	job := &models.Job{
	//		CommitSHA: commit.GetID(),
	//		CloneURL:  event.Repo.GetCloneURL(),
	//		Token:     token.Token,
	//	}
	//	result = db.DB.Save(job)
	//	if result.Error != nil {
	//		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
	//		return
	//	}

	//	err = mq.MQ.PublishJob(job)
	//	if err != nil {
	//		fmt.Println(err)
	//	}

	//	//services.UpdateStatus()
	//}
}
