package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/FilipSolich/shark-ci/shared/message_queue"
	pb "github.com/FilipSolich/shark-ci/shared/proto"
	"github.com/FilipSolich/shark-ci/worker"
	"github.com/FilipSolich/shark-ci/worker/config"
	"golang.org/x/exp/slog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	config, err := config.NewConfigFromEnv()
	if err != nil {
		slog.Error("creating config failed", "err", err)
		os.Exit(1)
	}

	compressedReposPath, err := worker.CreateTmpDir()
	if err != nil {
		slog.Error("creating tmp dir failed", "err", err)
		os.Exit(1)
	}

	slog.Info("connecting to RabbitMQ")
	rabbitMQ, err := message_queue.NewRabbitMQ(config.MQ.URI)
	if err != nil {
		slog.Error("mq: connecting to RabbitMQ failed", "err", err)
		os.Exit(1)
	}
	defer rabbitMQ.Close(context.TODO())
	slog.Info("RabbitMQ connected")

	slog.Info("creating gRPC client")
	conn, err := grpc.Dial("localhost:8010", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("grpc: connecting to gRPC server failed", "err", err)
		os.Exit(1)
	}
	defer conn.Close()
	gRPCClient := pb.NewPipelineReporterClient(conn)
	slog.Info("gRPC client created")

	err = worker.Run(rabbitMQ, gRPCClient, config.Worker.MaxWorkers, config.Worker.ReposPath, compressedReposPath)
	if err != nil {
		slog.Error("worker: running worker failed", "err", err)
		os.Exit(1)
	}

	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, os.Interrupt)
	<-signalCh
}
