# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=app
BINARY_DIR=bin
BINARY_UNIX=$(BINARY_NAME)_unix

all: test build

build: clean
	$(GOBUILD) -o $(BINARY_DIR)/$(BINARY_NAME) -v ./main.go

test:
	$(GOTEST) -v ./...

clean:
	$(GOCLEAN)
	rm -rf $(BINARY_DIR)

run: clean build
	$(BINARY_DIR)/$(BINARY_NAME)

deps:
	$(GOGET) -v -d ./...

tidy:
	$(GOMOD) tidy

# Cross compilation
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_DIR)/$(BINARY_UNIX) -v

docker-build:
	docker build -t $(BINARY_DIR)/$(BINARY_NAME):latest .

# Docker Compose commands
docker-up:
	docker-compose up --build -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

docker-ps:
	docker-compose ps

.PHONY: all build test clean run deps tidy build-linux docker-build docker-up docker-down docker-logs docker-ps