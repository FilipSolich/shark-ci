APP_NAME=ci-server
TARGET=main

DOCKER_USERNAME=filipsolich

.PHONY: build run test clean build-docker tag

build:
	go fmt ./...
	go build

run:
	go run $(TARGET).go

test:
	go test ./...

clean:
	rm $(APP_NAME)

build-docker:
	docker build -t $(DOCKER_USERNAME)/$(APP_NAME):latest .

tag:
	git tag $(VERSION)
