.PHONY: build run test test-race docker-up docker-down docker-logs mock clean help test-upload

# Variables
BINARY_NAME=streaming-api
DOCKER_COMPOSE=docker-compose
MAIN_PATH=cmd/server/main.go

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^## [-a-zA-Z0-9_]+:' Makefile | sed 's/## //g' | awk -F: '{printf "  %-15s %s
	", $$1, $$2}'

## build: Build the Go binary locally
build:
	go build -o bin/$(BINARY_NAME) $(MAIN_PATH)

## run: Run the application locally (expects local DB)
run:
	go run $(MAIN_PATH)

## test: Run all unit and integration tests
test:
	go test ./...

## test-race: Run tests with data race detection
test-race:
	go test -race ./...

## docker-up: Start all services in Docker and build
docker-up:
	$(DOCKER_COMPOSE) up --build -d

## docker-down: Stop and remove all Docker containers and volumes
docker-down:
	$(DOCKER_COMPOSE) down -v

docker-restart:
	$(DOCKER_COMPOSE) up --build -d app

## docker-logs: Follow application logs in Docker
docker-logs:
	$(DOCKER_COMPOSE) logs -f app

## mock: Regenerate all mocks using mockery
mock:
	mockery

## test-upload: Send a test video to the API using curl (Port 8085)
## make sure you have a test file in your local project root to run
test-upload:
	curl -v POST http://localhost:8085/api/video/ \
		-F "title=Test Video Title" \
		-F "description=Test Description" \
		-F "video=@test_video.mp4" \

## clean: Remove build artifacts and local storage files
clean:
	rm -rf bin/
	rm -rf storage/*
