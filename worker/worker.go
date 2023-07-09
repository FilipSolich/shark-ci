package worker

import (
	"archive/tar"
	"context"
	"io"
	"log"
	"os"
	"path"
	"strings"

	"github.com/FilipSolich/shark-ci/message_queue"
	"github.com/FilipSolich/shark-ci/model"
	"github.com/docker/docker/api/types"
	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"gopkg.in/yaml.v3"
)

func Run(mq message_queue.MessageQueuer, maxWorkers int, reposPath string, compressedReposPath string) error {
	jobCh, err := mq.JobChannel()
	if err != nil {
		return err
	}

	for i := 0; i < maxWorkers; i++ {
		go func() {
			for job := range jobCh {
				log.Printf("Processing job %s\n", job.ID)
				ctx := context.TODO()
				err := processJob(ctx, job, reposPath, compressedReposPath)
				if err != nil {
					// TODO: Should be failed job returned to queue?
					job.Nack()
					log.Printf("Job %s failed: %s\n", job.ID, err.Error())
					continue
				}
				job.Ack()
				log.Printf("Job %s processed successfully\n", job.ID)
			}
		}()
	}

	return nil
}

func processJob(ctx context.Context, job model.Job, reposPath string, compressedReposPath string) error {
	// Update repository.
	repoPath := path.Join(reposPath, job.UniqueName)
	repo, err := updateRepo(ctx, repoPath, job.CloneURL, job.Token.AccessToken)
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
	defer file.Close()

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
		Image:      pipeline.BaseImage,
		Cmd:        []string{"sh"},
		Tty:        true,
		WorkingDir: "/repo",
	}, nil, nil, nil, "")
	if err != nil {
		return err
	}

	// Start container.
	err = cli.ContainerStart(ctx, container.ID, types.ContainerStartOptions{})
	if err != nil {
		return err
	}

	// Create compressed repository.
	compressedRepo, err := archiveRepo(repoPath, compressedReposPath, job.UniqueName, job.CommitSHA)
	if err != nil {
		return err
	}

	a, err := os.Open(compressedRepo)
	if err != nil {
		return err
	}
	defer file.Close()

	tarReader := tar.NewReader(a)

	// This doesnt work.
	err = cli.CopyToContainer(ctx, container.ID, "/", tarReader, types.CopyToContainerOptions{})
	if err != nil {
		return err
	}

	for name, j := range pipeline.Jobs {
		log.Printf("Job %s runs %s\n", job.ID, name)
		for _, step := range j.Steps {
			log.Printf("Job %s runs %s step %s\n", job.ID, name, step.Name)
			exec, err := cli.ContainerExecCreate(ctx, container.ID, types.ExecConfig{
				AttachStdout: true,
				AttachStderr: true,
				Cmd:          strings.Split(step.Run, " "),
			})
			if err != nil {
				return err
			}

			hijacked, err := cli.ContainerExecAttach(ctx, exec.ID, types.ExecStartCheck{})
			if err != nil {
				return err
			}

			logs, err := io.ReadAll(hijacked.Reader)
			if err != nil {
				hijacked.Close()
				return err
			}
			hijacked.Close()

			status, err := cli.ContainerExecInspect(ctx, exec.ID) // TODO: Will containen status code.
			if err != nil {
				return err
			}

			log.Printf("Job %s runs %s step %s logs: %d:%s\n", job.ID, name, step.Name, status.ExitCode, logs)
		}
	}

	// Stop container.
	err = cli.ContainerKill(ctx, container.ID, "SIGKILL")
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
