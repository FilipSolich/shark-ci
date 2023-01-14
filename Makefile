MODULE=github.com/shark-ci/shark-ci

CI_SERVER=ci-server
CI_SERVER_PATH=$(MODULE)/cmd/ci-server
RUNNER=runner
RUNNER_PATH=$(MODULE)/cmd/runner

BIN=bin

.PHONY: all build build-ci-server build-runner clean

all: build

build: build-ci-server build-runner

build-ci-server:
	go build -o $(BIN)/$(CI_SERVER) $(CI_SERVER_PATH)

build-runner:
	go build -o $(BIN)/$(RUNNER) $(RUNNER_PATH)

clean:
	go clean
	rm -rf $(BIN)
