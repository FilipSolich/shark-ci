package main

import (
	"github.com/shark-ci/shark-ci/mq"
	"github.com/shark-ci/shark-ci/runner"
)

func main() {
	mq.InitMQ("localhost", "5672", "user", "pass")
	runner.Run()
}
