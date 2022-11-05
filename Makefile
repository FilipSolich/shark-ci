CI_SERVER="ci-server"
CI_SERVER_PATH="cmd/ci-server/ci-server.go"
TARGET=main

DOCKER_USERNAME=filipsolich

.PHONY: all build run-ci-server test clean build-docker tag

all: build

build:
	go fmt ./...
	go build $(CI_SERVER_PATH)

run-ci-server: build
	./$(CI_SERVER)

test:
	go test ./...

clean:
	rm $(CI_SERVER)

#build-docker:
#	docker build -t $(DOCKER_USERNAME)/$(APP_NAME):latest .
#
#tag:
#	git tag $(VERSION)
