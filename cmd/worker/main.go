package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/FilipSolich/shark-ci/message_queue"
	"github.com/FilipSolich/shark-ci/worker"
	"github.com/FilipSolich/shark-ci/worker/config"
	"github.com/joho/godotenv"
)

func main() {
	// TODO: Remove.
	err := godotenv.Load()
	if err != nil {
		log.Fatalln(err)
	}

	config, err := config.NewConfigFromEnv()
	if err != nil {
		log.Fatalln(err)
	}

	compressedReposPath, err := worker.CreateTmpDir()
	if err != nil {
		log.Fatalln(err)
	}

	rabbitMQ, err := message_queue.NewRabbitMQ(config.MQ.URI)
	if err != nil {
		log.Fatalln(err)
	}
	defer rabbitMQ.Close(context.TODO())

	err = worker.Run(rabbitMQ, config.Worker.MaxWorkers, config.Worker.ReposPath, compressedReposPath)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("Press Ctrl+C to exit")
	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, os.Interrupt)
	<-signalCh
}
