package grpc

import (
	"context"
	"fmt"
	"time"

	"log/slog"

	"github.com/shark-ci/shark-ci/ci-server/service"
	"github.com/shark-ci/shark-ci/ci-server/store"
	pb "github.com/shark-ci/shark-ci/shared/proto"
)

type GRPCServer struct {
	pb.UnimplementedPipelineReporterServer
	l        *slog.Logger
	s        store.Storer
	services service.Services
}

func NewGRPCServer(l *slog.Logger, s store.Storer, services service.Services) *GRPCServer {
	return &GRPCServer{
		l:        l,
		s:        s,
		services: services,
	}
}

func (s *GRPCServer) PipelineStart(ctx context.Context, in *pb.PipelineStartRequest) (*pb.Empty, error) {
	err := s.changePipelineState(ctx, in.GetPipelineId(), in.GetStartedAt().AsTime(), true)
	return &pb.Empty{}, err
}

func (s *GRPCServer) PipelineFinnishedSuccessfuly(ctx context.Context, in *pb.PipelineEndRequest) (*pb.Empty, error) {
	err := s.changePipelineState(ctx, in.GetPipelineId(), in.GetFinishedAt().AsTime(), false)
	return &pb.Empty{}, err
}

func (s *GRPCServer) PipelineFailed(ctx context.Context, in *pb.PipelineEndRequest) (*pb.Empty, error) {
	// TODO: Implement.
	err := s.changePipelineState(ctx, in.GetPipelineId(), in.GetFinishedAt().AsTime(), false)
	return &pb.Empty{}, err
}

func (s *GRPCServer) changePipelineState(ctx context.Context, pipelineID int64, t time.Time, start bool) error {
	info, err := s.s.GetPipelineStateChangeInfo(ctx, pipelineID)
	if err != nil {
		s.l.Error("store: cannot get info for pipeline state change", "pipelineID", pipelineID, "err", err)
		return err
	}

	srv, ok := s.services[info.Service]
	if !ok {
		s.l.Error("service: service not found", "service", info.Service)
		return err
	}

	statusState := service.StatusRunning
	statusName := srv.StatusName(statusState)
	desc := "Pipeline is running"
	var startedAt *time.Time = &t
	var finishedAt *time.Time = nil
	if !start {
		statusState = service.StatusSuccess
		statusName = srv.StatusName(statusState)
		desc = fmt.Sprintf("Pipeline finished successfully in %s", t.Sub(*info.StartedAt).Round(time.Second))
		startedAt = nil
		finishedAt = &t
	}
	err = s.s.UpdatePipelineStatus(ctx, pipelineID, statusName, startedAt, finishedAt)
	if err != nil {
		s.l.Error("store: cannot update pipeline", "err", err)
		return err
	}

	status := service.Status{
		State:       statusState,
		TargetURL:   info.URL,
		Context:     info.Context,
		Description: desc,
	}
	err = srv.CreateStatus(ctx, &info.Token, info.RepoOwner, info.RepoName, info.CommitSHA, status)
	if err != nil {
		s.l.Error("service: cannot create status", "err", err)
		return err
	}

	return nil
}
