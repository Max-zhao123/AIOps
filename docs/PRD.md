# AIOps PRD（AI 执行）

| 版本 | v1.0 |
| 部署 | K8s `aiops`；Ingress → `aiops-gateway:8080` |
| 需求台账 | **[DEVELOPMENT_PLAN.md](./DEVELOPMENT_PLAN.md)**（30 条 REQ，AI 实现主文档） |

## 0. 规则

| ID | 规则 |
|----|------|
| G-001 | 仅策略 `ALLOW` 可自动执行；其余 `ASK`/`DENY`；禁止 LLM 原文当 Shell；执行依据：`executor` 内 `policy.Evaluate` → `plugin.Execute` |
| G-002 | 新能力 = `pkg/` + `cmd/<模块>` + 独立 Deployment + config；禁止改他模块主流程 |
| G-003 | 海量观测数据不入 MySQL；`dataQuery` 见 §3.5；审计 payload ≤4KB |
| G-004 | 完成 REQ 后更新 [DEVELOPMENT_PLAN.md](./DEVELOPMENT_PLAN.md) |
| G-005 | 一模块一 Pod；**gateway 与 platform 不合并**（入口 vs 业务，见 §3.1） |
| G-006 | **全部产品需求写完再联调**：[DEVELOPMENT_PLAN](./DEVELOPMENT_PLAN.md) 中 **A+B+C+D 全部 30 条 REQ** 代码实现完成且本机 `go build`/`go test` 通过后，才改 Dockerfile/`buildall.sh`/Helm 与跳板机部署；禁止按 REQ 穿插部署（见 [CONVENTIONS.md](./CONVENTIONS.md)） |

## 1. 脚手架（勿重做）

Gin、`pkg/admin`、`utils/jwt.go`、MySQL/GORM（示例 `App`）、`helm/`（待 REQ-019 拆多 Deployment）。无 `/api/v1` 业务。

## 2. 实现顺序（全产品）

**完整顺序与每条 REQ 的实现要点、验收见 [DEVELOPMENT_PLAN.md](./DEVELOPMENT_PLAN.md)。**

```
A: 019→001→002→003→010→011→013→012→020→021→030→031→032→033→014
B: 015→022→050→040→041
C: 060→061→062→080
D: 051→070→071→081→082→083
```

- **开发阶段**：按上表从上到下实现 **全部 30 条**，不得进入联调。
- **联调阶段**：30 条代码完成后，再统一部署与逐条验收。

## 3. 契约

### 3.1 微服务职责（gateway ≠ platform）

| Pod | 职责 | 不做 |
|-----|------|------|
| **aiops-gateway** | Ingress 入口、JWT 校验、生成/透传请求头、反向代理 `/api/v1` `/admin`、对外 `/healthz` `/readyz` | 业务 CRUD、策略、会话、执行、直连 MySQL |
| **aiops-platform** | 环境/用户/RBAC/策略 CRUD/审计入库/Admin 数据与页面 | 对外 Ingress、JWT 签发校验（仅消费 gateway 透传身份） |

**gateway 反向代理**（路由表写在 `cmd/gateway`）：`/api/v1/environments|security-boundaries|audit|plugins` → platform；`/api/v1/chat|sessions|stream` → chat；`/api/v1/actions/pending|confirm` → executor。

### 3.2 身份与请求头（强制）

| 头 | 设置方 | 说明 |
|----|--------|------|
| `Authorization` | gateway 透传 | 原样转发 Bearer JWT |
| `X-Request-Id` | gateway 生成 UUID，全链路透传 | 日志与排障 |
| `X-AIOps-User-Id` | gateway 解析 JWT 后设置 | 下游**禁止**只信 body 里的 userId |
| `X-AIOps-Role` | gateway 解析 JWT 后设置 | `admin\|operator\|readonly\|auditor` |
| `X-AIOps-Environment` | 来自请求 body/query 的 `environment`，gateway 可透传 | 与 PRD 环境 slug 一致 |

集群内 HTTP 调用**必须**带上游传入的上述头（至少 `X-Request-Id`、`X-AIOps-User-Id`、`X-AIOps-Role`）。

### 3.3 调用链与 evaluate 分工

```
Client → gateway（JWT+头）
       → chat → policy/evaluate   # 可选：UI 预展示 decision
              → executor/run → policy/evaluate  # 必须：执行依据
                              → plugin/execute
       → platform（环境/策略/审计 API）
```

| 环节 | 规则 |
|------|------|
| chat 调 evaluate | 仅用于回复文案、展示预决策，**不**写 `execution_records` |
| executor `/internal/v1/run` | **必须**再 evaluate；以此 decision 为准；**唯一**写入/更新 `execution_records` |
| 审计 | `POST platform/internal/v1/audit`；超时 ≤2s；**失败不**导致 chat/execute 返回 5xx，记本服务 warn 日志 |

### 3.4 内部 API

| 服务 | 方法 | 路径 |
|------|------|------|
| policy | POST | `/internal/v1/evaluate` |
| executor | POST | `/internal/v1/run` |
| executor | GET | `/internal/v1/actions/pending`（仅集群内；对外经 gateway 转发） |
| plugin-* | POST | `/internal/v1/execute` |
| platform | POST | `/internal/v1/audit` |

### 3.5 ActionPlan

```json
{
  "plugin": "kubernetes",
  "action": "list",
  "parameters": { "namespace": "default", "resource": "pods" },
  "risk": "read",
  "summary": "string",
  "commandPreview": "kubectl get pods -n default"
}
```

`risk`：`read` | `write` | `notify`。

