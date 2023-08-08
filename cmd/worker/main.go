package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/shark-ci/shark-ci/shared/message_queue"
	pb "github.com/shark-ci/shark-ci/shared/proto"
	"github.com/shark-ci/shark-ci/worker"
	"github.com/shark-ci/shark-ci/worker/config"
	"golang.org/x/exp/slog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	conf, err := config.NewConfigFromEnv()
	if err != nil {
		slog.Error("Creating config from environment failed.", "err", err)
		os.Exit(1)
	}
	config.Conf = conf

	compressedReposPath, err := worker.CreateTmpDir()
	if err != nil {
		slog.Error("Creating TMP dir failed.", "err", err)
		os.Exit(1)
	}

	slog.Info("Connecting to RabbitMQ.")
	rabbitMQ, err := message_queue.NewRabbitMQ(conf.MQ.URI)
	if err != nil {
		slog.Error("Connecting to RabbitMQ failed", "err", err)
		os.Exit(1)
	}
	defer rabbitMQ.Close(context.TODO())
	slog.Info("RabbitMQ connected.")

	slog.Info("Creating gRPC client.")
	conn, err := grpc.Dial(conf.CIServer.Host+":"+conf.CIServer.GRPCPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("Connecting to gRPC server failed.", "addr", conf.CIServer.Host+":"+conf.CIServer.GRPCPort, "err", err)
		os.Exit(1)
	}
	defer conn.Close()
	gRPCClient := pb.NewPipelineReporterClient(conn)
	slog.Info("gRPC client created.")

	err = worker.Run(rabbitMQ, gRPCClient, compressedReposPath)
	if err != nil {
		slog.Error("Running worker failed", "err", err)
		os.Exit(1)
	}

	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, os.Interrupt)
	<-signalCh
}
