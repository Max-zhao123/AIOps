# AIOps 架构

> 功能 REQ 与顺序：[DEVELOPMENT_PLAN.md](./DEVELOPMENT_PLAN.md)（**主文档**）。契约 [PRD.md](./PRD.md)。

## 1. Pod 一览

### 阶段 A（必部署）

| Deployment | 端口 | 镜像 | REQ |
|------------|------|------|-----|
| `aiops-gateway` | 8080 | `aiops-gateway` | 019 |
| `aiops-platform` | 8081 | `aiops-platform` | 001–003,010,014,015,050,061… |
| `aiops-chat` | 8082 | `aiops-chat` | 030–033,041 |
| `aiops-policy` | 8083 | `aiops-policy` | 011,013 |
| `aiops-executor` | 8084 | `aiops-executor` | 012,033,080 |
| `aiops-plugin-mock` | 8090 | `aiops-plugin-mock` | 020 |
| `aiops-plugin-kubernetes` | 8091 | `aiops-plugin-kubernetes` | 021 |
| `aiops-mysql` | 3306 | — | — |
| Job `aiops-migrate` | — | `aiops-platform` 镜像 `/app/migrate` | 019 |

Ingress → 仅 `aiops-gateway:8080`。gateway/platform **不合并**。

### 阶段 B+（开发完成后联调时按需启用）

| Deployment | 端口 | 说明 | REQ |
|------------|------|------|-----|
| `aiops-plugin-prometheus` | 8092 | 指标查询 | 022 |
| `aiops-plugin-logs` | 8093 | 日志查询 | 051 |
| `aiops-module-kb` | 8085 | 知识库 | 040,071 |
| `aiops-worker` | 8086 | 巡检/定时任务 | 060–062 |

## 2. 调用链（阶段 A 核心）

```
Client → gateway → chat → policy(/evaluate 预展示)
                 │      → executor(/run) → policy(/evaluate 执行依据) → plugin/*
                 └──→ platform（环境/策略/审计）
```

审计 → `platform POST /internal/v1/audit`（失败不挡主路）。`execution_records` 仅 executor。

## 3. 存储

| MySQL（控制面） | 外置数据面（插件，G-003） |
|----------------|---------------------------|
| 用户、环境、策略、会话、执行、审计、巡检元数据、KB 元数据 | 日志、指标、告警 body |

`dataQuery` 上限见 PRD §3.8。

## 4. 可观测

`X-Request-Id` 全链路透传；日志字段 PRD §3.11。

## 5. 部署

构建/联调：[BASTION-DEV.md](./BASTION-DEV.md)（**仅联调阶段**）。
