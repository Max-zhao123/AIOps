# AIOps PRD 优化增补文档

| 版本 | v2.0-optimization |
| 基线 | PRD.md v1.0 / DEVELOPMENT_PLAN.md（30 条 REQ done） |
| 目的 | 消除验收歧义、补全边界条件、补充运维缺失功能、新增 AI 编码防跑偏规则 |

---

## 一、AI 编码防跑偏规则（G-007 ~ G-015）

> 为 PRD §0 规则表新增，防止 AI 编码时引入不必要的复杂性或遗漏关键约束。

### G-007 验收标准必须可测试

| | 要求 |
|-|------|
| **必须做** | 每条 REQ 的验收条件必须是具体的、可自动或人工验证的断言。格式：`输入X → 返回Y / 状态Z / 数据库存在W`。禁止使用"同前"、"功能正常"、"界面正确"等模糊表述。 |
| **禁止做** | 禁止验收条件引用其他 REQ 而不给出具体断言。禁止验收条件含"等"、"如"、"等类似"等模糊词。 |
| **正例** | `POST /api/v1/environments body={name:"test",slug:"test"} → 201; GET /api/v1/environments → 含 name="test" 的条目` |
| **反例** | `验收：环境管理功能正常` |

### G-008 统一错误响应格式

| | 要求 |
|-|------|
| **必须做** | 所有服务对外 API 错误统一返回 `{"code": int, "message": string, "request_id": string}`。`code` 使用 HTTP 状态码；`message` 为人类可读中文描述；`request_id` 来自 `X-Request-Id` 头。 |
| **禁止做** | 禁止 `gin.H{"error": ...}`、`gin.H{"err": ...}`、`gin.H{"msg": ...}` 等不一致格式。禁止 5xx 响应暴露内部堆栈或 SQL。 |
| **正例** | `404 {"code":404,"message":"策略不存在","request_id":"abc-123"}` |
| **反例** | `404 {"error":"record not found"}` 或 `500 {"msg":"pq: connection refused"}` |

### G-009 数据模型校验

| | 要求 |
|-|------|
| **必须做** | 所有 API 入参必须 struct tag 校验（`binding:"required,max=100"` 等）。数据库字段必须定义：类型、长度/精度、NOT NULL / DEFAULT、唯一索引。敏感字段（密码、密钥）禁止明文存储，必须 bcrypt 或 AES 加密。 |
| **禁止做** | 禁止 `string` 字段无长度限制存入 DB。禁止 `interface{}` / `map[string]interface{}` 作为持久化模型字段（JSON 列除外）。 |
| **正例** | `Name string \`gorm:"type:varchar(100);not null;uniqueIndex:idx_env_slug" json:"name" binding:"required,max=100"\`` |
| **反例** | `Name string \`json:"name"\`` 或 `Data map[string]interface{}` |

### G-010 API 命名与幂等性

| | 要求 |
|-|------|
| **必须做** | REST 路径使用小写 + 连字符、复数名词；内部路径 `/internal/v1/` 前缀。写操作（POST/PUT/DELETE）必须幂等或返回 409 冲突。确认类操作使用 `POST .../confirm`，拒绝使用 `POST .../reject`。 |
| **禁止做** | 禁止动词路径（`/createEnv`、`/deletePolicy`）。禁止同一操作多次调用产生重复副作用（无幂等保护）。 |
| **正例** | `POST /api/v1/environments`（幂等：slug 唯一索引，重复返回 409） |
| **反例** | `POST /api/v1/createEnvironment` 或 `GET /api/v1/deletePolicy?id=1` |

### G-011 超时与重试

| | 要求 |
|-|------|
| **必须做** | 所有跨服务 HTTP 调用必须设置超时（默认 30s，可按插件配置覆盖）。重试必须限定次数（≤3）+ 指数退避（1s/2s/4s）。重试仅对网络超时/5xx 生效，4xx 不重试。 |
| **禁止做** | 禁止 `http.Client{Timeout: 0}`。禁止无退避的立即重试。禁止对 4xx 响应重试。 |
| **正例** | `http.Client{Timeout: 30*time.Second}; retry: 3次, backoff: 1s,2s,4s, only on 5xx/timeout` |
| **反例** | `http.DefaultClient` 或 `for i:=0;i<10;i++ { resp, err := http.Get(url) }` |

### G-012 结构化日志

| | 要求 |
|-|------|
| **必须做** | 所有日志 JSON stdout，必须字段：`ts`、`level`、`service`、`request_id`。业务日志加 `environment`、`user_id`。error 级日志必须含 `error` 字段和调用栈（`caller`）。 |
| **禁止做** | 禁止 `fmt.Println`、`log.Printf`。禁止日志含敏感信息（密码、token、密钥原文）。禁止在 hot path 打 info 日志（如每条 SSE chunk）。 |
| **正例** | `{"ts":"2026-06-01T10:00:00Z","level":"error","service":"executor","request_id":"abc","error":"plugin timeout","caller":"executor/run.go:42"}` |
| **反例** | `log.Printf("error: %v", err)` 或 `fmt.Println("something happened")` |

### G-013 并发安全

| | 要求 |
|-|------|
| **必须做** | 共享状态（内存缓存、pending 队列）必须用 `sync.Mutex` / `sync.RWMutex` / channel 保护。数据库写操作必须用事务（`db.Transaction`）。乐观锁用 `updated_at` 版本号，冲突返回 409。 |
| **禁止做** | 禁止全局 `map` 无锁并发读写。禁止 `SELECT` 后延迟 `UPDATE` 不加 `FOR UPDATE` 或版本校验。 |
| **正例** | `db.Model(&r).Where("id=? AND updated_at=?", id, oldTS).Updates(vals); RowsAffected==0 → 409` |
| **反例** | `globalPendingMap[actionID] = action` （无锁） |

### G-014 测试约束

| | 要求 |
|-|------|
| **必须做** | 策略引擎、执行网关、认证鉴权必须有单元测试。测试用例覆盖：正常路径、边界值（空/超长/特殊字符）、权限拒绝、并发冲突。测试不依赖外部服务（mock HTTP/DB）。 |
| **禁止做** | 禁止仅写 happy path 测试。禁止测试用外部服务真实连接（真实 K8s/ES/Prometheus）。禁止 `t.Skip("TODO")`。 |
| **正例** | `TestEvaluate_HardDenyPattern: input "rm -rf /" → decision=DENY` |
| **反例** | `TestMain: t.Skip("need real k8s")` |

### G-015 禁止做清单

| | 要求 |
|-|------|
| **必须做** | 以下行为显式禁止：① 在业务代码中 `time.Sleep` 等待异步结果；② 用文件做 IPC（必须 HTTP/gRPC）；③ 在 gateway 以外的服务直连客户端；④ 在 executor 绕过 policy 直接调 plugin；⑤ 硬编码 IP/端口（必须从 config 读取）；⑥ 在日志/响应中泄露密钥或 token。 |
| **禁止做** | 违反以上任何一项即视为实现错误。 |
| **正例** | `endpoint := cfg.Services.Policy` |
| **反例** | `resp, _ := http.Post("http://10.0.0.5:8083/internal/v1/evaluate", ...)` |

---

## 二、现有 REQ 验收标准补全

> 对 DEVELOPMENT_PLAN.md 中验收模糊的 REQ，逐条给出可测试的明确断言。

