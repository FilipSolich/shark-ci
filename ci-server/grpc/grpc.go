package grpc

import (
	"context"

	ciserver "github.com/FilipSolich/shark-ci/ci-server"
	"github.com/FilipSolich/shark-ci/ci-server/models"
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

func (s *GRPCServer) ReportPipelineStatus(ctx context.Context, in *pb.PipelineStartRequest) *pb.Void {
	// Get All things needed

	// Update pipeline in DB
	srv := s.services["GitHub"]

	serviceUser := &models.ServiceUser{}
	pipeline := &models.Pipeline{}
	repoName := "T"

	status := service.Status{
		State:       service.StatusRunning,
		TargetURL:   pipeline.TargetURL,
		Context:     ciserver.CIServer,
		Description: "Pipeline is running",
	}
	err := srv.CreateStatus(ctx, serviceUser, repoName, pipeline.CommitSHA, status)
	if err != nil {
		s.l.Error("cannot create status", "err", err)
	}

	return &pb.Void{}
}