### 3.6 安全边界 `spec`

```yaml
defaultDecision: ASK
rules:
  - id: string
    match: { plugin: "", actions: [], commandPattern: "", resourceLabels: {} }
    decision: ALLOW | ASK | DENY
    message: string
```

评估：`DENY` 即返回 → rules 顺序首条 → `defaultDecision`。异常 fail-close：`write→DENY`，`read→ASK`。

硬禁止：`rm\s+-rf`、`drop\s+database`、`kubectl\s+delete`、`truncate\s+table`。

### 3.7 Plugin

- `pkg/plugins/<name>/` + `cmd/plugin-<name>/` 独立 Pod
- executor 仅 HTTP 调 `plugins.<name>.endpoint`

### 3.8 config

```yaml
services:
  platform: http://aiops-platform:8081
  chat: http://aiops-chat:8082
  policy: http://aiops-policy:8083
  executor: http://aiops-executor:8084
llm: { provider: openai-compatible, baseUrl: "", apiKey: "", model: "", timeout: 60s }
plugins:
  mock: { enabled: false, endpoint: http://aiops-plugin-mock:8090 }
  kubernetes: { enabled: false, endpoint: http://aiops-plugin-kubernetes:8091, mode: inCluster }
security: { defaultDecision: ASK, confirmTimeout: 15m, hardDenyPatterns: [] }
dataQuery: { maxTimeRange: 24h, maxRows: 500, maxBytes: 1048576, queryTimeout: 30s, promMaxPoints: 10000 }
```

### 3.9 MySQL 表（阶段 A）

`environments`, `users`, `audit_logs`, `security_boundary_policies`, `execution_records`, `chat_sessions`, `chat_messages`

### 3.10 对外 API（经 gateway）

| 方法 | 路径 | 后端 |
|------|------|------|
| * | `/api/v1/environments` | platform |
| * | `/api/v1/security-boundaries`、enable/disable、import-template | platform |
| GET | `/api/v1/plugins` | platform |
| POST | `/api/v1/chat` | chat |
| GET | `/api/v1/chat/stream` | chat |
| GET | `/api/v1/chat/sessions`、`:id/messages` | chat |
| GET | `/api/v1/actions/pending` | **executor** |
| POST | `/api/v1/actions/confirm` | **executor** |
| GET | `/api/v1/audit` | platform |

鉴权：gateway JWT；无 token → 401；`readonly` → `POST .../confirm` 403。

`POST /api/v1/chat` body 必含 `environment`，否则 400。

### 3.11 日志（REQ-019）

各服务 stdout JSON 至少：`service`、`request_id`（来自 `X-Request-Id`）、`environment`；decision 类加 `decision`、`matched_rule_ids`。

## 4. 阶段 A REQ

### REQ-019 微服务骨架（**第一个做**）

- Helm：ARCHITECTURE §1；Ingress → gateway only
- `cmd/gateway|platform|chat|policy|executor|plugin-mock|plugin-kubernetes`
- migrate Job：**镜像 `aiops-platform`**，`command: ["/app/migrate"]`（与 platform 同镜像）
- **验收**：7 业务 Deployment Ready；各 `/healthz` 200；gateway 发出的请求带 `X-Request-Id`

### REQ-001 环境

- **服务**：platform
- **验收**：gateway `GET /api/v1/environments` 200

### REQ-002 RBAC

- **服务**：gateway（JWT 解析+设头）+ platform（用户/角色数据）
- **验收**：401；readonly confirm 403；executor 收到的 internal 请求含 `X-AIOps-User-Id`

### REQ-003 审计

- 各服务 → `POST platform/internal/v1/audit`（短超时）；失败不挡主流程
- **验收**：audit 链完整；platform 宕机时 chat 仍可完成但日志有 audit_failed

### REQ-010 策略 CRUD

- platform 写库；policy **每次 evaluate 读 MySQL**（`enabled` 策略）；enable 后下次 evaluate 生效
- **验收**：evaluate 与 spec 一致

### REQ-011 / REQ-013

- policy；evaluate + risk
- **验收**：同前

### REQ-012 执行网关

- executor；run 内**必须** evaluate；**唯一**写 `execution_records`
- **验收**：无 in-process 插件 import；跳过第二次 evaluate 的测试失败

### REQ-020 / REQ-021

- 插件独立 Pod；executor HTTP 调用
- **验收**：同前

### REQ-030–033

- chat + gateway；pending/confirm 路由到 executor
- **验收**：同前；confirm 后 executor 写 audit

### REQ-014 策略 Admin

- **服务**：platform（数据+API）；静态页可由 gateway 反代 platform 或 gateway 嵌模板，**业务逻辑仍在 platform**
- **验收**：YAML 与 API 一致

## 5. 阶段 B/C/D REQ 索引

条文与验收在 [DEVELOPMENT_PLAN.md](./DEVELOPMENT_PLAN.md) 各阶段表；此处不重复。

| 阶段 | 主题 | REQ |
|------|------|-----|
| B | 策略模拟、Prometheus、数据查询 API、知识库、RAG | 015,022,050,040,041 |
| C | Worker、应用巡检、告警联动、故障自愈 | 060,061,062,080 |
| D | 日志插件、IM、HelpDesk、高危预警、修复辅助、RCA | 051,070,071,081,082,083 |

## 6. 产品 DoD（联调后）

- [ ] DEVELOPMENT_PLAN **30/30** `done`
- [ ] 全产品 `go build` 通过且集群验收通过
- [ ] 见 DEVELOPMENT_PLAN「产品 DoD」章节
