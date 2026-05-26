# syntax=docker/dockerfile:1
# 示例: docker build --build-arg SERVICE=gateway --target runtime -t aiops-gateway .
# platform: docker build --build-arg SERVICE=platform --target runtime-with-migrate -t aiops-platform .

FROM golang:1.21-alpine AS builder

ARG APK_MIRROR=https://mirrors.aliyun.com/alpine
ARG GOPROXY=https://goproxy.cn,direct
ARG GOSUMDB=sum.golang.google.cn
ARG SERVICE=gateway

RUN sed -i "s|https://dl-cdn.alpinelinux.org/alpine|${APK_MIRROR}|g" /etc/apk/repositories && \
    apk add --no-cache git ca-certificates tzdata

ENV GOPROXY=${GOPROXY} \
    GOSUMDB=${GOSUMDB} \
    GOMODCACHE=/go/pkg/mod \
    GOCACHE=/root/.cache/go-build

WORKDIR /build

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /build/server ./cmd/${SERVICE}/ && \
    if [ "${SERVICE}" = "platform" ]; then \
      CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /build/migrate ./cmd/migrate/; \
    fi

FROM alpine:3.19 AS runtime

COPY --from=builder /etc/ssl/certs/ /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo/Asia/Shanghai /usr/share/zoneinfo/Asia/Shanghai
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

ENV TZ=Asia/Shanghai
WORKDIR /app

COPY --from=builder /build/server ./server
COPY config/config.yaml ./config/config.yaml
COPY templates ./templates

ARG SERVICE=gateway
ENV GIN_MODE=release \
    AIOPS_SERVICE=${SERVICE} \
    EWA_CONFIG=/app/config/config.yaml

EXPOSE 8080

CMD ["./server"]

FROM runtime AS runtime-with-migrate

COPY --from=builder /build/migrate ./migrate
