FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.22.3-alpine as builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

RUN apk update && apk add libcap

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o neko ./cmd/server
RUN setcap cap_net_raw=+ep neko

FROM --platform=${BUILDPLATFORM:-linux/amd64} alpine:3.19
WORKDIR /app
RUN apk update && apk add libcap

COPY --from=builder /app/neko .
CMD /app/neko