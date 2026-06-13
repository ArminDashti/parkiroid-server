# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /server ./cmd/server

FROM alpine:3.21

RUN apk add --no-cache ca-certificates wget \
    && adduser -D -h /data -u 1000 appuser

WORKDIR /data

COPY --from=builder /server /usr/local/bin/server

USER appuser

EXPOSE 8080

ENV GIN_MODE=release \
    PARKIROID_LISTEN_ADDRESS=:8080 \
    PARKIROID_DATABASE_PATH=/data/parkiroid.db \
    PARKIROID_FRAMES_DIR=/data/frames

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8080/parkiroid/api/v1/health || exit 1

ENTRYPOINT ["server"]
