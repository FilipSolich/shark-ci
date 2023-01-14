package runner

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/go-git/go-git/v5"
	git_http "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/shark-ci/shark-ci/ci-server/db"
	"github.com/shark-ci/shark-ci/mq"
)

func Run() error {
	msgs, err := mq.MQ.GetJobsChanel()
	if err != nil {
		return err
	}

	for msg := range msgs {
		// TODO: Send info to CI server
		var job db.Job
		err := json.Unmarshal(msg.Body, &job)
		if err != nil {
			// TODO: Send info to CI server
			log.Println(err)
			continue
		}

		go processJob(&job)
	}

	return nil
}

func processJob(job *db.Job) {
	// TODO: Clone or fetch repo

	err := os.Mkdir("testdir", 0750)
	if err != nil && !os.IsExist(err) {
		log.Fatal(err)
	}

	// Clones the repository into the given dir, just as a normal git clone does
	_, err = git.PlainClone("testdir", false, &git.CloneOptions{
		URL: job.CloneURL,
		Auth: &git_http.BasicAuth{
			Username: "abc123", // anything except an empty string
			Password: job.Token.AccessToken,
		},
	})
	fmt.Println(err)

	// TODO: Parse YAML
	// TODO: Create container
	// TODO: Start container with mounted repo and run commands
	// TODO: Report result
	// TODO: Delete container
	fmt.Println(job.ID.String(), job.CloneURL)
}
