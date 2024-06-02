package grpc

import (
	"context"
	"log/slog"

	pb "github.com/shark-ci/shark-ci/internal/proto"
	"github.com/shark-ci/shark-ci/internal/server/service"
	"github.com/shark-ci/shark-ci/internal/server/store"
	"github.com/shark-ci/shark-ci/internal/types"
)

type GRPCServer struct {
	pb.UnimplementedPipelineReporterServer
	s        store.Storer
	services service.Services
}

var _ pb.PipelineReporterServer = &GRPCServer{}

func NewGRPCServer(s store.Storer, services service.Services) *GRPCServer {
	return &GRPCServer{
		s:        s,
		services: services,
	}
}

func (s *GRPCServer) PipelineStarted(ctx context.Context, in *pb.PipelineStartedRequest) (*pb.Empty, error) {
	info, err := s.s.GetPipelineStateChangeInfo(ctx, in.PipelineId)
	if err != nil {
		slog.Error("store: cannot get info for pipeline state change", "pipelineID", in.PipelineId, "err", err)
		return nil, err
	}

	srv, ok := s.services[info.Service]
	if !ok {
		slog.Error("service: service not found", "service", info.Service)
		return nil, err
	}

	pipelineStatus := types.Running
	err = s.s.PipelineStarted(ctx, in.PipelineId, pipelineStatus, in.GetStartedAt().AsTime())
	if err != nil {
		slog.Error("store: cannot update pipeline", "err", err)
		return nil, err
	}

	status := service.Status{
		State:       pipelineStatus,
		TargetURL:   info.URL,
		Context:     "Shark CI",
		Description: "Pipeline is running",
	}
	err = srv.CreateStatus(ctx, &info.Token, info.RepoOwner, info.RepoName, info.CommitSHA, status)
	if err != nil {
		slog.Error("service: cannot create status", "err", err)
		return nil, err
	}
	return &pb.Empty{}, err
}

func (s *GRPCServer) PipelineFinnished(ctx context.Context, in *pb.PipelineFinnishedRequest) (*pb.Empty, error) {
	info, err := s.s.GetPipelineStateChangeInfo(ctx, in.PipelineId)
	if err != nil {
		slog.Error("store: cannot get info for pipeline state change", "pipelineID", in.PipelineId, "err", err)
		return nil, err
	}

	srv, ok := s.services[info.Service]
	if !ok {
		slog.Error("service: service not found", "service", info.Service)
		return nil, err
	}

	pipelineStatus := types.Success
	description := "Pipeline finnished successfully"
	if in.Status == pb.PipelineFinnishedStatus_FAILURE {
		pipelineStatus = types.Error
		description = "Pipeline failed"
	}
	err = s.s.PipelineFinnished(ctx, in.PipelineId, pipelineStatus, in.GetFinishedAt().AsTime())
	if err != nil {
		slog.Error("store: cannot update pipeline", "err", err)
		return nil, err
	}

	status := service.Status{
		State:       pipelineStatus,
		TargetURL:   info.URL,
		Context:     "Shark CI",
		Description: description,
	}
	err = srv.CreateStatus(ctx, &info.Token, info.RepoOwner, info.RepoName, info.CommitSHA, status)
	if err != nil {
		slog.Error("service: cannot create status", "err", err)
		return nil, err
	}
	return &pb.Empty{}, nil
}

func (s *GRPCServer) CommandOutput(ctx context.Context, in *pb.CommandOutputRequest) (*pb.Empty, error) {
	_, err := s.s.CreatePipelineLog(ctx, types.PipelineLog{
		Order:      int(in.Order),
		Cmd:        in.Cmd,
		Output:     in.Output,
		ExitCode:   int(in.ExitCode),
		PipelineID: in.PipelineId,
	})
	if err != nil {
		slog.Error("Cannot create pipeline log.", "err", err)
		return nil, err
	}
	return &pb.Empty{}, nil
}
