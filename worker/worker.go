package worker

import (
	"context"
	"io"
	"log"
	"os"
	"path"

	"github.com/docker/docker/api/types"
	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
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
				ctx := context.TODO()
				err := processJob(ctx, job, reposPath)
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

func processJob(ctx context.Context, job models.Job, reposPath string) error {
	// Update repository.
	repoPath := path.Join(reposPath, job.UniqueName)
	repo, err := UpdateRepo(ctx, repoPath, job.CloneURL, job.Token.AccessToken)
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

	// Parse pipeline.
	file, err := os.Open(path.Join(repoPath, ".shark-ci/workflow.yaml"))
	if err != nil {
		return err
	}

	var pipeline Pipeline
	err = yaml.NewDecoder(file).Decode(&pipeline)
	if err != nil {
		return err
	}

	// Pull base image.
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer cli.Close()

	out, err := cli.ImagePull(ctx, pipeline.BaseImage, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	io.ReadAll(out)
	out.Close()

	// Create container.
	container, err := cli.ContainerCreate(ctx, &containertypes.Config{
		Image: pipeline.BaseImage,
		Cmd:   []string{"echo", "hello world"},
		Tty:   true,
	}, nil, nil, nil, "")
	if err != nil {
		return err
	}

	err = cli.ContainerStart(ctx, container.ID, types.ContainerStartOptions{})
	if err != nil {
		return err
	}

	// TODO: Start container with copied repo and run commands
	// TODO: Report result

	// Stop container.
	wait := 0
	err = cli.ContainerStop(ctx, container.ID, containertypes.StopOptions{Timeout: &wait})
	if err != nil {
		return err
	}

	// Delete container.
	err = cli.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{Force: true})
	if err != nil {
		return err
	}

	return nil
}
