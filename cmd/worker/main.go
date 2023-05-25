package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/shark-ci/shark-ci/message_queue"
	"github.com/shark-ci/shark-ci/worker"
)

func main() {
	rabbitMQ, err := message_queue.NewRabbitMQ("localhost", "5672", "user", "pass")
	if err != nil {
		log.Fatalln(err)
	}
	defer rabbitMQ.Close(context.TODO())

	err = worker.Run(rabbitMQ)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("Press Ctrl+C to exit")
	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, os.Interrupt)
	<-signalCh
}
