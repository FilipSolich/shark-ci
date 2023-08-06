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
	pipeline, err := s.s.GetPipeline(ctx, pipelineID)
	if err != nil {
		s.l.Error("store: cannot get pipeline", "err", err)
		return err
	}

	repo, err := s.s.GetRepo(ctx, pipeline.RepoID)
	if err != nil {
		s.l.Error("store: cannot get repo", "err", err)
		return err
	}

	srv, ok := s.services[repo.Service]
	if !ok {
		s.l.Error("service: service not found", "service", repo.Service)
		return err
	}

	var statusState service.StatusState
	var description string
	if start {
		statusState = service.StatusRunning
		pipeline.StartedAt = &t
		pipeline.Status = srv.StatusName(statusState)
		description = "Pipeline is running"
	} else {
		statusState = service.StatusSuccess
		pipeline.FinishedAt = &t
		pipeline.Status = srv.StatusName(statusState)
		description = "Pipeline finished successfully"
	}
	err = s.s.UpdatePipelineStatus(ctx, pipeline.ID, pipeline.Status, pipeline.StartedAt, pipeline.FinishedAt)
	if err != nil {
		s.l.Error("store: cannot update pipeline", "err", err)
		return err
	}

	serviceUser, err := s.s.GetServiceUserByRepo(ctx, repo.ID)
	if err != nil {
		s.l.Error("store: cannot get service user", "err", err)
		return err
	}

	status := service.Status{
		State:       statusState,
		TargetURL:   pipeline.TargetURL,
		Context:     ciserver.CIServer,
		Description: description,
	}
	err = srv.CreateStatus(ctx, serviceUser, serviceUser.Username, repo.Name, pipeline.CommitSHA, status)
	if err != nil {
		s.l.Error("service: cannot create status", "err", err)
		return err
	}

	return nil
}
