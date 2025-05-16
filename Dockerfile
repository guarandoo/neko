ARG BUILDPLATFORM
FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.24.2-alpine AS builder

ARG TARGETARCH
RUN --mount=type=cache,id=apk-${TARGETARCH},sharing=locked,target=/var/cache/apk \
    apk add --no-cache libcap

WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,id=go-mod,sharing=locked,target=/go/pkg/mod \
    go mod download

COPY . .
ARG TARGETOS
ARG LDFLAGS
RUN --mount=type=cache,id=go-mod,sharing=locked,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
        -ldflags "-X \"${LDFLAGS}\"" \
        -o neko \
        ./cmd/neko
RUN setcap cap_net_raw=+ep neko

FROM --platform=${BUILDPLATFORM:-linux/amd64} alpine@sha256:a8560b36e8b8210634f77d9f7f9efd7ffa463e380b75e2e74aff4511df3ef88c AS runtime
WORKDIR /app

COPY --from=builder /app/neko .
ENTRYPOINT [ "/app/neko" ]