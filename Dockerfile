# syntax=docker/dockerfile:1

FROM golang:1.21-alpine AS builder

# 国内/内网构建建议指定镜像源，例如:
#   docker build --build-arg APK_MIRROR=https://mirrors.aliyun.com/alpine .
#   docker build --build-arg APK_MIRROR=https://mirrors.huaweicloud.com/alpine .
ARG APK_MIRROR=https://mirrors.aliyun.com/alpine
# Go 模块代理（默认 goproxy.cn，避免走 proxy.golang.org）
ARG GOPROXY=https://goproxy.cn,direct
ARG GOSUMDB=sum.golang.google.cn

RUN sed -i "s|https://dl-cdn.alpinelinux.org/alpine|${APK_MIRROR}|g" /etc/apk/repositories && \
    apk add --no-cache git ca-certificates tzdata

ENV GOPROXY=${GOPROXY} \
    GOSUMDB=${GOSUMDB} \
    GOMODCACHE=/go/pkg/mod \
    GOCACHE=/root/.cache/go-build

WORKDIR /build

# 仅 go.mod/go.sum 变化时才重新 download（配合 BuildKit 缓存更快）
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /build/bin/aiops ./cmd/main.go && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /build/bin/migrate ./cmd/migrate/main.go

FROM alpine:3.19

# 运行时不再 apk install，避免 stage-1 访问国外 CDN 失败
COPY --from=builder /etc/ssl/certs/ /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo/Asia/Shanghai /usr/share/zoneinfo/Asia/Shanghai
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

ENV TZ=Asia/Shanghai
WORKDIR /app

COPY --from=builder /build/bin/aiops /build/bin/migrate ./
COPY config/config.yaml ./config/config.yaml
COPY templates ./templates

ENV GIN_MODE=release
ENV EWA_CONFIG=/app/config/config.yaml

EXPOSE 8080

CMD ["./aiops"]
