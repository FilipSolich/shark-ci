MODULE = github.com/FilipSolich/shark-ci

CI_SERVER      = ci-server
CI_SERVER_PATH = $(MODULE)/cmd/ci-server
WORKER         = worker
WORKER_PATH    = $(MODULE)/cmd/worker

BIN=bin

.PHONY: all build build-ci-server build-worker clean

all: build

build: build-ci-server build-worker

build-ci-server:
	go build -o $(BIN)/$(CI_SERVER) $(CI_SERVER_PATH)

build-worker:
	go build -o $(BIN)/$(WORKER) $(WORKER_PATH)

clean:
	go clean
	rm -rf $(BIN)
