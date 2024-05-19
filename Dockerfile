FROM golang:1.22.3-alpine AS builder
ENV GOOS linux
ENV CGO_ENABLED 0

RUN apk update && apk add libcap

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o neko ./cmd/server
RUN setcap cap_net_raw=+ep neko

FROM alpine:3.19
WORKDIR /app
RUN apk update && apk add libcap

COPY --from=builder /app/neko .
CMD /app/neko