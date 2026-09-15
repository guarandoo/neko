NAME := neko
IMAGE ?= docker.io/guarandoo/$(NAME)
TAG ?= latest

GO := go
MAKE := make
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
DOCKER := $(shell command -v podman >/dev/null 2>&1 && echo podman || echo docker)

DOCKER_BUILD_PLATFORMS ?= linux/386,linux/amd64,linux/arm/v6,linux/arm/v7,linux/arm64
ifeq ($(GOOS), windows)
BIN_SUFFIX := .exe
endif
BIN_NAME := neko$(BIN_SUFFIX)

VERSION := $(shell git describe --tags --abbrev=0)
COMMIT := $(shell git describe --always --dirty --exclude='*')
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -s -w -X 'main.Version=$(VERSION)' -X 'main.Commit=$(COMMIT)' -X 'main.BuildTime=$(BUILD_TIME)'

export DOCKER_CLI_EXPERIMENTAL=enabled

DOCKER_BUILD_PLATFORMS := $(GOOS)/$(GOARCH)
DOCKER_VARIANT :=

.PHONY: default
default: binary

test:
	$(GO) test ./...

binary:
	CGO_ENABLED=0 $(GO) build \
	  -ldflags "$(LDFLAGS)" \
	  -o ./dist/$(GOOS)/$(GOARCH)/$(BIN_NAME) \
	  ./cmd/neko

run:
	$(GO) run \
		-ldflags "$(LDFLAGS)" \
		./...

binary-linux-386: export GOOS := linux
binary-linux-386: export GOARCH := 386
binary-linux-386:
	$(MAKE) binary

binary-linux-amd64: export GOOS := linux
binary-linux-amd64: export GOARCH := amd64
binary-linux-amd64:
	$(MAKE) binary

binary-windows-amd64: export GOOS := windows
binary-windows-amd64: export GOARCH := amd64
binary-windows-amd64: export BIN_NAME := neko.exe
binary-windows-amd64:
	$(MAKE) binary

binary-darwin-amd64: export GOOS := darwin
binary-darwin-amd64: export GOARCH := amd64
binary-darwin-amd64:
	$(MAKE) binary

all-binaries: binary-windows-amd64 binary-linux-386 binary-linux-amd64 binary-darwin-amd64

docker-image: 
docker-image:
	$(DOCKER) buildx build \
		-t $(IMAGE):$(TAG) \
		--platform $(DOCKER_BUILD_PLATFORMS) \
		--build-arg LDFLAGS="$(LDFLAGS)" \
		-f Dockerfile$(if $(strip $(DOCKER_VARIANT)),\.,)$(DOCKER_VARIANT) \
		.

tag-docker-image:
tag-docker-image:
	@if ! git diff-index --quiet HEAD --; then \
		echo 'Working tree is dirty'; \
		exit 1; \
	fi; \
	if ! git describe --tags --exact-match HEAD > /dev/null 2>&1; then \
		echo 'HEAD does not match a tag'; \
		exit 1; \
	fi; \
	version="$(VERSION)"; \
	tag="$${version#v}$(if $(strip $(DOCKER_VARIANT)),-,)$(DOCKER_VARIANT)"; \
	$(DOCKER) tag $(IMAGE):latest $(IMAGE):$$tag; \
	echo $(IMAGE):$$tag
