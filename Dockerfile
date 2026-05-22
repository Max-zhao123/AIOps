# syntax=docker/dockerfile:1

FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /build/bin/freeaiops ./cmd/main.go && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /build/bin/migrate ./cmd/migrate/main.go

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

WORKDIR /app

COPY --from=builder /build/bin/freeaiops /build/bin/migrate ./
COPY config/config.yaml ./config/config.yaml
COPY templates ./templates

ENV GIN_MODE=release
ENV EWA_CONFIG=/app/config/config.yaml

EXPOSE 8080

CMD ["./freeaiops"]
