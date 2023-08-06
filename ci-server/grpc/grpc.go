package grpc

import (
	"context"
	"time"

	ciserver "github.com/FilipSolich/shark-ci/ci-server"
	"github.com/FilipSolich/shark-ci/ci-server/service"
	"github.com/FilipSolich/shark-ci/ci-server/store"
	pb "github.com/FilipSolich/shark-ci/shared/proto"
	"golang.org/x/exp/slog"
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

func (s *GRPCServer) PipelineStart(ctx context.Context, in *pb.PipelineStartRequest) (*pb.Void, error) {
	err := s.changePipelineState(ctx, in.GetPipelineId(), in.GetStartedAt().AsTime(), true)
	return &pb.Void{}, err
}

func (s *GRPCServer) PipelineEnd(ctx context.Context, in *pb.PipelineEndRequest) (*pb.Void, error) {
	err := s.changePipelineState(ctx, in.GetPipelineId(), in.GetFinishedAt().AsTime(), false)
	return &pb.Void{}, err
}

func (s *GRPCServer) changePipelineState(ctx context.Context, pipelineID int64, t time.Time, start bool) error {
	commitSHA, targetURL, repoName, repoOwner, srvName, token, err := s.s.GetInfoForPipelineStateChange(ctx, pipelineID)
	if err != nil {
		s.l.Error("store: cannot get info for pipeline state change", "err", err)
		return err
	}

	srv, ok := s.services[srvName]
	if !ok {
		s.l.Error("service: service not found", "service", srvName)
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
		desc = "Pipeline finished successfully"
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
		TargetURL:   targetURL,
		Context:     ciserver.CIServer,
		Description: desc,
	}
	err = srv.CreateStatus(ctx, token, repoOwner, repoName, commitSHA, status)
	if err != nil {
		s.l.Error("service: cannot create status", "err", err)
		return err
	}

	return nil
}
