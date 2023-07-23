MODULE = github.com/FilipSolich/shark-ci

CI_SERVER      = ci-server
CI_SERVER_PATH = $(MODULE)/cmd/ci-server
WORKER         = worker
WORKER_PATH    = $(MODULE)/cmd/worker

BIN=bin

.PHONY: all
all: build

$(BIN)/$(CI_SERVER): build-ci-server
$(BIN)/$(WORKER): build-worker

.PHONY: build
build: build-ci-server build-worker

.PHONY: build-ci-server
build-ci-server:
	go build -o $(BIN)/$(CI_SERVER) $(CI_SERVER_PATH)

.PHONY: build-worker
build-worker:
	go build -o $(BIN)/$(WORKER) $(WORKER_PATH)

.PHONY: run-ci-server
run-ci-server: $(BIN)/$(CI_SERVER)
	$(BIN)/$(CI_SERVER)

.PHONY: run-worker
run-worker: $(BIN)/$(WORKER)
	$(BIN)/$(WORKER)

.PHONY: clean
clean:
	go clean
	rm -rf $(BIN)

.PHONY: docker-build
docker-build: docker-build-ci-server docker-build-worker

.PHONY: docker-build-ci-server
docker-build-ci-server:
	docker build -f Dockerfile.ci-server -t filipsolich/ci-server .

.PHONY: docker-build-worker
docker-build-worker:
	docker build -f Dockerfile.worker -t filipsolich/worker .

.PHONY: create-migration
create-migration:
	migrate create -ext sql -dir migrations -format 20060102150405 $(NAME)
