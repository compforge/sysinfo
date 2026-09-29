GO ?= go
DIST_DIR ?= dist
VERSION := $(shell sed -n 's/^const Version = "\([^"]*\)"/\1/p' version.go)
ARTIFACT_PREFIX := sysinfo-$(VERSION)
TEST_PACKAGES ?= ./...
TEST_WORKERS ?= 4
LINT_WORKERS ?= 4

.DEFAULT_GOAL := build
.PHONY: fix lint test build build-all build-linux-amd64 build-linux-arm64
fix:
	$(GO) fmt ./...

lint:
	@test -z "$$(gofmt -l $$(find . -name '*.go' -not -path './.git/*'))" || { echo 'Run make fix'; exit 1; }
	$(GO) vet -p $(LINT_WORKERS) ./...
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) vet -p $(LINT_WORKERS) ./...
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GO) vet -p $(LINT_WORKERS) ./...

test:
	$(GO) test -p $(TEST_WORKERS) -parallel 1 $(TEST_PACKAGES)

# One static binary per architecture serves supported Linux distributions and
# kernels. Do not create kernel/glibc-labelled copies of the same executable.
build: build-linux-amd64 build-linux-arm64

build-all: build

build-linux-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOAMD64=v1 $(GO) build -trimpath -ldflags='-s -w' -o "$(DIST_DIR)/$(ARTIFACT_PREFIX)-linux-amd64" ./cmd/sysinfo

build-linux-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 GOARM64=v8.0 $(GO) build -trimpath -ldflags='-s -w' -o "$(DIST_DIR)/$(ARTIFACT_PREFIX)-linux-arm64" ./cmd/sysinfo
