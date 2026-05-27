# AIOps 前端 PRD

| 版本 | v1.0 |
| 部署 | K8s `aiops`；Nginx 静态资源 + 反向代理，域名 `https://aiops.tclpv.com/` |
| 后端需求台账 | [DEVELOPMENT_PLAN.md](./DEVELOPMENT_PLAN.md)（30 条 REQ，已完成） |

## 0. 规则

| ID | 规则 |
|----|------|
| F-001 | 前端独立于后端代码仓库，不在 Go 项目中嵌入 SPA |
| F-002 | 所有 API 请求经 Gateway (`/api/v1/*`)，统一携带 JWT Authorization 头 |
| F-003 | Chat 流式对话使用 SSE (Server-Sent Events)，前端处理 `text/event-stream` |
| F-004 | 权限粒度：页面可见性基于角色（admin/operator/readonly/auditor），readonly 禁止写操作和确认执行 |
| F-005 | 数据面板类（审计、指标、日志）必须分页，默认每页 20 条，单次数据查询行数 ≤500 |
| F-006 | 策略 YAML 编辑器需实时语法高亮；模拟评估结果即时展示 decision 颜色标记 |
| F-007 | WebSocket 或 SSE 连接失败自动重连（退避策略），聊天界面重连后恢复历史消息 |

## 1. 技术选型

| 维度 | 选择 | 理由 |
|------|------|------|
| 框架 | React 18 + TypeScript | 生态成熟，企业级项目首选 |
| 构建 | Vite 5 | 开发体验快，HMR |
| UI 库 | Ant Design 5 | 企业后台组件丰富 |
| 路由 | React Router 6 | 嵌套路由，懒加载 |
| 状态管理 | Zustand | 轻量，TS 友好 |
| 图表 | ECharts + echarts-for-react | 监控面板必备 |
| HTTP | Axios | 拦截器统一处理 JWT/401/错误 |
| 代码编辑器 | Monaco Editor (`@monaco-editor/react`) | YAML 策略编辑、JSON 数据展示 |
| Markdown | react-markdown + rehype-highlight | AI 对话消息渲染 |

## 2. 页面树

```
/login                      # 登录页
/register                   # 注册页（可选）
/dashboard                  # 概览仪表盘
/environments               # 环境管理（列表/新增/编辑）
/policies                   # 安全边界策略（列表/新建/编辑/模拟/模板导入）
/policies/:id/simulate      # 策略模拟器
/audit                      # 审计日志
/plugins                    # 插件列表
/knowledge                  # 知识库（文档列表/新建/编辑）
/data-query                 # 数据查询面板
/inspections                # 巡检任务（列表/新建/报告详情）
/inspections/:id/reports    # 巡检报告列表
/risk-alerts                # 高危预警（列表/扫描）
/runbooks                   # 自愈手册（列表/新建/执行记录）
/im                         # IM 机器人配置
/helpdesk                   # HelpDesk 服务台（管理端）
/chat                       # AI 对话主界面（核心页面）
/chat/sessions/:id          # 历史对话详情
/actions                    # 待确认操作（pending 列表）
```

## 3. 模块详细设计

### 3.1 登录/注册

- **API**：`POST /api/v1/auth/login`
- **页面**：居中卡片，用户名+密码表单
- **逻辑**：成功后存储 token 到 localStorage，跳转 `/dashboard`
- **401 处理**：Axios 响应拦截器自动跳转 `/login`
- **JWT 过期**：前端解析 exp，提前 5 分钟提示刷新

### 3.2 概览仪表盘（Dashboard）

**数据来源**：聚合多个 API
- **指标卡片**（4 格）：环境数量、策略数量（已启用/总数）、待确认操作数、今日审计条目
- **近期操作**（时序列表）：最近 20 条审计日志
- **系统状态**：各插件在线状态（调用 `/api/v1/plugins`）

**API 依赖**：
| 数据 | API |
|------|-----|
| 环境数 | `GET /api/v1/environments` |
| 策略数 | `GET /api/v1/security-boundaries` |
| 待确认 | `GET /api/v1/actions/pending` |
| 审计 | `GET /api/v1/audit` |
| 插件 | `GET /api/v1/plugins` |

### 3.3 AI 对话（核心页面 Chat）

**API**：`POST /api/v1/chat`（非流式）、`GET /api/v1/chat/stream`（SSE 流式）

**页面布局**（三栏）：
```
┌──────────────────────────────────────────────┐
│ Header: 环境选择器 │ 会话标题                 │
├─────────┬───────────────────────┬────────────┤
│ 历史    │                       │ 当前会话    │
│ 会话    │    对话消息区          │ 详情面板    │
│ 列表    │    (Markdown渲染)     │ - 已解析    │
│         │    (ActionPlan卡片)   │   ActionPlan│
│         │    (代码高亮)         │ - pending   │
│         │                       │   确认按钮  │
├─────────┴───────────────────────┴────────────┤
│ 输入框（支持 Shift+Enter 换行）               │
│ [环境: development ▼]          [发送]         │
└──────────────────────────────────────────────┘
```

