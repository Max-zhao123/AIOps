# AIOps 开发计划（AI 执行主文档）

> **AI 实现产品时的第一必读文件。** 逐条实现下表全部 REQ；契约见 [PRD.md](./PRD.md) §3；Pod/端口见 [ARCHITECTURE.md](./ARCHITECTURE.md)。

## 开发节奏（G-006，强制）

| 阶段 | 做什么 | 禁止 |
|------|--------|------|
| **① 开发** | 下表 **全部 30 条 REQ** Go 代码 + `go build ./cmd/...` | Dockerfile、`buildall.sh`、`helm/`、跳板机 |
| **② 联调** | ① 完成后统一部署，**逐条 REQ** 验收 | 未写完代码就上集群 |

| 当前阶段 | **联调中** |
| 产品 REQ | **30** |
| 代码完成 | **30 / 30** |
| 联调 `done` | **0 / 30** |
| 最后更新 | 2026-05-26 |

**本机构建**：`go build ./cmd/...`、`go test ./...` 已通过（含 boundary 单测/bench、executor 源码约束测试）。

---

## 全局实现顺序

```
A: 019→001→002→003→010→011→013→012→020→021→030→031→032→033→014
B: 015→022→050→040→041
C: 060→061→062→080
D: 051→070→071→081→082→083
```

---

## 阶段 A

| REQ | 名称 | 服务 | 代码 | 状态 | 实现位置（摘要） |
|-----|------|------|------|------|------------------|
| 019 | 微服务骨架 | 全部 | 完成 | `pending` | `cmd/*`、`pkg/runtime` |
| 001 | 环境 | platform | 完成 | `pending` | `GET/POST /api/v1/environments`、seed |
| 002 | RBAC | gateway+platform | 完成 | `pending` | `pkg/gateway/auth`、login |
| 003 | 审计 | platform | 完成 | `pending` | `pkg/audit`、`/internal/v1/audit` |
| 010 | 策略 CRUD | platform | 完成 | `pending` | `pkg/platform/handlers` |
| 011 | 规则引擎 | policy | 完成 | `pending` | `pkg/boundary`+单测 |
| 013 | 高危 | policy | 完成 | `pending` | `pkg/boundary/risk` |
| 012 | 执行网关 | executor | 完成 | `pending` | `pkg/executor`；confirm 二次 evaluate |
| 020 | mock 插件 | plugin-mock | 完成 | `pending` | `pkg/plugins/mock` |
| 021 | K8s 插件 | plugin-kubernetes | 完成 | `pending` | client-go；本机可用 `KUBECONFIG`，集群内用 inCluster |
| 030 | LLM | chat | 完成 | `pending` | `pkg/llm` |
| 031 | 会话/Plan | chat | 完成 | `pending` | `ParseActionPlans` |
| 032 | Chat/SSE | chat | 完成 | `pending` | stream 分块 SSE |
| 033 | ASK | executor | 完成 | `pending` | pending/confirm/409 |
| 014 | Admin | platform | 完成 | `pending` | `initialize.Admin` |

---

## 阶段 B

| REQ | 名称 | 服务 | 代码 | 状态 | 实现位置（摘要） |
|-----|------|------|------|------|------------------|
| 015 | 策略模拟 | platform | 完成 | `pending` | `POST .../simulate` |
| 022 | Prometheus | plugin-prometheus | 完成 | `pending` | `cmd/plugin-prometheus` |
| 050 | 数据查询 | platform | 完成 | `pending` | `POST /api/v1/data/query` |
| 040 | 知识库 | module-kb | 完成 | `pending` | `cmd/module-kb`、`pkg/module/kb` |
| 041 | RAG | chat | 完成 | `pending` | `useRag` + kb search |

---

## 阶段 C

| REQ | 名称 | 服务 | 代码 | 状态 | 实现位置（摘要） |
|-----|------|------|------|------|------------------|
| 060 | Worker | worker | 完成 | `pending` | `cmd/worker`：先 `/process` 再 `/complete` |
| 061 | 巡检 | platform | 完成 | `pending` | `inspection.go` 经 executor 逐步执行 |
| 062 | 告警联动 | platform | 完成 | `pending` | 失败写 `notify_records` |
| 080 | 自愈 | platform+executor | 完成 | `pending` | runbooks execute |

---

## 阶段 D

| REQ | 名称 | 服务 | 代码 | 状态 | 实现位置（摘要） |
|-----|------|------|------|------|------------------|
| 051 | 日志插件 | plugin-logs | 完成 | `pending` | ES `_search`；未配置返回错误 |
| 070 | IM | platform | 完成 | `pending` | `/api/v1/im/webhook` |
| 071 | HelpDesk | platform+kb | 完成 | `pending` | `helpdesk.go`：KB+RAG+LLM |
| 081 | 高危预警 | platform | 完成 | `pending` | risk-alerts scan/list |
| 082 | 修复辅助 | chat | 完成 | `pending` | `/api/v1/chat/assist` |
| 083 | RCA | chat | 完成 | `pending` | `/api/v1/chat/rca` |

---

## 联调阶段 TODO

- [x] 扩展 `buildall.sh` 至 11 个业务镜像
- [x] Helm `microservices` 启用 B+ Pod（prometheus/logs/kb/worker）
- [x] DB 服务挂载集群 ConfigMap（gateway/chat/policy/executor/kb）
- [ ] 跳板机 build + `helm upgrade` + 逐 REQ 验收标 `done`

## 产品 DoD

- [ ] 30 条联调 `done`
- [ ] MVP + HelpDesk/巡检/IM 演示

## 变更记录

| 日期 | 变更 |
|------|------|
| 2026-05-22 | 30 REQ 代码实现完成；进入待联调 |
| 2026-05-26 | 加固：去掉 K8s FAKE；巡检真执行；confirm 二次 evaluate；worker process；logs 未配置报错；dataquery 截断 |
| 2026-05-26 | 联调准备：buildall 11 镜像；Helm B+ Pod；服务间 env；dataquery maxRows |
