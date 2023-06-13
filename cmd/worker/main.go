package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/joho/godotenv"
	"github.com/shark-ci/shark-ci/config"
	"github.com/shark-ci/shark-ci/message_queue"
	"github.com/shark-ci/shark-ci/worker"
)

func main() {
	// TODO: Remove.
	err := godotenv.Load()
	if err != nil {
		log.Fatalln(err)
	}

	config, err := config.NewWorkerConfigFromEnv()
	if err != nil {
		log.Fatalln(err)
	}

	rabbitMQ, err := message_queue.NewRabbitMQ(config.RabbitMQURI)
	if err != nil {
		log.Fatalln(err)
	}
	defer rabbitMQ.Close(context.TODO())

	err = worker.Run(rabbitMQ, config.MaxWorkers, config.ReposPath)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("Press Ctrl+C to exit")
	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, os.Interrupt)
	<-signalCh
}
