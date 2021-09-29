# Binary
TAG ?= dev-local
BUILD_HASH := $(shell git rev-parse HEAD)
BUILD_TIME := $(shell date -u +%Y%m%d.%H%M%S)
LDFLAGS := '-s -w -X main.BuildVersion=${BUILD_HASH} -X main.BuildTime=${BUILD_TIME} -linkmode external -extldflags "-static"'

## Golang
GO ?= go
GO_TEST_FLAGS ?= -race

GOLANGCILINT_VER := v1.41.1
GOLANGCILINT_BIN := golangci-lint
GOLANGCILINT_GEN := $(TOOLS_BIN_DIR)/$(GOLANGCILINT_BIN)

## Binaries.
GO_INSTALL = ./scripts/go_install.sh
TOOLS_BIN_DIR := $(abspath bin)

OUTDATED_VER := master
OUTDATED_BIN := go-mod-outdated
OUTDATED_GEN := $(TOOLS_BIN_DIR)/$(OUTDATED_BIN)

## Docker
CHAOS_ENGINE_IMAGE ?= mattermost/mattermost-app-chaosengine:test

## Docker Build Versions
DOCKER_BUILD_IMAGE = golang:1.16.8
DOCKER_BASE_IMAGE = alpine:3.14.2
################################################################################

.PHONY: all
## all: builds and runs the service
all: run

.PHONY: build-image
## build-image: builds the docker image
build-image:
	@echo Building Chaos Engine Docker Image
	docker build \
	--build-arg DOCKER_BUILD_IMAGE=$(DOCKER_BUILD_IMAGE) \
	--build-arg DOCKER_BASE_IMAGE=$(DOCKER_BASE_IMAGE) \
	. -f build/Dockerfile -t $(CHAOS_ENGINE_IMAGE)

.PHONY: build-linux
## build-linux: builds linux binary
build-linux:
	@echo Building binary for linux
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 $(GO) build -ldflags $(LDFLAGS) -gcflags all=-trimpath=$(PWD) -asmflags all=-trimpath=$(PWD) -a -installsuffix cgo -o build/_output/bin/mattermost-app-chaosengine-linux-amd64 ./cmd

.PHONY: build
## build: build the executable
build:
	@echo Building for local use only
	$(GO) build -o build/_output/bin/mattermost-app-chaosengine ./cmd

.PHONY: check-modules
## check-modules: Check outdated modules
check-modules: $(OUTDATED_GEN) #
	@echo Checking outdated modules
	$(GO) list -u -m -json all | $(OUTDATED_GEN) -update -direct

.PHONY: check-style
## check-style: Runs govet and gofmt against all packages.
check-style: govet lint
	@echo Checking for style guide compliance

.PHONY: clean
## clean: deletes all
clean:
	rm -rf build/_output/bin/

.PHONY: vet
## govet: Runs govet against all packages.
govet:
	@echo Running govet
	$(GO) vet ./...
	@echo Govet success

.PHONY: push-docker-pr
## push-docker-pr: Pushes the Docker image for the particular PR
push-docker-pr:
	@echo Pushing Docker Image for pull request
	sh -c "./scripts/push_docker_pr.sh"

.PHONY: lint
## lint: Run golangci-lint on codebase
lint: $(GOLANGCILINT_GEN)
	@echo Running lint with GolangCI
	@if ! [ -x "$$(command -v golangci-lint)" ]; then \
		echo "golangci-lint is not installed. Please see https://github.com/golangci/golangci-lint#install for installation instructions."; \
		exit 1; \
	fi; \

	@echo Running golangci-lint
	golangci-lint run ./...

.PHONY: push-docker
## push-docker: Pushes the Docker image 
push-docker:
	@echo Pushing Docker Image
	sh -c "./scripts/push_docker.sh"

.PHONY: run
## run: runs the service
run: build
	@echo Running chaos engine with debug
	CHAOS_ENGINE_DEBUG=true build/_output/bin/mattermost-app-chaosengine

.PHONY: test
## test: tests all packages
test:
	@echo Running tests
	$(GO) test $(GO_TEST_FLAGS) ./...

.PHONY: help
## help: prints this help message
help:
	@echo "Usage:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

## --------------------------------------
## Tooling Binaries
## --------------------------------------

$(OUTDATED_GEN): ## Build go-mod-outdated.
	GOBIN=$(TOOLS_BIN_DIR) $(GO_INSTALL) github.com/psampaz/go-mod-outdated $(OUTDATED_BIN) $(OUTDATED_VER)

$(GOLANGCILINT_GEN): ## Build golang-ci lint.
	GOBIN=$(TOOLS_BIN_DIR) $(GO_INSTALL) github.com/golangci/golangci-lint/cmd/golangci-lint $(GOLANGCILINT_BIN) $(GOLANGCILINT_VER)
