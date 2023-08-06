package worker

import (
	"archive/tar"
	"context"
	"io"
	"log"
	"os"
	"path"
	"strings"
	"time"

	dockertypes "github.com/docker/docker/api/types"
	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"golang.org/x/exp/slog"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/yaml.v3"

	"github.com/FilipSolich/shark-ci/shared/message_queue"
	pb "github.com/FilipSolich/shark-ci/shared/proto"
	"github.com/FilipSolich/shark-ci/shared/types"
)

func Run(mq message_queue.MessageQueuer, gRPCCLient pb.PipelineReporterClient, maxWorkers int, reposPath string, compressedReposPath string) error {
	workCh, err := mq.WorkChannel()
	if err != nil {
		return err
	}

	for i := 0; i < maxWorkers; i++ {
		go func() {
			for work := range workCh {
				tStart := time.Now()
				work.Pipeline.StartedAt = &tStart
				slog.Info("start processing pipeline", "PipelineID", work.Pipeline.ID)
				_, err = gRPCCLient.PipelineStart(context.TODO(), &pb.PipelineStartRequest{
					PipelineId: work.Pipeline.ID,
					StartedAt:  timestamppb.New(*work.Pipeline.StartedAt),
				})
				if err != nil {
					slog.Warn("sending pipeline start message failed", "PipelineID", work.Pipeline.ID, "err", err)
				}

				//err := processWork(context.TODO(), work, reposPath, compressedReposPath)
				//if err != nil {
				//	// Send error message.
				//	slog.Info("processing pipeline failed", "PipelineID", work.Pipeline.ID, "err", err)
				//	continue
				//}

				tEnd := time.Now()
				work.Pipeline.FinishedAt = &tEnd
				slog.Info("finished processing pipeline successfully", "PipelineID", work.Pipeline.ID)
				_, err = gRPCCLient.PipelineEnd(context.TODO(), &pb.PipelineEndRequest{
					PipelineId: work.Pipeline.ID,
					FinishedAt: timestamppb.New(*work.Pipeline.FinishedAt),
				})
				if err != nil {
					slog.Warn("sending pipeline end message failed", "PipelineID", work.Pipeline.ID, "err", err)
				}
			}
		}()
	}

	return nil
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

	out, err := cli.ImagePull(ctx, pipeline.BaseImage, dockertypes.ImagePullOptions{})
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
	err = cli.ContainerStart(ctx, container.ID, dockertypes.ContainerStartOptions{})
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
		log.Printf("Pipelin %d runs %s\n", work.Pipeline.ID, name)
		for _, step := range j.Steps {
			log.Printf("Pipeline %d runs %s step %s\n", work.Pipeline.ID, name, step.Name)
			exec, err := cli.ContainerExecCreate(ctx, container.ID, dockertypes.ExecConfig{
				AttachStdout: true,
				AttachStderr: true,
				Cmd:          strings.Split(step.Run, " "),
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
	err = cli.ContainerRemove(ctx, container.ID, dockertypes.ContainerRemoveOptions{Force: true})
	if err != nil {
		return err
	}

	return nil
}
