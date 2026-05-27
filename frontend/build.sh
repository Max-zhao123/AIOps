#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "${SCRIPT_DIR}"

REGISTRY="${REGISTRY:-swr.cn-south-1.myhuaweicloud.com/ops-images/}"
TAG="${TAG:-latest}"
PLATFORM="${PLATFORM:-linux/amd64}"
PUSH="${PUSH:-true}"
IMAGE="${REGISTRY}aiops-frontend:${TAG}"

if [[ -n "${REGISTRY}" && "${REGISTRY}" != */ ]]; then
  REGISTRY="${REGISTRY}/"
  IMAGE="${REGISTRY}aiops-frontend:${TAG}"
fi

echo "===== 构建前端镜像: ${IMAGE} ====="

docker build --platform "${PLATFORM}" -t "${IMAGE}" .

if [[ "${PUSH}" == "true" ]]; then
  docker push "${IMAGE}"
  echo "已推送 ${IMAGE}"
fi

echo "完成!"
