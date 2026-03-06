BINARY_NAME ?= navilist
BIN_DIR     ?= bin
VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS     := -ldflags="-s -w -X main.version=$(VERSION)"

.PHONY: build test test-coverage clean run fmt vet \
        build-all docker-build up down logs \
        version bump-patch bump-minor bump-major

## Build the binary
build:
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/server

## Run all tests with race detector
test:
	go test -v -race ./...

## Run tests and produce an HTML coverage report
test-coverage:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

## Remove build artifacts
clean:
	rm -f $(BINARY_NAME) coverage.out coverage.html
	rm -rf $(BIN_DIR)

## Run the server locally (requires env vars or a .env file)
run:
	go run $(LDFLAGS) ./cmd/server

## Format all Go source files
fmt:
	go fmt ./...

## Run go vet
vet:
	go vet ./...

## Cross-compile for Linux amd64, macOS amd64/arm64, Windows amd64
build-all:
	mkdir -p $(BIN_DIR)
	GOOS=linux   GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME)-linux-amd64    ./cmd/server
	GOOS=darwin  GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME)-darwin-amd64   ./cmd/server
	GOOS=darwin  GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME)-darwin-arm64   ./cmd/server
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/server

## Build the Docker image
docker-build:
	docker build --build-arg VERSION=$(VERSION) -t $(BINARY_NAME):$(VERSION) -t $(BINARY_NAME):latest .

## Start the stack with Docker Compose (builds first)
up:
	VERSION=$(VERSION) docker compose up -d --build

## Stop the stack
down:
	docker compose down

## Tail Docker Compose logs
logs:
	docker compose logs -f

## Print the current version (latest git tag)
version:
	@git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"

## Tag and push a patch release (v0.0.X+1)
bump-patch:
	@LATEST=$$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"); \
	MAJOR=$$(echo $$LATEST | sed 's/v//' | cut -d. -f1); \
	MINOR=$$(echo $$LATEST | sed 's/v//' | cut -d. -f2); \
	PATCH=$$(echo $$LATEST | sed 's/v//' | cut -d. -f3); \
	NEW="v$$MAJOR.$$MINOR.$$((PATCH + 1))"; \
	echo "$$LATEST -> $$NEW"; \
	git tag -a "$$NEW" -m "Release $$NEW"; \
	git push origin "$$NEW"

## Tag and push a minor release (v0.X+1.0)
bump-minor:
	@LATEST=$$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"); \
	MAJOR=$$(echo $$LATEST | sed 's/v//' | cut -d. -f1); \
	MINOR=$$(echo $$LATEST | sed 's/v//' | cut -d. -f2); \
	NEW="v$$MAJOR.$$((MINOR + 1)).0"; \
	echo "$$LATEST -> $$NEW"; \
	git tag -a "$$NEW" -m "Release $$NEW"; \
	git push origin "$$NEW"

## Tag and push a major release (vX+1.0.0)
bump-major:
	@LATEST=$$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"); \
	MAJOR=$$(echo $$LATEST | sed 's/v//' | cut -d. -f1); \
	NEW="v$$((MAJOR + 1)).0.0"; \
	echo "$$LATEST -> $$NEW"; \
	git tag -a "$$NEW" -m "Release $$NEW"; \
	git push origin "$$NEW"
