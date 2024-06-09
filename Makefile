.DEFAULT_GOAL := default

IMAGE ?= harbor.guarandoo.cloud/guarandoo/neko:latest

export DOCKER_CLI_EXPERIMENTAL=enabled

.PHONY: build
build:
	docker buildx build \
		--output "type=docker,push=false" \
		--tag $(IMAGE) \
		.

.PHONY publish
publish:
	docker buildx build \
		--platform linux/386,linux/amd64,linux/arm/v6,linux/arm/v7,linux/arm64,linux/ppc64le,linux/s390x \
		--output "type=image,push=true" \
		--tag $(IMAGE) \
		.