| REQ | 原验收 | 补全后验收 |
|-----|--------|------------|
| 019 | 7 业务 Deployment Ready；各 `/healthz` 200；gateway 发出的请求带 `X-Request-Id` | ① `kubectl get deploy -n aiops` 显示 7 个 Deployment AVAILABLE=1；② 对每个 Pod `curl http://<pod>:<port>/healthz` 返回 200；③ `curl gateway:8080/api/v1/environments -H "Authorization: Bearer <jwt>"` 响应头含 `X-Request-Id` 且值匹配日志中 request_id；④ migrate Job 状态 Completed |
| 001 | gateway `GET /api/v1/environments` 200 | ① `POST /api/v1/environments body={name:"开发",slug:"dev"}` → 201，返回体含 id；② `GET /api/v1/environments` → 200，数组含 slug="dev" 条目；③ 重复 POST 相同 slug → 409；④ POST name 为空 → 400；⑤ 无 Authorization → 401 |
| 002 | 401；readonly confirm 403；executor 收到的 internal 请求含 `X-AIOps-User-Id` | ① `POST /api/v1/auth/login body={username:"admin",password:"xxx"}` → 200 返回 JWT；② 无 token GET /api/v1/environments → 401；③ token 角色 readonly `POST /api/v1/actions/confirm` → 403；④ executor 日志含 `X-AIOps-User-Id` 头且值与 JWT 中 sub 一致；⑤ 过期 token → 401 |
| 003 | audit 链完整；platform 宕机时 chat 仍可完成但日志有 audit_failed | ① 执行一次巡检后 `GET /api/v1/audit` 返回对应记录且 action_type/username/environment 非空；② 临时将 platform Service port 改为 0 模拟宕机 → chat 仍返回 200；③ chat 日志含 `audit_failed` warn；④ 审计写入超时 ≤2s（日志时间差可验证） |
| 010 | evaluate 与 spec 一致 | ① `POST /api/v1/security-boundaries` 创建策略 spec 含 `defaultDecision:ASK, rules:[{match:{plugin:"kubernetes"},decision:ALLOW}]` → 201；② `POST .../enable` → 200；③ `POST policy/internal/v1/evaluate body={plugin:"kubernetes",action:"list",risk:"read"}` → decision=ALLOW；④ `POST evaluate body={plugin:"unknown"}` → decision=ASK（defaultDecision） |
| 011 | 同前 | ① `POST evaluate` 传入匹配 rule 的 ActionPlan → decision 为 rule 指定值；② 传入不匹配任何 rule → decision 为 defaultDecision；③ rules 按顺序首条匹配（rule1=ALLOW, rule2=DENY, 传入同时匹配的 → ALLOW）；④ write risk 且策略异常 → DENY（fail-close） |
| 013 | 同前 | ① `POST evaluate` 传入 `commandPreview:"rm -rf /"` → decision=DENY，matched_rule_ids 含 hardDeny 标识；② 传入 `kubectl delete pod --all` → DENY；③ 传入正常只读命令 → 不触发高危规则 |
| 012 | 无 in-process 插件 import；跳过第二次 evaluate 的测试失败 | ① executor 代码中无 `import "aiops/pkg/plugins/*"` 的直接引用；② `POST executor/internal/v1/run` 日志显示两次 policy evaluate 调用（执行前 + confirm 后）；③ run 不经过 evaluate 直接执行 → 测试失败；④ `execution_records` 仅由 executor 写入（其他服务写入 → 测试失败） |
| 020 | 同前 | ① `POST plugin-mock:8090/internal/v1/execute body={action:"echo",parameters:{message:"hi"}}` → 200 返回 `{"result":"hi"}`；② executor 通过 HTTP 调用 mock 插件（非 in-process） |
| 021 | 同前 | ① `POST plugin-kubernetes:8091/internal/v1/execute body={action:"list",parameters:{namespace:"default",resource:"pods"}}` → 200 返回 pod 列表；② 集群内用 inCluster 模式，本机用 KUBECONFIG；③ 无权限时返回明确错误而非 panic |
| 030 | 同前 | ① `POST /api/v1/chat body={message:"查看default命名空间pod",environment:"dev"}` → 200 返回含 ActionPlan 的回复；② LLM 超时 → 返回 `{"code":504,"message":"LLM 响应超时"}` |
| 031 | 同前 | ① chat 返回的 LLM 文本中解析出 `▶PLAN_START◀...▶PLAN_END◀` 块 → `ParseActionPlans` 返回结构化 ActionPlan 数组；② 无 Plan 标记 → 空数组；③ Plan 格式错误 → 忽略不 panic |
| 032 | 同前 | ① `GET /api/v1/chat/stream?session_id=x&message=hi&environment=dev` → SSE `data: {token}\n\n` 格式；② 每条 SSE 事件间隔 ≤100ms；③ 流结束发 `data: [DONE]\n\n`；④ 无 token → 401 |
| 033 | 同前；confirm 后 executor 写 audit | ① write risk 操作 → 状态 pending；② `GET /api/v1/actions/pending` 含该条目；③ `POST /api/v1/actions/confirm body={action_id:x,approved:true}` → 200；④ 再次 confirm 同 id → 409；⑤ confirm 后 `GET /api/v1/audit` 含执行审计记录 |
| 014 | YAML 与 API 一致 | ① Admin 页面展示的策略 YAML 与 `GET /api/v1/security-boundaries/:id` 返回的 spec JSON 一致；② Admin 修改 YAML 保存 → `GET` 返回更新后的值；③ 静态页由 gateway 或 platform 反代，业务逻辑在 platform |
| 015 | （开发计划未详细展开） | ① `POST /api/v1/security-boundaries/:id/simulate body={actionplan}` → 200 返回 `{decision,matched_rule_ids,message}`；② 模拟不改变策略状态；③ 模拟不写审计 |
| 022 | 同前 | ① `POST plugin-prometheus:8092/internal/v1/execute body={action:"query",parameters:{query:"up",start:"2026-01-01T00:00:00Z",end:"2026-01-01T01:00:00Z"}}` → 200 返回时序数据；② 超过 maxPoints=10000 → 截断并返回警告 |
| 050 | 同前 | ① `POST /api/v1/data/query body={plugin:"prometheus",environment:"dev",query:"up",timeRange:{start:...,end:...}}` → 200；② 时间范围超 24h → 400；③ 结果超 maxRows=500 → 截断 + `truncated:true` 标记 |
| 040 | 同前 | ① `POST /api/v1/kb/documents body={title:"K8s运维手册",content:"...",category:"k8s"}` → 201；② `GET /api/v1/kb/documents` → 含该文档；③ `DELETE /api/v1/kb/documents/:id` → 200 |
| 041 | 同前 | ① chat 开启 RAG 后，提问 → LLM 回答引用 KB 文档内容；② KB 无相关文档 → 回答无引用但正常返回；③ RAG search 超时 → 降级为无 RAG 回答 |
| 060 | 同前 | ① worker 从队列取任务 → 调用 executor `/process` → 成功后调 `/complete`；② 任务执行失败 → worker 记录状态为 failed，不无限重试 |
| 061 | 同前 | ① `POST /api/v1/inspections body={name:"pod巡检",environment:"dev",steps:[{plugin:"kubernetes",action:"list",parameters:{namespace:"default",resource:"pods"}}]}` → 201；② `POST /api/v1/inspections/:id/run` → 200，巡检逐步执行；③ `GET /api/v1/inspections/:id/reports` → 含步骤执行结果 |
| 062 | 同前 | ① 告警触发后 `notify_records` 表有记录；② 通知失败 → 状态为 failed，有 error_message |
| 080 | 同前 | ① `POST /api/v1/runbooks body={name:"pod重启",steps:[...]}` → 201；② `POST /api/v1/runbooks/:id/execute` → 经 executor 执行，策略评估；③ 执行结果记录 |
| 051 | 同前 | ① `POST plugin-logs:8093/internal/v1/execute body={action:"search",parameters:{query:"error",index:"app-*"}}` → ES 搜索结果；② 未配置 ES → 返回 `{"code":503,"message":"日志服务未配置"}` |
| 070 | 同前 | ① `POST /api/v1/im/webhook body={...}` → 200；② webhook URL 配置页面可查看；③ 测试发送消息成功 |
| 071 | 同前 | ① `POST /api/v1/helpdesk/chat body={message:"如何查看pod"} → 200 返回 KB+RAG 回答；② 无相关 KB → 返回通用回答 |
| 081 | 同前 | ① `POST /api/v1/risk-alerts/scan` → 触发扫描；② `GET /api/v1/risk-alerts` → 含预警列表含 severity 字段（critical/high/medium/low） |
| 082 | 同前 | ① `POST /api/v1/chat/assist body={message:"pod crashloopbackoff",environment:"dev"}` → 200 返回修复建议；② LLM 生成含可操作步骤 |
| 083 | 同前 | ① `POST /api/v1/chat/rca body={event:"高CPU告警",environment:"dev"}` → 200 返回根因分析含因果链；② 分析结果含时间线 |

---

## 三、新增 REQ（运维场景）

### REQ-090 定时调度

- **服务**：worker + platform
- **优先级**：P0
- **依赖**：REQ-060（Worker）、REQ-061（巡检）
- **实现要点**：
  - platform 新增 `schedules` 表，CRUD API：`GET/POST/PUT/DELETE /api/v1/schedules`
  - schedule 字段含 cron 表达式 + 时区 + 目标类型（inspection/runbook）+ 目标 ID
  - worker 内置 cron 调度器（`github.com/robfig/cron/v3`），启动时从 DB 加载 enabled schedules
  - platform 修改 schedule 时通过内部 API `POST worker/internal/v1/schedules/reload` 通知 worker 热加载
  - 执行历史写入 `schedule_executions` 表
- **验收**：
  ① `POST /api/v1/schedules body={name:"每日pod巡检",cron:"0 9 * * *",timezone:"Asia/Shanghai",target_type:"inspection",target_id:1,enabled:true}` → 201；
  ② worker 日志显示 cron 任务注册成功；
  ③ 到达 cron 时间后自动触发巡检执行；
  ④ `GET /api/v1/schedules/:id/executions` → 含执行历史
- **边界条件**：
  - cron 表达式校验：非法表达式 → 400
  - 时区默认 UTC，支持 IANA 时区名
  - 同一 target 不允许重叠执行（上次未完成 → 跳过，记录 skipped）
  - schedule 删除后 worker 自动取消
  - 执行超时：继承目标 inspection/runbook 的超时设置

### REQ-091 告警通知通道

- **服务**：platform
- **优先级**：P0
- **依赖**：REQ-062（告警联动）
- **实现要点**：
  - platform 新增 `notification_channels` 表（type: email/webhook/sms, config JSON）
  - platform 新增 `notification_policies` 表（severity → channel_id 映射）
  - platform 新增 `notification_templates` 表（channel_type + 事件类型 → 模板文本）
  - 实际推送逻辑：
    - email: SMTP 发送（`net/smtp`，配置 host/port/sender/password）
    - webhook: HTTP POST JSON payload 到企业微信/钉钉 Webhook URL
    - sms: 预留接口，初期仅记录（`notify_records.type=sms`，status=pending）
  - 通知静默：同 alert_key 在静默窗口（默认 30min）内不重复发送
  - 通知抑制：高级别告警抑制低级别（critical 抑制 warning）
  - 失败重试：3 次，间隔 1min/2min/4min
- **验收**：
  ① `POST /api/v1/notification-channels body={type:"webhook",name:"企业微信",config:{url:"https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx"}}` → 201；
  ② `POST /api/v1/notification-policies body={severity:"critical",channel_ids:[1]}` → 201；
  ③ 触发 critical 告警 → webhook URL 收到 HTTP POST 且 payload 符合企业微信格式；
  ④ 30min 内同 alert_key 第二次告警 → 不重复发送；
  ⑤ 发送失败 → 重试 3 次 + `notify_records.status=failed`
- **边界条件**：
  - webhook 超时 10s
  - email 发送失败不阻断主流程（异步+重试）
  - 通知模板变量缺失 → 使用默认值，不 panic
  - 通知频率限制：单个 channel 每分钟最多 60 条

### REQ-092 执行超时与重试

- **服务**：executor + plugin-*
- **优先级**：P0
- **依赖**：REQ-012（执行网关）
- **实现要点**：
  - executor config 新增 `pluginTimeouts` map：`{mock:30s, kubernetes:120s, prometheus:60s, logs:60s}`
  - executor 调用 plugin 时使用 `context.WithTimeout` 设置 per-plugin 超时
  - executor config 新增 `retryPolicy`：`{maxRetries:3, backoffBase:1s, backoffMax:30s, retryableErrors:["timeout","connection_refused","5xx"]}`
  - 超时后 `execution_records` 状态标记为 `timeout`，error_message 含超时详情
  - 巡检步骤失败重试：`inspections` 表步骤新增 `retry_count` 和 `max_retries` 字段
  - 执行进度追踪：executor 每完成一个 plugin 调用更新 `execution_records.progress`
- **验收**：
  ① 配置 `pluginTimeouts.kubernetes=5s`，执行超时 K8s 操作 → 5s 后返回 timeout 错误；
  ② 超时后 `execution_records` 状态为 `timeout`，error_message 含 "context deadline exceeded"；
  ③ plugin 临时不可用 → 重试 3 次，日志含重试记录；
  ④ 4xx 错误不重试（仅一次）；
  ⑤ 巡检步骤失败且有 max_retries → 自动重试
- **边界条件**：
  - 总执行时间上限 600s（超出整体超时，强制终止）
  - 重试退避：1s → 2s → 4s，最大 30s
  - 幂等保护：重试时检查 `execution_records` 是否已 completed，已完则跳过

### REQ-093 凭证管理

- **服务**：platform
- **优先级**：P1
- **依赖**：无（可独立于其他 E 阶段 REQ 先行）
- **实现要点**：
  - platform 新增 `credentials` 表（name, type: smtp/api_key/database/kubeconfig, encrypted_value TEXT, expires_at, last_rotated_at）
  - 加密方式：AES-256-GCM，密钥来源为环境变量 `AIOPS_ENCRYPTION_KEY`（部署时注入，K8s Secret）
  - 解密仅在运行时内存，日志/响应禁止输出原文
  - 凭证轮转提醒：`expires_at` 前 7 天发通知（复用 REQ-091 通道）
  - CRUD API：`GET/POST/PUT/DELETE /api/v1/credentials`（encrypted_value 写入时加密，读取时脱敏返回 `****`）
  - 运行时获取：`GET /api/v1/credentials/:id/value`（内部 API，仅集群内可调用）
  - config.yaml 中的 `llm.apiKey` 等迁移为引用凭证 ID
- **验收**：
  ① `POST /api/v1/credentials body={name:"LLM-API-Key",type:"api_key",value:"sk-xxx"}` → 201，数据库 encrypted_value 非原文；
  ② `GET /api/v1/credentials` → 返回 `value:"****"`，不泄露原文；
  ③ `GET /api/v1/credentials/:id/value`（内部）→ 返回解密后原文；
  ④ 凭证过期前 7 天 → 触发通知；
  ⑤ 无 `AIOPS_ENCRYPTION_KEY` 环境变量 → 服务启动失败
- **边界条件**：
  - value 最大 4096 字节
  - 同名凭证 → 409
  - 凭证被引用时禁止删除（返回 409 + 引用列表）
  - 加密密钥轮转：支持 re-encrypt（管理 API `POST /api/v1/credentials/re-encrypt`）

### REQ-094 变更审批流

- **服务**：executor + platform
- **优先级**：P0
- **依赖**：REQ-033（ASK）、REQ-012（执行网关）
- **实现要点**：
  - 策略 spec 新增 `approvalConfig`：
    ```yaml
    approvalConfig:
      write: { minApprovers: 2, timeout: 30m, timeoutAction: REJECT }
      notify: { minApprovers: 1, timeout: 15m, timeoutAction: REJECT }
    ```
  - `execution_records` 新增 `approvers` JSON 字段：`[{user_id, role, approved_at}]`
  - confirm 逻辑改造：
    - `minApprovers=1`：现有逻辑（单人确认）
    - `minApprovers=2`：双人确认（两人 confirm 后才执行），且确认人角色不能相同
    - 审批超时：定时器到期 → 自动标记 REJECTED，不执行
  - platform 新增 `approval_policies` 表（risk → minApprovers/timeout 映射，覆盖策略级默认）
  - 审计记录含审批链：谁发起、谁审批、审批时间
- **验收**：
  ① write risk + minApprovers=2 → 第一人 confirm 后状态为 `partially_approved`，不执行；
  ② 第二人 confirm → 状态变 `approved`，执行开始；
  ③ 同一人 confirm 两次 → 仍为 `partially_approved`（去重）；
  ④ 审批超时 30m → 自动 REJECTED，`execution_records.status=rejected_timeout`；
  ⑤ 审计记录含完整审批链
- **边界条件**：
  - 审批人不能是操作发起人
  - 超时粒度：分钟，最小 5min
  - 并发 confirm：用乐观锁防竞态（`WHERE id=? AND status='partially_approved'`）
  - 策略未配 approvalConfig → 默认 minApprovers=1

### REQ-095 操作回滚

- **服务**：executor + plugin-*
- **优先级**：P1
- **依赖**：REQ-012（执行网关）
- **实现要点**：
  - 执行前快照：write risk 操作执行前，plugin 返回 `snapshot` 字段（当前状态 JSON）
  - `execution_records` 新增 `pre_snapshot` JSON 字段
  - 回滚 ActionPlan：策略 spec 新增 `rollbackPlans`：
    ```yaml
    rollbackPlans:
      - match: { plugin: "kubernetes", action: "scale" }
        rollback: { plugin: "kubernetes", action: "scale", parameters_from_snapshot: ["replicas"] }
    ```
  - 回滚触发：
    - 手动：`POST /api/v1/actions/:id/rollback` → executor 根据 rollbackPlan + snapshot 生成回滚 ActionPlan
    - 自动：策略 `rollbackPolicy: AUTO_ON_FAILURE`，执行失败自动触发
  - 回滚操作同样经过策略评估
  - 回滚记录写入 `execution_records`（`parent_execution_id` 关联原执行）
- **验收**：
  ① 执行 K8s scale（replicas: 2→5）→ `execution_records.pre_snapshot` 含 `{"replicas":2}`；
  ② `POST /api/v1/actions/:id/rollback` → executor 生成 scale replicas=2 的 ActionPlan → 经策略评估后执行；
  ③ 无 rollbackPlan 的操作 → rollback 返回 400 `"无可用回滚计划"`；
  ④ 回滚执行记录的 `parent_execution_id` 指向原执行
- **边界条件**：
  - 快照大小限制 1MB
  - 回滚操作也需审批（write risk）
  - 回滚失败 → 记录状态 `rollback_failed`，不自动二次回滚
  - 只读操作（read risk）不需要快照

### REQ-096 SLA/MTTR 追踪

- **服务**：platform
- **优先级**：P1
- **依赖**：REQ-062（告警联动）、REQ-091（通知通道）
- **实现要点**：
  - platform 新增 `sla_definitions` 表（name, severity, response_time_min, resolution_time_min）
  - platform 新增 `sla_records` 表（alert_id, acknowledged_at, resolved_at, response_sla_met, resolution_sla_met）
  - 告警联动触发时创建 SLA 记录，response_time 计时开始
  - 首次确认操作 → acknowledged_at 填充，判断 response SLA
  - execution_records 完成 → resolved_at 填充，判断 resolution SLA
  - SLA 未达标 → 通知升级（通知上级负责人，复用 REQ-091）
  - Dashboard API：`GET /api/v1/sla/stats?period=30d` → 返回 SLA 达标率
  - MTTR 计算：`AVG(resolved_at - created_at)` 按环境/严重级别分组
- **验收**：
  ① `POST /api/v1/sla-definitions body={name:"P1响应SLA",severity:"critical",response_time_min:15,resolution_time_min:60}` → 201；
  ② critical 告警触发 → 15min 内未确认 → `sla_records.response_sla_met=false` + 升级通知；
  ③ `GET /api/v1/sla/stats?period=30d` → 返回各 severity 的 MTTR 和 SLA 达标率；
  ④ 告警 30min 内解决 → `resolution_sla_met=true`
- **边界条件**：
  - SLA 时间为工作日/日历日（配置项，默认日历日）
  - 未 resolved 的告警不计入 MTTR 统计
  - SLA 升级不循环（最多升一级）

### REQ-097 多环境配置隔离

- **服务**：platform + gateway
- **优先级**：P1
- **依赖**：REQ-001（环境）
- **实现要点**：
  - platform 新增 `environment_configs` 表（environment_slug, config_key, config_value JSON, override_type: plugin/service/llm）
  - config.yaml 作为默认值，`environment_configs` 覆盖
  - 获取配置优先级：environment_configs > config.yaml > 硬编码默认值
  - platform 新增内部 API：`GET platform/internal/v1/config/:environment/:key` → 返回合并后配置
  - 各服务启动时加载默认配置，运行时按 environment 从 platform 获取覆盖
  - 资源配额：`environment_configs` 中可配置 `maxExecutionsPerDay`、`maxConcurrentExecutions`
- **验收**：
  ① `POST /api/v1/environments/dev/config body={key:"plugins.kubernetes",value:{endpoint:"http://k8s-dev:8091"}}` → 200；
  ② `GET platform/internal/v1/config/dev/plugins.kubernetes` → 返回 dev 环境覆盖值；
  ③ `GET platform/internal/v1/config/prod/plugins.kubernetes` → 返回 config.yaml 默认值；
  ④ 环境配额超限 → 执行请求返回 429
- **边界条件**：
  - config_value 最大 10KB
  - 不存在的 environment → 404
  - 不存在的 config_key → 返回默认值（非 404）
  - 配置热加载：变更后 30s 内生效（polling 或 watch DB）

### REQ-098 审计导出

- **服务**：platform
- **优先级**：P2
- **依赖**：REQ-003（审计）
- **实现要点**：
  - platform 新增 API：`GET /api/v1/audit/export?format=csv&environment=dev&start=...&end=...` → 流式返回 CSV
  - platform 新增 API：`GET /api/v1/audit/export?format=pdf&environment=dev&start=...&end=...` → 返回 PDF
  - CSV：首行为列名，每行一条审计记录，UTF-8 BOM
  - PDF：含标题、查询条件摘要、表格（使用 `github.com/jung-kurt/gofpdf` 或同类库）
  - 合规报告模板：`GET /api/v1/audit/report?template=compliance&period=monthly` → 按月生成合规摘要
  - 大数据量：CSV 流式输出（不全部加载到内存），PDF 限制最多 10000 条
- **验收**：
  ① `GET /api/v1/audit/export?format=csv&environment=dev` → 200，Content-Type: text/csv，UTF-8 BOM 开头；
  ② CSV 内容含审计表所有列，且与 `GET /api/v1/audit` 同数据；
  ③ `GET /api/v1/audit/export?format=pdf` → 200，Content-Type: application/pdf，可打开；
  ④ 无审计数据 → CSV 返回仅表头，PDF 返回空报告页；
  ⑤ 非 admin/auditor 角色 → 403
- **边界条件**：
  - 导出时间范围必须指定（start+end），最大 90 天
  - CSV 行数无上限（流式），PDF 上限 10000 条
  - 并发导出限制：同用户同时最多 1 个导出任务

### REQ-099 配置热更新

- **服务**：全部（主要 platform + chat）
- **优先级**：P1
- **依赖**：REQ-097（多环境配置）
- **实现要点**：
  - platform 维护 `config_versions` 表（version INT 自增, updated_at）
  - 各服务启动时记录 `config_version`，定时（每 30s）轮询 `GET platform/internal/v1/config/version`
  - version 变化 → 调用 `GET platform/internal/v1/config/:environment/*` 重新加载配置到内存
  - LLM 配置变更（provider/model/apiKey/timeout）热更新后立即生效（下一次 chat 请求使用新配置）
  - 策略变更已有 enable/disable 机制，此处不重复
  - 可选：K8s 部署时 ConfigMap 挂载 → fsnotify 监听文件变更
- **验收**：
  ① 修改 LLM model 从 gpt-4 → gpt-3.5-turbo → 30s 内 chat 服务日志显示 `config reloaded`；
  ② 新 chat 请求使用 gpt-3.5-turbo 模型；
  ③ 配置加载失败 → 保留旧配置继续运行，日志 warn
- **边界条件**：
  - 热更新期间正在处理的请求使用旧配置完成
  - 配置校验失败（如非法 URL）→ 不更新 + warn 日志 + 保持旧配置
  - version 不变 → 不重新加载

### REQ-100 会话管理

- **服务**：chat
- **优先级**：P2
- **依赖**：REQ-031（会话/Plan）、REQ-032（Chat/SSE）
- **实现要点**：
  - `chat_sessions` 表新增 `ttl_days INT DEFAULT 90`、`last_active_at DATETIME`、`status ENUM(active,archived,deleted)`
  - chat 启动时 goroutine 每小时扫描：`last_active_at < NOW() - INTERVAL ttl_days DAY AND status='active'` → 更新 status=archived
  - 归档会话不在侧边栏展示，但可通过 `GET /api/v1/chat/sessions?status=archived` 查看
  - 存储配额：每用户最大 100 个活跃会话，超出后最旧会话自动归档
  - 消息清理：归档后 30 天可配置 `auto_delete_archived_after_days`，到期物理删除
  - 手动操作：`POST /api/v1/chat/sessions/:id/archive`、`DELETE /api/v1/chat/sessions/:id`
- **验收**：
  ① 创建 101 个会话 → 第 1 个自动归档，侧边栏不显示；
  ② `GET /api/v1/chat/sessions?status=archived` → 含已归档会话；
  ③ 90 天不活跃的会话 → 1 小时内自动归档；
  ④ `POST /api/v1/chat/sessions/:id/archive` → 手动归档成功
- **边界条件**：
  - TTL 最小 7 天，最大 365 天
  - 归档会话消息只读（不可发送新消息 → 400）
  - 删除操作为软删除（status=deleted）
  - 扫描任务不影响 chat 请求性能

### REQ-101 密码策略

- **服务**：platform
- **优先级**：P1
- **依赖**：REQ-002（RBAC）
- **实现要点**：
  - `users` 表新增 `password_changed_at DATETIME`、`failed_login_count INT DEFAULT 0`、`locked_until DATETIME`
  - 密码复杂度（配置项，`security.passwordPolicy`）：
    - minLength: 12（默认）
    - requireUppercase: true
    - requireLowercase: true
    - requireDigit: true
    - requireSpecialChar: true
  - 账户锁定：连续 5 次失败 → 锁定 30 分钟
  - 密码轮换：`maxPasswordAge: 90d`，超期 → 下次登录强制改密
  - 密码历史：最近 5 次密码 hash 存储，新密码不能与历史重复
  - 首次登录强制改密（`must_change_password` 标志）
- **验收**：
  ① 注册密码 "123" → 400 `"密码不满足复杂度要求"`；
  ② 连续 5 次错误密码 → 账户锁定 30min，期间登录 → 403 `"账户已锁定"`；
  ③ 密码 90 天未改 → 登录后返回 `must_change_password:true` + 302 改密页；
  ④ 新密码与最近 5 次相同 → 400 `"不能使用最近使用过的密码"`；
  ⑤ 首次登录 → 强制改密
- **边界条件**：
  - 密码最大 128 字节
  - 锁定计数在成功登录后清零
  - admin 账户不可锁定自己（至少保留一个未锁定 admin）
  - 密码 hash 使用 bcrypt（cost=12）

### REQ-102 系统监控指标

- **服务**：全部
- **优先级**：P1
- **依赖**：REQ-019（微服务骨架）
- **实现要点**：
  - 每个服务暴露 `GET /metrics`（prometheus 格式），使用 `github.com/prometheus/client_golang`
  - 基础指标（自动）：
    - `http_request_duration_seconds`（histogram：method, path, status）
    - `http_requests_total`（counter：method, path, status）
    - `go_goroutines`、`go_memstats_*`
  - 业务指标（手动埋点）：
    - `aiops_chat_requests_total`（counter：environment, model）
    - `aiops_executions_total`（counter：environment, plugin, decision, status）
    - `aiops_policy_evaluations_total`（counter：decision, matched_rule_id）
    - `aiops_llm_request_duration_seconds`（histogram：model）
  - gateway 不认证 `/metrics` 路径（集群内 Prometheus 访问）
  - 各服务 main.go 注册 prometheus handler
- **验收**：
  ① `curl executor:8084/metrics` → 返回 prometheus 格式文本含 `http_request_duration_seconds`；
  ② 执行一次 chat → `chat:8082/metrics` 中 `aiops_chat_requests_total` +1；
  ③ 执行一次策略评估 → `policy:8083/metrics` 中 `aiops_policy_evaluations_total` +1；
  ④ `/metrics` 无需 JWT 认证
- **边界条件**：
  - metrics 端口与业务端口共用（不另开端口）
  - 高基数 label（如 user_id）不加入指标
  - histogram bucket 使用默认 Prometheus 分布

### REQ-103 多租户数据隔离

- **服务**：platform + 全部（数据层）
- **优先级**：P1
- **依赖**：REQ-001（环境）、REQ-097（多环境配置）
- **实现要点**：
  - 数据库层：所有业务表新增 `environment_slug VARCHAR(50)` 列，联合索引第一列
  - 查询过滤：GORM scope 全局 `db.Where("environment_slug = ?", envSlug)`，从 `X-AIOps-Environment` 头获取
  - platform API 强制要求 environment 参数（无则 400）
  - 写入时自动填充 environment_slug（从请求头取，不信任 body）
  - 资源配额：`environment_quotas` 表（environment_slug, resource_type: session/execution/inspection, daily_limit, current_count, reset_at DATE）
  - 配额检查：写操作前检查 `current_count < daily_limit`，超限 → 429
  - 配额重置：每日 00:00 UTC 重置 current_count
- **验收**：
  ① `GET /api/v1/chat/sessions` + `X-AIOps-Environment: dev` → 仅返回 dev 环境会话；
  ② `GET /api/v1/chat/sessions` + `X-AIOps-Environment: prod` → 不含 dev 会话；
  ③ `POST /api/v1/chat` + `X-AIOps-Environment: dev` → chat_sessions 记录 environment_slug=dev；
  ④ body 中传 environment_slug=prod 但头为 dev → 写入 dev（以头为准）；
  ⑤ 日执行超配额 → 429
- **边界条件**：
  - 跨环境查询仅 admin 角色允许（普通用户只能查自己环境）
  - 配额重置幂等（基于 reset_at 日期判断）
  - environment_slug 索引确保查询性能
  - 已有数据迁移：添加列后回填 environment_slug（默认值 "default"）

---

## 四、开发阶段 E（新增阶段）

> 在 A/B/C/D 之后新增 E 阶段，建议实现顺序按依赖关系排列。

```
E: 093→101→102→103→097→099→092→094→090→091→096→095→098→100
```

| 顺序 | REQ | 名称 | 依赖说明 |
|------|-----|------|----------|
| 1 | 093 | 凭证管理 | 无外部依赖，为后续提供加密存储基础 |
| 2 | 101 | 密码策略 | 依赖 002 RBAC，增强认证安全 |
| 3 | 102 | 系统监控指标 | 仅依赖 019 骨架，为后续提供可观测性 |
| 4 | 103 | 多租户数据隔离 | 依赖 001 环境 + 097 配置隔离（先做表结构） |
| 5 | 097 | 多环境配置隔离 | 依赖 001 环境，103 已建表结构 |
| 6 | 099 | 配置热更新 | 依赖 097 配置隔离 |
| 7 | 092 | 执行超时与重试 | 依赖 012 执行网关 |
| 8 | 094 | 变更审批流 | 依赖 033 ASK + 012 执行网关 |
| 9 | 090 | 定时调度 | 依赖 060 Worker + 061 巡检 |
| 10 | 091 | 告警通知通道 | 依赖 062 告警联动 |
| 11 | 096 | SLA/MTTR | 依赖 062 告警 + 091 通知 |
| 12 | 095 | 操作回滚 | 依赖 012 执行网关 |
| 13 | 098 | 审计导出 | 依赖 003 审计 |
| 14 | 100 | 会话管理 | 依赖 031/032 Chat |

**E 阶段 DoD**：
- [ ] 14 条 E 阶段 REQ 代码完成 + `go build` 通过
- [ ] 14 条 E 阶段 REQ 验收通过
- [ ] 新增数据库表 migrate 成功
- [ ] `/metrics` 端点所有服务可访问
- [ ] 配额限流生效
- [ ] 审批流双人确认可验证
- [ ] 定时巡检可触发

---

## 五、API 契约补全

> 对新增 REQ 涉及的 API，给出完整方法/路径/请求体/响应体。所有对外 API 经 gateway，内部 API 前缀 `/internal/v1/`。

### 5.1 定时调度（REQ-090）

#### `POST /api/v1/schedules`（创建调度）

请求体：
```json
{
  "name": "每日pod巡检",
  "cron": "0 9 * * *",
  "timezone": "Asia/Shanghai",
  "target_type": "inspection",
  "target_id": 1,
  "enabled": true
}
```

响应体 201：
```json
{
  "id": 1,
  "name": "每日pod巡检",
  "cron": "0 9 * * *",
  "timezone": "Asia/Shanghai",
  "target_type": "inspection",
  "target_id": 1,
  "enabled": true,
  "created_at": "2026-06-01T10:00:00Z",
  "updated_at": "2026-06-01T10:00:00Z"
}
```

#### `GET /api/v1/schedules`（列表）

响应体 200：
```json
{
  "items": [
    { "id":1, "name":"每日pod巡检", "cron":"0 9 * * *", "timezone":"Asia/Shanghai", "target_type":"inspection", "target_id":1, "enabled":true, "last_run_at":"2026-06-01T09:00:00Z", "next_run_at":"2026-06-02T09:00:00+08:00" }
  ],
  "total": 1
}
```

#### `PUT /api/v1/schedules/:id`（更新）

请求体同 POST，响应 200 返回更新后对象。

#### `DELETE /api/v1/schedules/:id`（删除）

响应 200：`{"message": "已删除"}`

#### `GET /api/v1/schedules/:id/executions`（执行历史）

响应体 200：
```json
{
  "items": [
    { "id":1, "schedule_id":1, "status":"completed", "started_at":"2026-06-01T09:00:00+08:00", "completed_at":"2026-06-01T09:02:30+08:00", "result_summary":"3/3 步骤成功" }
  ],
  "total": 1
}
```

#### `POST worker/internal/v1/schedules/reload`（内部：热加载）

请求体：无。响应 200：`{"loaded": 5}`

### 5.2 告警通知通道（REQ-091）

#### `POST /api/v1/notification-channels`（创建通道）

请求体：
```json
{
  "name": "企业微信运维群",
  "type": "webhook",
  "config": {
    "url": "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx",
    "method": "POST",
    "headers": {},
    "body_template": "{\"msgtype\":\"markdown\",\"markdown\":{\"content\":\"{{.Severity}} 告警: {{.Message}}\"}}"
  }
}
```

响应体 201：
```json
{
  "id": 1,
  "name": "企业微信运维群",
  "type": "webhook",
  "config": { "url": "https://...", "method": "POST" },
  "created_at": "2026-06-01T10:00:00Z"
}
```

#### `POST /api/v1/notification-policies`（创建通知策略）

请求体：
```json
{
  "name": "Critical告警通知策略",
  "severity": "critical",
  "channel_ids": [1, 2],
  "silence_window_min": 30,
  "suppress_lower_severity": true
}
```

响应体 201：
```json
{
  "id": 1,
  "name": "Critical告警通知策略",
  "severity": "critical",
  "channel_ids": [1, 2],
  "silence_window_min": 30,
  "suppress_lower_severity": true
}
```

#### `GET /api/v1/notification-templates`（模板列表）

响应体 200：
```json
{
  "items": [
    { "id":1, "channel_type":"webhook", "event_type":"alert", "subject":"","body":"【{{.Severity}}】{{.Message}} - 环境: {{.Environment}}" }
  ]
}
```

### 5.3 执行超时与重试（REQ-092）

无新增对外 API。config 变更：

```yaml
executor:
  pluginTimeouts:
    mock: 30s
    kubernetes: 120s
    prometheus: 60s
    logs: 60s
  retryPolicy:
    maxRetries: 3
    backoffBase: 1s
    backoffMax: 30s
    retryableErrors: ["timeout", "connection_refused", "5xx"]
  totalExecutionTimeout: 600s
```

`execution_records` 新增字段：
```
timeout: BOOL
retry_count: INT
retry_history: JSON  # [{attempt:1, error:"timeout", at:"2026-06-01T10:00:01Z"}]
progress: JSON       # {completed_steps:2, total_steps:3, current_step:"prometheus查询"}
```

### 5.4 凭证管理（REQ-093）

#### `POST /api/v1/credentials`（创建凭证）

请求体：
```json
{
  "name": "LLM-API-Key",
  "type": "api_key",
  "value": "sk-xxxxxxxxxxxxxxxx",
  "expires_at": "2027-01-01T00:00:00Z"
}
```

响应体 201：
```json
{
  "id": 1,
  "name": "LLM-API-Key",
  "type": "api_key",
  "value": "****",
  "expires_at": "2027-01-01T00:00:00Z",
  "created_at": "2026-06-01T10:00:00Z"
}
```

#### `GET /api/v1/credentials`（列表，脱敏）

响应体 200：
```json
{
  "items": [
    { "id":1, "name":"LLM-API-Key", "type":"api_key", "value":"****", "expires_at":"2027-01-01T00:00:00Z", "last_rotated_at":"2026-06-01T10:00:00Z" }
  ]
}
```

#### `PUT /api/v1/credentials/:id`（更新凭证值）

请求体：
```json
{
  "value": "sk-new-key-xxxxxxxx",
  "expires_at": "2027-06-01T00:00:00Z"
}
```

响应 200 返回更新后脱敏对象。

#### `GET platform/internal/v1/credentials/:id/value`（内部：获取解密值）

响应体 200：
```json
{
  "id": 1,
  "value": "sk-xxxxxxxxxxxxxxxx",
  "type": "api_key"
}
```

#### `POST /api/v1/credentials/re-encrypt`（密钥轮转）

请求体：无。响应 200：`{"re_encrypted": 5}`

### 5.5 变更审批流（REQ-094）

无新增独立 API。改造现有 confirm 接口：

#### `POST /api/v1/actions/confirm`（改造）

请求体：
```json
{
  "action_id": "abc-123",
  "approved": true,
  "reason": "确认扩容"
}
```

响应体 200（需更多审批人）：
```json
{
  "status": "partially_approved",
  "current_approvals": 1,
  "required_approvals": 2,
  "approvers": [
    { "user_id": "u1", "role": "operator", "approved_at": "2026-06-01T10:00:00Z" }
  ],
  "timeout_at": "2026-06-01T10:30:00Z"
}
```

响应体 200（审批完成）：
```json
{
  "status": "approved",
  "current_approvals": 2,
  "required_approvals": 2,
  "approvers": [
    { "user_id": "u1", "role": "operator", "approved_at": "2026-06-01T10:00:00Z" },
    { "user_id": "u2", "role": "admin", "approved_at": "2026-06-01T10:05:00Z" }
  ],
  "execution_started": true
}
```

响应体 409（审批人重复）：
```json
{
  "code": 409,
  "message": "您已审批过此操作",
  "request_id": "req-xxx"
}
```

策略 spec 新增 `approvalConfig`：
```yaml
approvalConfig:
  write: { minApprovers: 2, timeout: 30m, timeoutAction: REJECT }
  notify: { minApprovers: 1, timeout: 15m, timeoutAction: REJECT }
```

### 5.6 操作回滚（REQ-095）

#### `POST /api/v1/actions/:id/rollback`（手动回滚）

请求体：
```json
{
  "reason": "扩容导致资源不足"
}
```

响应体 200：
```json
{
  "rollback_execution_id": 42,
  "parent_execution_id": 41,
  "status": "pending",
  "rollback_plan": {
    "plugin": "kubernetes",
    "action": "scale",
    "parameters": { "namespace": "default", "deployment": "api", "replicas": 2 }
  }
}
```

响应体 400（无回滚计划）：
```json
{
  "code": 400,
  "message": "无可用回滚计划",
  "request_id": "req-xxx"
}
```

### 5.7 SLA/MTTR（REQ-096）

#### `POST /api/v1/sla-definitions`（创建 SLA 定义）

请求体：
```json
{
  "name": "P1响应SLA",
  "severity": "critical",
  "response_time_min": 15,
  "resolution_time_min": 60,
  "escalation_channel_id": 1
}
```

响应体 201：
```json
{
  "id": 1,
  "name": "P1响应SLA",
  "severity": "critical",
  "response_time_min": 15,
  "resolution_time_min": 60,
  "escalation_channel_id": 1
}
```

#### `GET /api/v1/sla/stats`（SLA 统计）

查询参数：`period=30d`、`environment=dev`

响应体 200：
```json
{
  "period": "30d",
  "environment": "dev",
  "severity_stats": {
    "critical": { "total": 10, "response_sla_met": 8, "resolution_sla_met": 7, "avg_mttr_min": 45.2 },
    "high": { "total": 25, "response_sla_met": 22, "resolution_sla_met": 20, "avg_mttr_min": 92.1 }
  },
  "overall_mttr_min": 78.6,
  "overall_sla_rate": 0.857
}
```

### 5.8 多环境配置隔离（REQ-097）

#### `POST /api/v1/environments/:slug/config`（设置环境配置）

请求体：
```json
{
  "key": "plugins.kubernetes",
  "value": { "endpoint": "http://k8s-dev:8091", "mode": "inCluster" },
  "override_type": "plugin"
}
```

响应体 200：
```json
{
  "environment_slug": "dev",
  "key": "plugins.kubernetes",
  "value": { "endpoint": "http://k8s-dev:8091", "mode": "inCluster" },
  "override_type": "plugin",
  "updated_at": "2026-06-01T10:00:00Z"
}
```

#### `GET platform/internal/v1/config/:environment/:key`（内部：获取合并配置）

响应体 200：
```json
{
  "environment": "dev",
  "key": "plugins.kubernetes",
  "value": { "endpoint": "http://k8s-dev:8091", "mode": "inCluster" },
  "source": "environment_override"
}
```

### 5.9 审计导出（REQ-098）

#### `GET /api/v1/audit/export`（导出）

查询参数：`format=csv|pdf`、`environment=dev`、`start=2026-05-01`、`end=2026-06-01`

CSV 响应 200：
- Content-Type: `text/csv; charset=utf-8`
- Content-Disposition: `attachment; filename=audit_export_20260601.csv`
- 首行 BOM + 列名：`id,timestamp,username,action,environment,resource_type,resource_id,result,error_message`
- 后续行数据

PDF 响应 200：
- Content-Type: `application/pdf`
- Content-Disposition: `attachment; filename=audit_export_20260601.pdf`

#### `GET /api/v1/audit/report`（合规报告）

查询参数：`template=compliance`、`period=monthly`、`environment=dev`

响应体 200：
```json
{
  "template": "compliance",
  "period": "2026-05",
  "environment": "dev",
  "total_events": 1234,
  "by_action": { "execute": 500, "confirm": 300, "login": 434 },
  "by_risk": { "read": 800, "write": 400, "notify": 34 },
  "denied_count": 12,
  "sla_compliance_rate": 0.95,
  "export_url": "/api/v1/audit/export?format=pdf&environment=dev&start=2026-05-01&end=2026-06-01"
}
```

### 5.10 配置热更新（REQ-099）

#### `GET platform/internal/v1/config/version`（内部：配置版本号）

响应体 200：
```json
{
  "version": 42,
  "updated_at": "2026-06-01T10:30:00Z"
}
```

#### `GET platform/internal/v1/config/:environment`（内部：全量配置）

响应体 200：
```json
{
  "environment": "dev",
  "llm": { "provider": "openai-compatible", "model": "gpt-3.5-turbo", "timeout": "60s" },
  "plugins": { "kubernetes": { "endpoint": "http://k8s-dev:8091" } },
  "version": 42
}
```

### 5.11 会话管理（REQ-100）

#### `POST /api/v1/chat/sessions/:id/archive`（手动归档）

响应体 200：
```json
{
  "id": "sess-123",
  "status": "archived",
  "archived_at": "2026-06-01T10:00:00Z"
}
```

#### `DELETE /api/v1/chat/sessions/:id`（软删除）

响应体 200：`{"message": "已删除"}`

#### `GET /api/v1/chat/sessions`（改造：新增 status 过滤）

查询参数：`status=active|archived`（默认 active）、`page=1`、`page_size=20`

### 5.12 密码策略（REQ-101）

#### `PUT /api/v1/users/:id/password`（改密）

请求体：
```json
{
  "old_password": "OldPass123!",
  "new_password": "NewPass456!"
}
```

响应体 200：
```json
{
  "message": "密码修改成功",
  "must_change_password": false
}
```

响应体 400（不满足复杂度）：
```json
{
  "code": 400,
  "message": "密码不满足复杂度要求：至少12位，含大小写字母、数字、特殊字符",
  "request_id": "req-xxx"
}
```

#### `POST /api/v1/auth/login`（改造：返回 must_change_password）

响应体 200（需改密）：
```json
{
  "token": "jwt-xxx",
  "must_change_password": true,
  "password_expires_at": "2026-05-01T00:00:00Z"
}
```

### 5.13 系统监控指标（REQ-102）

#### `GET /metrics`（每个服务）

- 无需认证
- Content-Type: `text/plain; version=0.0.4; charset=utf-8`
- 格式：Prometheus exposition format
- 关键指标：
  - `http_request_duration_seconds_bucket{method="POST",path="/api/v1/chat",status="200",le="0.1"} 42`
  - `aiops_chat_requests_total{environment="dev",model="gpt-4"} 150`
  - `aiops_executions_total{environment="dev",plugin="kubernetes",decision="ALLOW",status="completed"} 88`
  - `aiops_policy_evaluations_total{decision="DENY",matched_rule_id="hardDeny"} 5`
  - `aiops_llm_request_duration_seconds_bucket{model="gpt-4",le="5"} 30`

### 5.14 多租户数据隔离（REQ-103）

无新增独立 API。改造现有 API：
- 所有 `GET /api/v1/*` 列表接口新增隐式 `WHERE environment_slug = ?` 过滤
- 所有 `POST /api/v1/*` 写入接口自动填充 `environment_slug`（来自 `X-AIOps-Environment` 头）

#### `GET /api/v1/environments/:slug/quotas`（环境配额查询）

响应体 200：
```json
{
  "environment": "dev",
  "quotas": {
    "sessions": { "daily_limit": 1000, "current_count": 234, "reset_at": "2026-06-02" },
    "executions": { "daily_limit": 500, "current_count": 89, "reset_at": "2026-06-02" },
    "inspections": { "daily_limit": 100, "current_count": 12, "reset_at": "2026-06-02" }
  }
}
```

#### `PUT /api/v1/environments/:slug/quotas`（设置配额）

请求体：
```json
{
  "resource_type": "executions",
  "daily_limit": 500
}
```

---

## 六、数据模型补全

> 对新增 REQ 涉及的 MySQL 表，给出完整字段定义。所有表使用 InnoDB，utf8mb4，默认 `created_at`/`updated_at`。

### 6.1 schedules（定时调度 - REQ-090）

```sql
CREATE TABLE schedules (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(100) NOT NULL COMMENT '调度名称',
  cron VARCHAR(50) NOT NULL COMMENT 'cron 表达式',
  timezone VARCHAR(50) NOT NULL DEFAULT 'UTC' COMMENT 'IANA 时区',
  target_type ENUM('inspection','runbook') NOT NULL COMMENT '目标类型',
  target_id BIGINT UNSIGNED NOT NULL COMMENT '目标 ID',
  enabled TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否启用',
  environment_slug VARCHAR(50) NOT NULL COMMENT '环境标识',
  last_run_at DATETIME NULL COMMENT '上次执行时间',
  next_run_at DATETIME NULL COMMENT '下次预计执行时间',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_env_enabled (environment_slug, enabled),
  INDEX idx_next_run (enabled, next_run_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 6.2 schedule_executions（调度执行历史 - REQ-090）

```sql
CREATE TABLE schedule_executions (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  schedule_id BIGINT UNSIGNED NOT NULL COMMENT '调度 ID',
  status ENUM('running','completed','failed','skipped','timeout') NOT NULL DEFAULT 'running',
  started_at DATETIME NOT NULL COMMENT '开始时间',
  completed_at DATETIME NULL COMMENT '完成时间',
  result_summary VARCHAR(500) NULL COMMENT '结果摘要',
  error_message TEXT NULL COMMENT '错误信息',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_schedule (schedule_id, started_at DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 6.3 notification_channels（通知通道 - REQ-091）

```sql
CREATE TABLE notification_channels (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(100) NOT NULL COMMENT '通道名称',
  type ENUM('email','webhook','sms') NOT NULL COMMENT '通道类型',
  config JSON NOT NULL COMMENT '通道配置（含 URL/SMTP 等，敏感字段加密后存入）',
  enabled TINYINT(1) NOT NULL DEFAULT 1,
  environment_slug VARCHAR(50) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE INDEX idx_env_name (environment_slug, name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 6.4 notification_policies（通知策略 - REQ-091）

```sql
CREATE TABLE notification_policies (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(100) NOT NULL COMMENT '策略名称',
  severity ENUM('critical','high','medium','low') NOT NULL COMMENT '告警级别',
  channel_ids JSON NOT NULL COMMENT '关联通道 ID 数组',
  silence_window_min INT NOT NULL DEFAULT 30 COMMENT '静默窗口（分钟）',
  suppress_lower_severity TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否抑制低级别',
  environment_slug VARCHAR(50) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE INDEX idx_env_severity (environment_slug, severity)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 6.5 notification_templates（通知模板 - REQ-091）

```sql
CREATE TABLE notification_templates (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  channel_type ENUM('email','webhook','sms') NOT NULL,
  event_type VARCHAR(50) NOT NULL COMMENT '事件类型：alert/execution/sla',
  subject VARCHAR(200) NULL COMMENT '邮件主题（email 专用）',
  body TEXT NOT NULL COMMENT '模板正文，支持 Go template 语法',
  environment_slug VARCHAR(50) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE INDEX idx_type_event (channel_type, event_type, environment_slug)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 6.6 credentials（凭证管理 - REQ-093）

```sql
CREATE TABLE credentials (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(100) NOT NULL COMMENT '凭证名称',
  type ENUM('smtp','api_key','database','kubeconfig','other') NOT NULL COMMENT '凭证类型',
  encrypted_value TEXT NOT NULL COMMENT 'AES-256-GCM 加密后的值',
  expires_at DATETIME NULL COMMENT '过期时间',
  last_rotated_at DATETIME NULL COMMENT '上次轮转时间',
  environment_slug VARCHAR(50) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE INDEX idx_env_name (environment_slug, name),
  INDEX idx_expires (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 6.7 approval_policies（审批策略 - REQ-094）

```sql
CREATE TABLE approval_policies (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  risk_level ENUM('write','notify') NOT NULL COMMENT '风险级别',
  min_approvers INT NOT NULL DEFAULT 1 COMMENT '最小审批人数',
  timeout_min INT NOT NULL DEFAULT 30 COMMENT '审批超时（分钟）',
  timeout_action ENUM('REJECT','ESCALATE') NOT NULL DEFAULT 'REJECT' COMMENT '超时动作',
  require_different_roles TINYINT(1) NOT NULL DEFAULT 1 COMMENT '审批人是否需不同角色',
  environment_slug VARCHAR(50) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE INDEX idx_env_risk (environment_slug, risk_level)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 6.8 execution_records 扩展（REQ-092/094/095）

在现有 `execution_records` 表新增字段：

```sql
ALTER TABLE execution_records
  ADD COLUMN timeout TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否超时',
  ADD COLUMN retry_count INT NOT NULL DEFAULT 0 COMMENT '重试次数',
  ADD COLUMN retry_history JSON NULL COMMENT '重试历史',
  ADD COLUMN progress JSON NULL COMMENT '执行进度',
  ADD COLUMN approvers JSON NULL COMMENT '审批人列表 [{user_id,role,approved_at}]',
  ADD COLUMN approval_status ENUM('pending','partially_approved','approved','rejected','rejected_timeout') NULL COMMENT '审批状态',
  ADD COLUMN approval_timeout_at DATETIME NULL COMMENT '审批超时时间',
  ADD COLUMN pre_snapshot JSON NULL COMMENT '执行前快照（用于回滚）',
  ADD COLUMN parent_execution_id BIGINT UNSIGNED NULL COMMENT '回滚关联的原始执行 ID',
  ADD COLUMN rollback_policy ENUM('manual','auto_on_failure') NULL COMMENT '回滚策略',
  ADD COLUMN environment_slug VARCHAR(50) NOT NULL DEFAULT 'default' COMMENT '环境标识';

CREATE INDEX idx_exec_env ON execution_records(environment_slug);
CREATE INDEX idx_exec_approval ON execution_records(approval_status, approval_timeout_at);
```

### 6.9 sla_definitions（SLA 定义 - REQ-096）

```sql
CREATE TABLE sla_definitions (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(100) NOT NULL COMMENT 'SLA 名称',
  severity ENUM('critical','high','medium','low') NOT NULL,
  response_time_min INT NOT NULL COMMENT '响应时间（分钟）',
  resolution_time_min INT NOT NULL COMMENT '解决时间（分钟）',
  escalation_channel_id BIGINT UNSIGNED NULL COMMENT 'SLA 逾期通知通道',
  environment_slug VARCHAR(50) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE INDEX idx_env_severity (environment_slug, severity)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 6.10 sla_records（SLA 记录 - REQ-096）

```sql
CREATE TABLE sla_records (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  alert_id BIGINT UNSIGNED NOT NULL COMMENT '关联告警 ID',
  sla_definition_id BIGINT UNSIGNED NOT NULL COMMENT '关联 SLA 定义 ID',
  created_at DATETIME NOT NULL COMMENT '告警创建时间',
  acknowledged_at DATETIME NULL COMMENT '首次响应时间',
  resolved_at DATETIME NULL COMMENT '解决时间',
  response_sla_met TINYINT(1) NULL COMMENT '响应 SLA 是否达标',
  resolution_sla_met TINYINT(1) NULL COMMENT '解决 SLA 是否达标',
  escalated TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否已升级通知',
  environment_slug VARCHAR(50) NOT NULL,
  INDEX idx_alert (alert_id),
  INDEX idx_sla_def (sla_definition_id),
  INDEX idx_env_time (environment_slug, created_at DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 6.11 environment_configs（环境配置 - REQ-097）

```sql
CREATE TABLE environment_configs (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  environment_slug VARCHAR(50) NOT NULL COMMENT '环境标识',
  config_key VARCHAR(200) NOT NULL COMMENT '配置键（点分路径如 plugins.kubernetes）',
  config_value JSON NOT NULL COMMENT '配置值',
  override_type ENUM('plugin','service','llm','security','other') NOT NULL DEFAULT 'other',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE INDEX idx_env_key (environment_slug, config_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 6.12 config_versions（配置版本 - REQ-099）

```sql
CREATE TABLE config_versions (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  version INT UNSIGNED NOT NULL COMMENT '配置版本号',
  changed_by VARCHAR(100) NULL COMMENT '变更人',
  change_summary VARCHAR(500) NULL COMMENT '变更摘要',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE INDEX idx_version (version DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 6.13 environment_quotas（环境配额 - REQ-103）

```sql
CREATE TABLE environment_quotas (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  environment_slug VARCHAR(50) NOT NULL COMMENT '环境标识',
  resource_type ENUM('session','execution','inspection') NOT NULL COMMENT '资源类型',
  daily_limit INT NOT NULL DEFAULT 1000 COMMENT '每日上限',
  current_count INT NOT NULL DEFAULT 0 COMMENT '当日已用',
  reset_at DATE NOT NULL COMMENT '重置日期',
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE INDEX idx_env_resource (environment_slug, resource_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 6.14 chat_sessions 扩展（REQ-100）

在现有 `chat_sessions` 表新增字段：

```sql
ALTER TABLE chat_sessions
  ADD COLUMN ttl_days INT NOT NULL DEFAULT 90 COMMENT '会话 TTL（天）',
  ADD COLUMN last_active_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '最后活跃时间',
  ADD COLUMN status ENUM('active','archived','deleted') NOT NULL DEFAULT 'active' COMMENT '会话状态',
  ADD COLUMN environment_slug VARCHAR(50) NOT NULL DEFAULT 'default' COMMENT '环境标识';

CREATE INDEX idx_session_env_status (environment_slug, status, last_active_at);
```

### 6.15 users 扩展（REQ-101）

在现有 `users` 表新增字段：

```sql
ALTER TABLE users
  ADD COLUMN password_changed_at DATETIME NULL COMMENT '密码修改时间',
  ADD COLUMN failed_login_count INT NOT NULL DEFAULT 0 COMMENT '连续登录失败次数',
  ADD COLUMN locked_until DATETIME NULL COMMENT '锁定截止时间',
  ADD COLUMN must_change_password TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否需要修改密码',
  ADD COLUMN password_history JSON NULL COMMENT '最近5次密码hash';

CREATE INDEX idx_user_locked (locked_until);
```

### 6.16 security_boundary_policies 扩展（REQ-094/095）

在现有 `security_boundary_policies` 表新增字段：

```sql
ALTER TABLE security_boundary_policies
  ADD COLUMN approval_config JSON NULL COMMENT '审批配置 {write:{minApprovers:2,timeout:"30m"},notify:{minApprovers:1,timeout:"15m"}}',
  ADD COLUMN rollback_plans JSON NULL COMMENT '回滚计划 [{match:{plugin,action},rollback:{plugin,action,parameters_from_snapshot}}]',
  ADD COLUMN rollback_policy ENUM('manual','auto_on_failure') NULL DEFAULT 'manual' COMMENT '回滚策略';
```

### 6.17 inspection_steps 扩展（REQ-092）

在巡检步骤表（如有独立表）或 inspections JSON 步骤中新增字段：

```sql
-- 如 inspections 表有 steps JSON 字段，在 JSON 内新增：
-- retry_count: INT (默认0)
-- max_retries: INT (默认0)
-- timeout: VARCHAR(20) (如 "60s", 默认继承 pluginTimeouts)
```

---

## 附录：config.yaml 新增配置段

```yaml
# REQ-092: 执行超时与重试
executor:
  pluginTimeouts:
    mock: 30s
    kubernetes: 120s
    prometheus: 60s
    logs: 60s
  retryPolicy:
    maxRetries: 3
    backoffBase: 1s
    backoffMax: 30s
    retryableErrors: ["timeout", "connection_refused", "5xx"]
  totalExecutionTimeout: 600s

# REQ-093: 凭证加密密钥（生产环境从 K8s Secret 注入）
encryption:
  keyEnvVar: AIOPS_ENCRYPTION_KEY

# REQ-101: 密码策略
security:
  passwordPolicy:
    minLength: 12
    requireUppercase: true
    requireLowercase: true
    requireDigit: true
    requireSpecialChar: true
  maxPasswordAge: 90d
  passwordHistoryCount: 5
  maxFailedLogins: 5
  lockoutDuration: 30m

# REQ-100: 会话管理
chat:
  sessionTTLDefault: 90d
  maxActiveSessionsPerUser: 100
  autoDeleteArchivedAfterDays: 30
  archiveCheckInterval: 1h

# REQ-099: 配置热更新
configReload:
  pollInterval: 30s
  platformEndpoint: http://aiops-platform:8081
```

---

## 附录：E 阶段前端 REQ 建议

为配合后端 E 阶段 14 条新 REQ，建议前端新增以下 REQ：

| 前端 REQ | 对应后端 | 说明 |
|----------|----------|------|
| F-040 | 090 | 定时调度管理页：CRUD + 执行历史列表 |
| F-041 | 091 | 通知通道配置页：通道 CRUD + 通知策略 + 模板编辑 + 测试发送 |
| F-042 | 093 | 凭证管理页：凭证列表（脱敏）+ 新建/轮转/过期提醒 |
| F-043 | 094 | 审批流改造：待审批列表 + 多人审批状态展示 + 审批超时提示 |
| F-044 | 095 | 回滚操作：执行详情页新增"回滚"按钮 + 回滚状态追踪 |
| F-045 | 096 | SLA 仪表盘：SLA 达标率图表 + MTTR 趋势 + SLA 定义配置 |
| F-046 | 097 | 环境配置覆盖：环境详情页新增配置编辑 Tab |
| F-047 | 098 | 审计导出：审计页新增导出按钮（CSV/PDF）+ 合规报告 |
| F-048 | 100 | 会话管理：侧边栏归档分区 + 会话归档/删除操作 |
| F-049 | 101 | 改密页面 + 密码过期提示 |
| F-050 | 102 | 系统监控页：Prometheus Grafana 嵌入或自建指标面板 |
| F-051 | 103 | 环境配额展示：环境详情页显示配额使用量进度条 |
