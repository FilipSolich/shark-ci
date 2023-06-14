package worker

import (
	"context"
	"log"
	"os"
	"path"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-yaml/yaml"
	"github.com/shark-ci/shark-ci/message_queue"
	"github.com/shark-ci/shark-ci/models"
)

func Run(mq message_queue.MessageQueuer, maxWorkers int, reposPath string) error {
	jobCh, err := mq.JobChannel()
	if err != nil {
		return err
	}

	for i := 0; i < maxWorkers; i++ {
		go func() {
			for job := range jobCh {
				log.Printf("Processing job %s\n", job.ID)
				err := processJob(job, reposPath)
				if err != nil {
					// TODO: Should be failed job returned to queue?
					job.Nack()
					log.Printf("Job %s failed: %s\n", job.ID, err.Error())
				}
				job.Ack()
				log.Printf("Job %s processed successfully\n", job.ID)
			}
		}()
	}

	return nil
}

func processJob(job models.Job, reposPath string) error {
	repoPath := path.Join(reposPath, job.UniqueName)
	repo, err := UpdateRepo(context.TODO(), repoPath, job.CloneURL, job.Token.AccessToken)
	if err != nil {
		return err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	err = worktree.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(job.CommitSHA),
	})
	if err != nil {
		return err
	}

	// TODO: Parse YAML
	file, err := os.Open(path.Join(repoPath, ".shark-ci/workflow.yaml"))
	if err != nil {
		return err
	}

	var pipeline Pipeline
	err = yaml.NewDecoder(file).Decode(&pipeline)
	if err != nil {
		return err
	}

	// TODO: Create container
	// TODO: Start container with copied repo and run commands
	// TODO: Report result
	// TODO: Delete container
	return nil
}
