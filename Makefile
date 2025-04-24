IMAGE ?= gitea.calliope.rip/guarandoo/neko

GO := go
MAKE := make
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

DOCKER_BUILD_PLATFORMS ?= linux/386,linux/amd64,linux/arm/v6,linux/arm/v7,linux/arm64
ifeq ($(GOOS), windows)
BIN_SUFFIX = .exe
endif
BIN_NAME := neko$(BIN_SUFFIX)

export DOCKER_CLI_EXPERIMENTAL=enabled

.PHONY: default
default: binary

test:
	go test ./...

binary:
	CGO_ENABLED=0 go build \
	  -o ./dist/${GOOS}/${GOARCH}/${BIN_NAME} \
	  ./cmd/server

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

docker-image: export DOCKER_BUILDX_ARGS := --load
docker-image: export DOCKER_BUILD_PLATFORMS := $(GOOS)/$(GOARCH)
docker-image:
	$(MAKE) docker-latest

docker-multiarch-image:
	$(MAKE) docker-latest

publish-docker-multiarch-image: export DOCKER_BUILDX_ARGS := --push
publish-docker-multiarch-image:
	$(MAKE) docker-latest

docker-%:
	docker buildx build $(DOCKER_BUILDX_ARGS) -t $(IMAGE):$* --platform $(DOCKER_BUILD_PLATFORMS) -f Dockerfile .
