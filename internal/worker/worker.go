package worker

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"os"
	"path"
	"strings"
	"time"

	dockertypes "github.com/docker/docker/api/types"
	containertypes "github.com/docker/docker/api/types/container"
	imagetypes "github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/yaml.v3"

	"github.com/shark-ci/shark-ci/internal/messagequeue"
	"github.com/shark-ci/shark-ci/internal/objectstore"
	pb "github.com/shark-ci/shark-ci/internal/proto"
	"github.com/shark-ci/shark-ci/internal/types"
)

func Run(mq messagequeue.MessageQueuer, objStore objectstore.ObjectStorer, gRPCCLient pb.PipelineReporterClient) error {
	workCh, err := mq.WorkChannel()
	if err != nil {
		return err
	}

	for work := range workCh {
		go runWorker(work, objStore, gRPCCLient)
	}

	return nil
}

func runWorker(work types.Work, objStore objectstore.ObjectStorer, gRPCCLient pb.PipelineReporterClient) {
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

	err = processWork(context.TODO(), objStore, work)
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
		logger.Info("Processing pipeline failed.", "err", e)
		return
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

func processWork(ctx context.Context, objStore objectstore.ObjectStorer, work types.Work) error {
	dir, err := cloneRepo(ctx, work.Pipeline.CloneURL, work.Pipeline.CommitSHA, work.Token)
	defer os.RemoveAll(dir)
	if err != nil {
		return err
	}

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
	container, err := cli.ContainerCreate(
		ctx,
		&containertypes.Config{
			Image:      pipeline.Image,
			Tty:        true,
			WorkingDir: "/app",
		},
		&containertypes.HostConfig{
			Binds: []string{dir + ":/app"},
		}, nil, nil, "")
	if err != nil {
		return err
	}

	// Start container.
	err = cli.ContainerStart(ctx, container.ID, containertypes.StartOptions{})
	if err != nil {
		return err
	}

	logsBuff := &bytes.Buffer{}
	for _, cmd := range pipeline.Cmds {
		exec, err := cli.ContainerExecCreate(ctx, container.ID, dockertypes.ExecConfig{
			AttachStdout: true,
			AttachStderr: true,
			Cmd:          strings.Split(cmd, " "),
		})
		if err != nil {
			return err
		}

		hijacked, err := cli.ContainerExecAttach(ctx, exec.ID, dockertypes.ExecStartCheck{})
		if err != nil {
			return err
		}

		_, err = stdcopy.StdCopy(logsBuff, logsBuff, hijacked.Reader)
		if err != nil {
			hijacked.Close()
			return err
		}
		hijacked.Close()

		_, err = cli.ContainerExecInspect(ctx, exec.ID) // TODO: Will containen status code.
		if err != nil {
			return err
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
