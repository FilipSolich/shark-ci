MODULE = github.com/FilipSolich/shark-ci

CI_SERVER      = ci-server
CI_SERVER_PATH = $(MODULE)/cmd/ci-server
WORKER         = worker
WORKER_PATH    = $(MODULE)/cmd/worker

BIN=bin

.PHONY: all build build-ci-server build-worker run run-ci-server run-worker clean

all: build

build: build-ci-server build-worker

build-ci-server:
	go build -o $(BIN)/$(CI_SERVER) $(CI_SERVER_PATH)

build-worker:
	go build -o $(BIN)/$(WORKER) $(WORKER_PATH)

run: run-ci-server run-worker

run-ci-server:
	$(BIN)/$(CI_SERVER)

run-worker:
	$(BIN)/$(WORKER)

clean:
	go clean
	rm -rf $(BIN)
