package worker

import (
	"bytes"
	"context"
	"fmt"
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
	git_config "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	git_http "github.com/go-git/go-git/v5/plumbing/transport/http"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/yaml.v3"

	"github.com/shark-ci/shark-ci/internal/config"
	"github.com/shark-ci/shark-ci/internal/messagequeue"
	"github.com/shark-ci/shark-ci/internal/objectstore"
	pb "github.com/shark-ci/shark-ci/internal/proto"
	"github.com/shark-ci/shark-ci/internal/types"
)

func Run(mq messagequeue.MessageQueuer, objStore objectstore.ObjectStorer, gRPCCLient pb.PipelineReporterClient, compressedReposPath string) error {
	workCh, err := mq.WorkChannel()
	if err != nil {
		return err
	}

	for range config.WorkerConf.MaxWorkers {
		go runWorker(workCh, objStore, gRPCCLient, compressedReposPath)
	}

	return nil
}

func runWorker(workCh chan types.Work, objStore objectstore.ObjectStorer, gRPCCLient pb.PipelineReporterClient, compressedReposPath string) {
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

		err = processWork(context.TODO(), objStore, work, config.WorkerConf.ReposPath, compressedReposPath)
		tEnd := time.Now()
		work.Pipeline.FinishedAt = &tEnd
		if err != nil {
			e := err.Error()
			_, err = gRPCCLient.PipelineFinnished(context.TODO(), &pb.PipelineFinnishedRequest{
				PipelineId: work.Pipeline.ID,
				FinishedAt: timestamppb.New(*work.Pipeline.FinishedAt),
				Status:     pb.PipelineFinnishedStatus_FAILURE,
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
			Status:     pb.PipelineFinnishedStatus_SUCCESS,
		})
		if err != nil {
			logger.Warn("Sending pipeline end message failed.", "err", err)
		}
	}
}

func processWork(ctx context.Context, objStore objectstore.ObjectStorer, work types.Work, reposPath string, compressedReposPath string) error {
	// Update repository.
	//repoPath := path.Join(reposPath, "FIX this")
	//repo, err := updateRepo(ctx, repoPath, work.Pipeline.CloneURL, work.Token.AccessToken)
	//if err != nil {
	//	return err
	//}

	dir, err := os.MkdirTemp("/tmp", "shark-ci-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	repo, err := git.PlainInit(dir, false)
	if err != nil {
		return err
	}
	_, err = repo.CreateRemote(&git_config.RemoteConfig{
		Name: "origin",
		URLs: []string{work.Pipeline.CloneURL},
	})
	if err != nil {
		return err
	}
	err = repo.FetchContext(ctx, &git.FetchOptions{
		RemoteName: "origin",
		Depth:      1,
		RefSpecs: []git_config.RefSpec{
			git_config.RefSpec(fmt.Sprintf("%s:refs/heads/test", work.Pipeline.CommitSHA)),
		},
		Auth: &git_http.BasicAuth{
			Username: "abc",
			Password: work.Token.AccessToken,
		},
		Progress: log.Writer(),
	})
	if err != nil {
		return err
	}
	tree, err := repo.Worktree()
	if err != nil {
		return err
	}
	err = tree.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(work.Pipeline.CommitSHA),
	})
	if err != nil {
		return err
	}

	//git.PlainCloneContext(ctx, dir, false, &git.CloneOptions{
	//	URL: work.Pipeline.CloneURL,
	//	Auth: &git_http.BasicAuth{
	//		Username: "abc",
	//		Password: work.Token.AccessToken,
	//	},
	//	Progress: log.Writer(),
	//})

	//worktree, err := repo.Worktree()
	//if err != nil {
	//	return err
	//}

	//err = worktree.Checkout(&git.CheckoutOptions{
	//	Hash: plumbing.NewHash(work.Pipeline.CommitSHA),
	//})
	//if err != nil {
	//	return err
	//}

	// Parse pipeline.
	file, err := os.Open(path.Join(dir, ".shark-ci/workflow.yaml"))
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
	//compressedRepo, err := archiveRepo(dir, compressedReposPath, "FIX: THIS", work.Pipeline.CommitSHA)
	//if err != nil {
	//	return err
	//}

	//a, err := os.Open(compressedRepo)
	//if err != nil {
	//	return err
	//}
	//defer file.Close()

	//tarReader := tar.NewReader(a)

	//// This doesnt work.
	//err = cli.CopyToContainer(ctx, container.ID, "/", tarReader, dockertypes.CopyToContainerOptions{})
	//if err != nil {
	//	return err
	//}

	logsBuff := &bytes.Buffer{}

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

			_, err = io.Copy(logsBuff, hijacked.Reader)
			if err != nil {
				hijacked.Close()
				return err
			}
			hijacked.Close()

			_, err = cli.ContainerExecInspect(ctx, exec.ID) // TODO: Will containen status code.
			if err != nil {
				return err
			}

			log.Printf("Job %d runs %s step %s\n", work.Pipeline.ID, name, step.Name)
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

	// Upload logs
	objStore.UploadLogs(ctx, work.Pipeline.ID, logsBuff, int64(logsBuff.Len()))

	return nil
}
