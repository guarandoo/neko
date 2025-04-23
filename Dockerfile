FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.22.3-alpine AS builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

RUN --mount=type=cache,id=apk-${TARGETARCH},sharing=locked,target=/var/cache/apk \
    apk add --no-cache libcap

WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,id=go-mod,sharing=locked,target=/go/pkg/mod \
    go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o neko ./cmd/server
RUN setcap cap_net_raw=+ep neko

FROM --platform=${BUILDPLATFORM:-linux/amd64} alpine:3.21
WORKDIR /app

COPY --from=builder /app/neko .
ENTRYPOINT [ "/app/neko" ]