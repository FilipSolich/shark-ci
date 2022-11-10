package main

import (
	"github.com/FilipSolich/ci-server/mq"
	"github.com/FilipSolich/ci-server/runner"
)

func main() {
	mq.InitMQ("localhost", "5672", "user", "pass")
	runner.Run()
}
