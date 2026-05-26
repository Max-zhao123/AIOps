#!/bin/bash
#
# AIOps 多服务镜像构建
# 用法:
#   ./buildall.sh                                    # 构建全部 11 个业务镜像
#   IMAGE_NAME=aiops-gateway TAG=v1 ./buildall.sh    # 仅构建 gateway
#   REGISTRY=harbor.example.com/proj/ TAG=latest PUSH=true ./buildall.sh
#

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "${SCRIPT_DIR}"

REGISTRY="${REGISTRY:-swr.cn-south-1.myhuaweicloud.com/ops-images/}"
TAG="${TAG:-latest}"
IMAGE_NAME="${IMAGE_NAME:-}"
PLATFORM="${PLATFORM:-linux/amd64}"
PUSH="${PUSH:-true}"
APK_MIRROR="${APK_MIRROR:-https://mirrors.huaweicloud.com/alpine}"
GOPROXY="${GOPROXY:-https://goproxy.cn,direct}"
GOSUMDB="${GOSUMDB:-sum.golang.google.cn}"
export DOCKER_BUILDKIT=1

if [[ -n "${REGISTRY}" && "${REGISTRY}" != */ ]]; then
  REGISTRY="${REGISTRY}/"
fi

# IMAGE_NAME -> SERVICE（cmd 目录名）
image_to_service() {
  case "$1" in
    aiops-gateway) echo "gateway" ;;
    aiops-platform) echo "platform" ;;
    aiops-chat) echo "chat" ;;
    aiops-policy) echo "policy" ;;
    aiops-executor) echo "executor" ;;
    aiops-plugin-mock) echo "plugin-mock" ;;
    aiops-plugin-kubernetes) echo "plugin-kubernetes" ;;
    aiops-plugin-prometheus) echo "plugin-prometheus" ;;
    aiops-plugin-logs) echo "plugin-logs" ;;
    aiops-module-kb) echo "module-kb" ;;
    aiops-worker) echo "worker" ;;
    aiops) echo "platform" ;;
    *) echo "" ;;
  esac
}

ALL_IMAGES=(
  aiops-gateway
  aiops-platform
  aiops-chat
  aiops-policy
  aiops-executor
  aiops-plugin-mock
  aiops-plugin-kubernetes
  aiops-plugin-prometheus
  aiops-plugin-logs
  aiops-module-kb
  aiops-worker
)

build_one() {
  local img="$1"
  local svc
  svc="$(image_to_service "${img}")"
  if [[ -z "${svc}" ]]; then
    echo "错误: 未知 IMAGE_NAME=${img}" >&2
    exit 1
  fi

  local target="runtime"
  if [[ "${svc}" == "platform" ]]; then
    target="runtime-with-migrate"
  fi

  local full_image="${REGISTRY}${img}:${TAG}"
  echo "------------------------------------------"
  echo "构建 ${full_image} (SERVICE=${svc}, target=${target})"
  echo "------------------------------------------"

  docker build \
    --platform "${PLATFORM}" \
    --build-arg APK_MIRROR="${APK_MIRROR}" \
    --build-arg GOPROXY="${GOPROXY}" \
    --build-arg GOSUMDB="${GOSUMDB}" \
    --build-arg SERVICE="${svc}" \
    --target "${target}" \
    -f Dockerfile \
    -t "${full_image}" \
    .

  if [[ "${PUSH}" == "true" ]]; then
    if [[ -z "${REGISTRY}" ]]; then
      echo "错误: PUSH=true 时必须设置 REGISTRY" >&2
      exit 1
    fi
    docker push "${full_image}"
    echo "已推送 ${full_image}"
  fi
}

if ! command -v docker &>/dev/null; then
  echo "错误: 未找到 docker" >&2
  exit 1
fi

echo "=========================================="
echo " AIOps 多服务镜像构建"
echo " REGISTRY=${REGISTRY} TAG=${TAG}"
echo "=========================================="

if [[ -n "${IMAGE_NAME}" ]]; then
  build_one "${IMAGE_NAME}"
else
  for img in "${ALL_IMAGES[@]}"; do
    build_one "${img}"
  done
fi

echo ""
echo "Helm 部署:"
echo "  helm upgrade --install aiops ./helm -n aiops --create-namespace"
