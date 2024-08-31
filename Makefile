.DEFAULT_GOAL := default

TAG ?= latest
IMAGE ?= gitea.calliope.rip/guarandoo/neko:$(TAG)

export DOCKER_CLI_EXPERIMENTAL=enabled

.PHONY: build
build:
	CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
	-o neko \
	./cmd/server

.PHONY: build-image
build-image:
	docker buildx build \
	--output "type=docker,push=false" \
	--tag $(IMAGE) \
	.

.PHONY: docker
docker:
	docker buildx build \
	--platform linux/386,linux/amd64,linux/arm/v6,linux/arm/v7,linux/arm64,linux/ppc64le,linux/s390x \
	--output "type=image,push=true" \
	--tag $(IMAGE) \
	.