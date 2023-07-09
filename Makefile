MODULE = github.com/FilipSolich/shark-ci

CI_SERVER      = ci-server
CI_SERVER_PATH = $(MODULE)/cmd/ci-server
WORKER         = worker
WORKER_PATH    = $(MODULE)/cmd/worker

BIN=bin

.PHONY: all build build-ci-server build-worker run-ci-server run-worker clean docker-build docker-build-ci-server docker-build-worker

all: build

$(BIN)/$(CI_SERVER): build-ci-server
$(BIN)/$(WORKER): build-worker

build: build-ci-server build-worker

build-ci-server:
	go build -o $(BIN)/$(CI_SERVER) $(CI_SERVER_PATH)

build-worker:
	go build -o $(BIN)/$(WORKER) $(WORKER_PATH)

run-ci-server: $(BIN)/$(CI_SERVER)
	$(BIN)/$(CI_SERVER)

run-worker: $(BIN)/$(WORKER)
	$(BIN)/$(WORKER)

clean:
	go clean
	rm -rf $(BIN)

docker-build: docker-build-ci-server docker-build-worker

docker-build-ci-server:
	docker build -f Dockerfile.ci-server -t filipsolich/ci-server .

docker-build-worker:
	docker build -f Dockerfile.worker -t filipsolich/worker .
