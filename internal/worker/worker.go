package worker

import (
	"archive/tar"
	"context"
	"io"
	"log"
	"log/slog"
	"os"
	"path"
	"strings"
	"time"

	dockertypes "github.com/docker/docker/api/types"
	containertypes "github.com/docker/docker/api/types/container"
	imagetypes "github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/yaml.v3"

	"github.com/shark-ci/shark-ci/internal/config"
	"github.com/shark-ci/shark-ci/internal/message_queue"
	pb "github.com/shark-ci/shark-ci/internal/proto"
	"github.com/shark-ci/shark-ci/internal/types"
)

func Run(mq message_queue.MessageQueuer, gRPCCLient pb.PipelineReporterClient, compressedReposPath string) error {
	workCh, err := mq.WorkChannel()
	if err != nil {
		return err
	}

	for i := 0; i < config.WorkerConf.MaxWorkers; i++ {
		go runWorker(workCh, gRPCCLient, compressedReposPath)
	}

	return nil
}

func runWorker(workCh chan types.Work, gRPCCLient pb.PipelineReporterClient, compressedReposPath string) {
	for work := range workCh {
		logger := slog.With("PipelineID", work.Pipeline.ID)

		tStart := time.Now()
		work.Pipeline.StartedAt = &tStart
		logger.Info("Start processing pipeline.")
		_, err := gRPCCLient.PipelineStarted(context.TODO(), &pb.PipelineStartedRequest{
			PipelineId: work.Pipeline.ID,
			StartedAt:  timestamppb.New(*work.Pipeline.StartedAt),
		})
		if err != nil {
			logger.Warn("Sending pipeline start message failed.", "err", err)
		}

		err = processWork(context.TODO(), work, config.WorkerConf.ReposPath, compressedReposPath)
		tEnd := time.Now()
		work.Pipeline.FinishedAt = &tEnd
		if err != nil {
			e := err.Error()
			_, err = gRPCCLient.PipelineFinnished(context.TODO(), &pb.PipelineFinnishedRequest{
				PipelineId: work.Pipeline.ID,
				FinishedAt: timestamppb.New(*work.Pipeline.FinishedAt),
				Error:      &e,
			})
			if err != nil {
				slog.Warn("Sending pipeline end message failed.", "time", tEnd.Sub(tStart), "err", err)
			}
			logger.Info("Processing pipeline failed.", "err", err)
			continue
		}

		logger.Info("Finished processing pipeline successfully.", "time", tEnd.Sub(tStart))
		_, err = gRPCCLient.PipelineFinnished(context.TODO(), &pb.PipelineFinnishedRequest{
			PipelineId: work.Pipeline.ID,
			FinishedAt: timestamppb.New(*work.Pipeline.FinishedAt),
		})
		if err != nil {
			logger.Warn("Sending pipeline end message failed.", "err", err)
		}
	}
}

func processWork(ctx context.Context, work types.Work, reposPath string, compressedReposPath string) error {
	// Update repository.
	repoPath := path.Join(reposPath, "FIX this")
	repo, err := updateRepo(ctx, repoPath, work.Pipeline.CloneURL, work.Token.AccessToken)
	if err != nil {
		return err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	err = worktree.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(work.Pipeline.CommitSHA),
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

	out, err := cli.ImagePull(ctx, pipeline.Image, imagetypes.PullOptions{})
	if err != nil {
		return err
	}
	io.ReadAll(out)
	out.Close()

	// Create container.
	container, err := cli.ContainerCreate(ctx, &containertypes.Config{
		Image:      pipeline.Image,
		Cmd:        []string{"sh"},
		Tty:        true,
		WorkingDir: "/repo",
	}, nil, nil, nil, "")
	if err != nil {
		return err
	}

	// Start container.
	err = cli.ContainerStart(ctx, container.ID, containertypes.StartOptions{})
	if err != nil {
		return err
	}

	// Create compressed repository.
	compressedRepo, err := archiveRepo(repoPath, compressedReposPath, "FIX: THIS", work.Pipeline.CommitSHA)
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
	err = cli.CopyToContainer(ctx, container.ID, "/", tarReader, dockertypes.CopyToContainerOptions{})
	if err != nil {
		return err
	}

	for name, j := range pipeline.Jobs {
		log.Printf("Pipeline %d runs %s\n", work.Pipeline.ID, name)
		for _, step := range j.Steps {
			log.Printf("Pipeline %d runs %s step %s\n", work.Pipeline.ID, name, step.Name)
			exec, err := cli.ContainerExecCreate(ctx, container.ID, dockertypes.ExecConfig{
				AttachStdout: true,
				AttachStderr: true,
				Cmd:          strings.Split(step.Cmds[0], " "),
			})
			if err != nil {
				return err
			}

			hijacked, err := cli.ContainerExecAttach(ctx, exec.ID, dockertypes.ExecStartCheck{})
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

			log.Printf("Job %d runs %s step %s logs: %d:%s\n", work.Pipeline.ID, name, step.Name, status.ExitCode, logs)
		}
	}

	// Stop container.
	err = cli.ContainerKill(ctx, container.ID, "SIGKILL")
	if err != nil {
		return err
	}

	// Delete container.
	err = cli.ContainerRemove(ctx, container.ID, containertypes.RemoveOptions{Force: true})
	if err != nil {
		return err
	}

	return nil
}
