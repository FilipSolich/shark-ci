package runner

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/FilipSolich/ci-server/db"
	"github.com/FilipSolich/ci-server/mq"
)

func Run() error {
	msgs, err := mq.MQ.GetJobsChanel()
	if err != nil {
		return err
	}
	fmt.Println("HERE")
	for msg := range msgs {
		// TODO: Send info to CI server
		var job db.Job
		err := json.Unmarshal(msg.Body, &job)
		if err != nil {
			// TODO: Send info to CI server
			log.Println(err)
			continue
		}

		processJob(&job)
	}

	return nil
}

func processJob(job *db.Job) {
	// TODO: Clone or fetch repo
	// TODO: Parse YAML
	// TODO: Create container
	// TODO: Start container with mounted repo and run commands
	// TODO: Report result
	// TODO: Delete container
	fmt.Print(job.ID, job.Repo, job.CloneURL)
}