**核心交互**：
1. 选择环境（必填，从 `/api/v1/environments` 获取）
2. 输入自然语言，点击发送 → `POST /api/v1/chat`
3. SSE 流式模式：逐 token 渲染回复
4. ActionPlan 卡片：解析 `risk` 字段 → 颜色标记（read=蓝，write=橙，notify=灰）
5. **ASK 决策**：消息中展示 [确认执行] [拒绝] 按钮 → `POST /api/v1/actions/confirm`

**会话管理**：
- 左侧历史会话列表 `GET /api/v1/chat/sessions`
- 切换会话 `GET /api/v1/chat/sessions/:id/messages`
- 新建对话：清空输入框 + 重新选择环境

**辅助功能**（右侧面板可切换）：
- 修复辅助：`POST /api/v1/chat/assist`  
- RCA 根因分析：`POST /api/v1/chat/rca`

### 3.4 环境管理（Environments）

**API**：`GET/POST /api/v1/environments`
- **列表页**：表格（名称、Slug、描述、创建时间）
- **新建/编辑**：Modal 表单（name、slug、description）
- **slug 规则**：小写英文+连字符，前端校验

### 3.5 安全边界策略（Policies）

**API**：`GET/POST/PUT /api/v1/security-boundaries`、`enable/disable`、`simulate`、`import-template`

**列表页**：表格（名称、状态开关 enable/disable、规则数、更新时间、操作按钮）

**新建/编辑页**：
```
┌─────────────────────────────────────────────┐
│ 策略名称：[________________]                  │
│ 默认决策：[ALLOW ▼] [ASK] [DENY]             │
├─────────────────────────────────────────────┤
│ YAML 编辑器（Monaco Editor）                  │
│ ┌─────────────────────────────────────────┐ │
│ │ defaultDecision: ASK                     │ │
│ │ rules:                                   │ │
│ │   - id: r1                               │ │
│ │     match:                               │ │
│ │       plugin: kubernetes                 │ │
│ │       actions: [list, get]               │ │
│ │       commandPattern: ""                 │ │
│ │     decision: ALLOW                      │ │
│ │     message: 允许只读K8s操作              │ │
│ └─────────────────────────────────────────┘ │
│                                              │
│ [模拟评估]  [导入模板]  [保存] [返回]         │
└─────────────────────────────────────────────┘
```

**策略模拟器**（`/policies/:id/simulate`）：
- 左侧：输入模拟的 ActionPlan JSON
- 右侧：显示评估结果（decision + matched_rule + message），颜色标记

**导入模板**：`GET /api/v1/security-boundaries/import-template` → 预置模板列表，选择后填充到编辑器

### 3.6 审计日志（Audit）

**API**：`GET /api/v1/audit?environment=xxx`
- **表格**：时间、用户、操作、环境、资源类型、资源ID、结果
- **过滤器**：环境下拉选择、时间范围、关键词搜索
- **分页**：后端分页，前端传 page/pageSize

### 3.7 插件列表（Plugins）

**API**：`GET /api/v1/plugins`
- **卡片布局**：插件名称、状态（在线/离线）、类型、描述
- **在线状态**：每个插件单独 `/healthz` 探活

### 3.8 知识库（Knowledge Base）

**API**：`GET/POST/DELETE /api/v1/kb/documents`

**列表页**：表格（标题、分类、创建时间、操作）

**新建/编辑**：Markdown 编辑器
- 标题输入框
- Markdown 正文编辑器（支持预览）
- 保存 → `POST /api/v1/kb/documents`

### 3.9 数据查询（Data Query）

**API**：`POST /api/v1/data/query`
- **查询面板**：
  - 插件选择：Kubernetes / Prometheus / Logs
  - 环境选择
  - 查询输入（自然语言或 DSL）
- **结果展示**：
  - K8s：表格（pod/deployment 列）
  - Prometheus：ECharts 时序折线图
  - Logs：日志列表，支持关键词高亮
- **限制提示**：超过 maxRows 时截断并提示

### 3.10 巡检（Inspections）

**API**：`GET/POST /api/v1/inspections`、`run`、`reports`

**巡检列表**：表格（名称、环境、步骤数、最近执行时间、状态）

**新建巡检**：多步骤表单
- 基本信息：名称、环境
- 步骤编辑器：拖拽排序（插件+action+参数）

**执行**：点击 [立即执行] → `POST /api/v1/inspections/:id/run` → 轮询或 SSE 等待结果

**报告详情**：`GET /api/v1/inspections/:id/reports`
- 步骤级执行结果（成功/失败/耗时）
- 失败步骤高亮 + 错误信息

### 3.11 高危预警（Risk Alerts）

**API**：`GET /api/v1/risk-alerts`、`scan`
- **列表**：卡片/表格展示预警（级别、描述、来源、时间）
- **扫描**：[立即扫描] 按钮 → `POST /api/v1/risk-alerts/scan`
- **预警级别**：Critical（红）、High（橙）、Medium（黄）、Low（蓝）

### 3.12 Runbook 自愈手册

**API**：`GET/POST /api/v1/runbooks`、`execute`

