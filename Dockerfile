ARG BUILDPLATFORM
FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.26.1-alpine3.23 AS builder

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
        -ldflags "${LDFLAGS}" \
        -o neko \
        ./cmd/neko
RUN setcap cap_net_raw=+ep neko

FROM --platform=${BUILDPLATFORM:-linux/amd64} alpine:3.23.3 AS runtime
WORKDIR /app

COPY --from=builder /app/neko .
ENTRYPOINT [ "/app/neko" ]