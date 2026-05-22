#!/bin/bash
#
# aiops 镜像构建脚本
# 用法:
#   ./buildall.sh                    # 本地构建，标签 aiops:latest
#   REGISTRY=harbor.example.com/proj/ TAG=v1.0.0 ./buildall.sh
#   PUSH=true REGISTRY=... TAG=... ./buildall.sh   # 构建并推送
#

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "${SCRIPT_DIR}"

# 镜像仓库前缀，如 harbor.example.com/aiops/ 需以 / 结尾
REGISTRY="${REGISTRY:-swr.cn-south-1.myhuaweicloud.com/ops-images/}"
TAG="${TAG:-latest}"
IMAGE_NAME="${IMAGE_NAME:-aiops}"
PLATFORM="${PLATFORM:-linux/amd64}"
PUSH="${PUSH:-true}"

if [[ -n "${REGISTRY}" && "${REGISTRY}" != */ ]]; then
  REGISTRY="${REGISTRY}/"
fi

FULL_IMAGE="${REGISTRY}${IMAGE_NAME}:${TAG}"

echo "=========================================="
echo " aiops 镜像构建"
echo " 镜像: ${FULL_IMAGE}"
echo " 平台: ${PLATFORM}"
echo "=========================================="

if ! command -v docker &>/dev/null; then
  echo "错误: 未找到 docker 命令，请先安装 Docker" >&2
  exit 1
fi

echo "[1/2] 构建应用镜像..."
docker build \
  --platform "${PLATFORM}" \
  -f Dockerfile \
  -t "${FULL_IMAGE}" \
  .

echo "[2/2] 构建完成: ${FULL_IMAGE}"
docker images "${FULL_IMAGE}" --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}\t{{.CreatedAt}}"

if [[ "${PUSH}" == "true" ]]; then
  if [[ -z "${REGISTRY}" ]]; then
    echo "错误: PUSH=true 时必须设置 REGISTRY（镜像仓库地址）" >&2
    exit 1
  fi
  echo "推送镜像到仓库..."
  docker push "${FULL_IMAGE}"
  echo "推送完成: ${FULL_IMAGE}"
fi

echo ""
echo "Helm 部署示例:"
echo "  helm upgrade --install aiops ./helm/aiops \\"
echo "    --set image.repository=${REGISTRY}${IMAGE_NAME} \\"
echo "    --set image.tag=${TAG}"
