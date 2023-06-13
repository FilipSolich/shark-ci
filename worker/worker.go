package worker

import (
	"context"
	"log"
	"path"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
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
				err := processJob(job, reposPath)
				if err != nil {
					log.Println(err)
					// TODO: Should be failed job returned to queue?
					job.Nack()
				}
				job.Ack()
			}
		}()
	}

	return nil
}

func processJob(job models.Job, reposPath string) error {
	log.Printf("Processing job %s\n", job.ID)

	repoPath := path.Join(reposPath, job.UniqueName)
	repo, err := UpdateRepo(context.TODO(), repoPath, job.CloneURL, job.Token.AccessToken)

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

	return nil

	// TODO: Parse YAML
	// TODO: Create container
	// TODO: Start container with mounted repo and run commands
	// TODO: Report result
	// TODO: Delete container
}
