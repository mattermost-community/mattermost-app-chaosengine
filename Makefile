# Binary
TAG ?= dev-local
LDFLAGS := -ldflags "-s -w -X main.BuildVersion=${TAG} -X main.BuildTime=$(shell date +%s)"
################################################################################

# Golang
GO ?= go
GO_TEST_FLAGS ?= -race

# Binaries.
TOOLS_BIN_DIR := $(abspath bin)

OUTDATED_VER := master
OUTDATED_BIN := go-mod-outdated
OUTDATED_GEN := $(TOOLS_BIN_DIR)/$(OUTDATED_BIN)
################################################################################

.PHONY: all
## all: builds and runs the service
all: run

.PHONY: build-linux
## build-linux: builds linux binary
build-linux:
	@echo Building binary for linux
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GO) build ${LDFLAGS} -o build/_output/bin/mattermost-app-chaosengine-linux-amd64 ./cmd

.PHONY: build-mac
## build-mac: builds mac binary
build-mac:
	@echo Building binary for mac
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 $(GO) build ${LDFLAGS} -o build/_output/bin/mattermost-app-chaosengine-darwin-amd64 ./cmd

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

.PHONY: vet
## govet: Runs govet against all packages.
govet:
	@echo Running govet
	$(GO) vet ./...
	@echo Govet success

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

.PHONY: lint
## lint: Run golangci-lint on codebase
lint:
	@echo Running lint with GolangCI
	@if ! [ -x "$$(command -v golangci-lint)" ]; then \
		echo "golangci-lint is not installed. Please see https://github.com/golangci/golangci-lint#install for installation instructions."; \
		exit 1; \
	fi; \

	@echo Running golangci-lint
	golangci-lint run ./...

.PHONY: clean
## clean: deletes all
clean:
	rm -rf build/_output/bin/

.PHONY: help
## help: prints this help message
help:
	@echo "Usage:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'