**列表页**：表格（名称、关联策略、步骤数、最近执行、成功次数/失败次数）

**详情/执行**：
- 查看 Runbook 步骤
- [手动执行] → `POST /api/v1/runbooks/:id/execute`
- 执行记录列表（状态、时间、结果）

### 3.13 待确认操作（Actions）

**API**：`GET /api/v1/actions/pending`、`POST /api/v1/actions/confirm`
- **列表页**：表格（时间、用户、操作摘要、命令预览、环境）
- **操作列**：[确认执行] [拒绝]
- **红色标记**：高危命令（硬禁止模式匹配）不可确认
- **轮询**：每 10 秒自动刷新待确认列表

### 3.14 HelpDesk / IM

**HelpDesk**：`POST /api/v1/helpdesk/chat`
- 类似 Chat 界面的简化版本
- 无环境选择，固定 helpdesk 上下文

**IM Webhook**：`POST /api/v1/im/webhook`
- 配置页面：展示 Webhook URL + 使用说明
- 测试发送功能

## 4. 全局组件

| 组件 | 说明 |
|------|------|
| `AuthGuard` | 路由守卫，未登录跳转 `/login` |
| `RoleGuard` | 角色权限，readonly 隐藏写操作按钮 |
| `PageHeader` | 顶部导航栏：Logo + 主菜单 + 用户头像下拉（退出登录） |
| `Sidebar` | 侧边导航：Dashboard/Chat/环境/策略/审计/巡检/知识库/预警/Runbook/待确认 |
| `EnvironmentSelector` | 环境选择器下拉，Chat/数据查询/巡检创建中使用 |
| `YamlEditor` | Monaco Editor 封装，YAML 语法高亮+校验 |
| `ActionPlanCard` | ActionPlan JSON 渲染卡片，risk 颜色区分 |
| `StreamingMessage` | SSE 流式消息渲染，打字机效果 |
| `ErrorBoundary` | 全局错误边界 |

## 5. 路由权限矩阵

| 页面 | admin | operator | readonly | auditor |
|------|-------|----------|----------|---------|
| Dashboard | ✅ | ✅ | ✅ | ✅ |
| Chat | ✅ | ✅ | ❌ | ❌ |
| Environments | ✅ | ✅ | ❌ | ❌ |
| Policies | ✅ | ✅ | ❌ | ❌ |
| Audit | ✅ | ✅ | ✅ | ✅ |
| Plugins | ✅ | ✅ | ✅ | ✅ |
| Knowledge Base | ✅ | ✅ | ❌ | ❌ |
| Data Query | ✅ | ✅ | ✅ | ✅ |
| Inspections | ✅ | ✅ | ❌ | ❌ |
| Risk Alerts | ✅ | ✅ | ❌ | ❌ |
| Runbooks | ✅ | ✅ | ❌ | ❌ |
| Confirm Actions | ✅ | ✅ | ❌ | ❌ |
| HelpDesk | ✅ | ✅ | ✅ | ✅ |
| IM Config | ✅ | ❌ | ❌ | ❌ |

## 6. 错误处理规范

| 场景 | 处理 |
|------|------|
| 401 Unauthorized | 清除 token，跳转 login，toast 提示"登录已过期" |
| 403 Forbidden | toast "无权限操作" |
| 4xx 参数错误 | 展示后端返回的 error message |
| 5xx 服务错误 | toast "服务异常，请稍后重试" |
| 网络超时 (30s) | toast "请求超时" |
| SSE 断连 | 显示重连提示，5s 后退避重连 |

## 7. 性能要求

| 指标 | 目标 |
|------|------|
| 首屏加载 (LCP) | < 2s |
| 路由切换 | < 300ms |
| 聊天消息渲染 | < 100ms（非流式单条） |
| 表格大数据集 | 虚拟滚动（超过 1000 行） |
| 构建产物 | Gzip < 500KB |

## 8. 部署架构

```
Client Browser
      │
      ▼
K8s Ingress (aiops.tclpv.com)
      │
      ├── /api/*  → aiops-gateway:8080  (后端 API)
      │
      └── /*      → aiops-frontend:80   (前端 SPA 静态资源)
                      │
                Nginx (SPA fallback: try_files $uri /index.html)
```

- 前端构建产物：`dist/` 目录
- Dockerfile：Nginx 基础镜像，COPY dist → /usr/share/nginx/html
- Nginx 配置：SPA history mode fallback + `/api` 反向代理到 gateway
- Helm：新增 `aiops-frontend` Deployment + Service

## 9. 产品 DoD

- [ ] 14 个页面全部可用
- [ ] JWT 登录/过期/刷新流程完整
- [ ] Chat SSE 流式对话 + ActionPlan 确认完整
- [ ] 策略 YAML 编辑器 + 模拟器可用
- [ ] 角色权限矩阵生效
- [ ] 所有后端 API 有对应前端页面
- [ ] 响应式布局（1920px / 1440px / 1366px）
- [ ] Dockerfile + Helm 部署可运行
- [ ] 通过 K8s Ingress 访问验收
