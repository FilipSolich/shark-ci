package worker

import (
	"fmt"
	"log"
	"os"

	"github.com/go-git/go-git/v5"
	git_http "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/shark-ci/shark-ci/message_queue"
	"github.com/shark-ci/shark-ci/models"
)

func Run(mq message_queue.MessageQueuer) error {
	jobCh, err := mq.JobChannel()
	if err != nil {
		return err
	}

	for i := 0; i < 10; i++ {
		go func() {
			for job := range jobCh {
				err := processJob(job)
				if err != nil {
					log.Println(err)
					job.Nack()
				}
				job.Ack()
			}
		}()
	}

	return nil
}

func processJob(job models.Job) error {
	// TODO: Clone or fetch repo

	fmt.Printf("Processing job %s\n", job.ID)

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
	return err

	// TODO: Parse YAML
	// TODO: Create container
	// TODO: Start container with mounted repo and run commands
	// TODO: Report result
	// TODO: Delete container
}
