# 跳板机 / 集群（AI）

> **联调阶段**文档。仅当 [DEVELOPMENT_PLAN.md](./DEVELOPMENT_PLAN.md) 文首为 **「开发完成，待联调」**（30 条 REQ 代码齐）后使用。开发阶段禁止 build/helm。见 G-006、[dev-then-integrate.mdc](../.cursor/rules/dev-then-integrate.mdc)。

凭证：`.cursor/rules/ops-bastion.mdc`。**写代码在本机**，下列在跳板机执行。

## 登录

```bash
# 本机
ssh -i "/Users/zhao/Documents/密钥/ops-private.pem" -o StrictHostKeyChecking=no root@10.51.6.240

# 跳板机
export KUBECONFIG=/root/.kube/gf-ops.config
cd /root/max/aiops
git pull
```

| 本机代码 | 跳板机 |
|----------|--------|
| `/Users/zhao/item/AIOps` | `/root/max/aiops` |

## 何时来跳板机

构建推送镜像、Helm 升级、kubectl 排障、REQ-021、MVP 现网 https://aiops.tclpv.com  — 本机 `go test` 不够时不要标 `done`。

## 构建（改哪个服务 build 哪个）

```bash
cd /root/max/aiops
TAG=latest PUSH=true
IMAGE_NAME=aiops-chat TAG=$TAG PUSH=true ./buildall.sh
# 全量 11 镜像：gateway platform chat policy executor plugin-mock plugin-kubernetes
#   plugin-prometheus plugin-logs module-kb worker
# 或：PUSH=true ./buildall.sh
```

镜像仓库前缀见 `helm/values.yaml`；`TAG` 与 Helm 该服务 `image.tag` 一致。

## Helm

```bash
helm upgrade --install aiops ./helm -n aiops --create-namespace
kubectl get jobs -n aiops    # migrate
kubectl get deploy,pods -n aiops
kubectl logs -n aiops deploy/aiops-chat --tail=100
```

## 验收 curl

```bash
# 集群内
kubectl run t --rm -it --restart=Never -n aiops --image=curlimages/curl -- \
  curl -sf http://aiops-policy:8083/healthz

# 对外（gateway）
kubectl port-forward -n aiops svc/aiops-gateway 8080:8080
curl http://127.0.0.1:8080/healthz
```

现网：Admin https://aiops.tclpv.com/admin/login/（`admin` / `admin@123`）；API `https://aiops.tclpv.com/api/v1/`。

## REQ-021

```bash
kubectl get deploy,sa -n aiops | grep plugin-kubernetes
kubectl logs -n aiops deploy/aiops-plugin-kubernetes --tail=50
```

## 禁止

未 `git pull` 就 build；编造 kubectl 输出；把海量日志写入 MySQL。
