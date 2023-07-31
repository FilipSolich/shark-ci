package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/FilipSolich/shark-ci/shared/message_queue"
	"github.com/FilipSolich/shark-ci/worker"
	"github.com/FilipSolich/shark-ci/worker/config"
	"golang.org/x/exp/slog"
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

	err = worker.Run(rabbitMQ, config.Worker.MaxWorkers, config.Worker.ReposPath, compressedReposPath)
	if err != nil {
		slog.Error("worker: running worker failed", "err", err)
		os.Exit(1)
	}

	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, os.Interrupt)
	<-signalCh
}
