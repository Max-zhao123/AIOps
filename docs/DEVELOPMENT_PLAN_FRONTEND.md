# AIOps 前端开发计划

> **AI 实现前端时的第一必读文件。** 逐条实现下表全部 REQ；契约见 [PRD_FRONTEND.md](./PRD_FRONTEND.md)。
> 后端 API 已完成 30 条 REQ 联调验收，前端配合后端 API 开发。

## 开发节奏

| 阶段 | 做什么 | 禁止 |
|------|--------|------|
| **① 开发** | 下表 **全部 28 条 REQ** React/TypeScript 代码 + `vite build` 通过 | 上 K8s、改 Helm、跳过任何 REQ |
| **② 联调** | ① 完成后 Docker 构建 + Helm 部署 + 逐 REQ 验收 | 未写完代码就部署 |

| 当前阶段 | **待开发** |
| 前端 REQ | **28** |
| 代码完成 | **0 / 28** |
| 联调 `done` | **0 / 28** |
| 最后更新 | 2026-05-27 |

**技术栈**：React 18 + TypeScript + Vite 5 + Ant Design 5 + Zustand + React Router 6 + Axios + Monaco Editor + react-markdown + ECharts

---

## 全局实现顺序

```
A: F-001 → F-002 → F-003 → F-004 → F-005 → F-006 → F-007 → F-008
B: F-010 → F-011 → F-012 → F-013 → F-014
C: F-020 → F-021 → F-022 → F-023 → F-024
D: F-030 → F-031 → F-032 → F-033 → F-034 → F-035 → F-036 → F-037 → F-038 → F-039
```

---

## 阶段 A：工程基础 + 核心页面（8 条）

### F-001 项目脚手架

| 项 | 说明 |
|----|------|
| **依赖** | 无 |
| **实现** | Vite + React + TypeScript 项目初始化；目录结构；ESLint + Prettier；路径别名 `@/` |
| **文件** | `package.json`、`vite.config.ts`、`tsconfig.json`、`src/main.tsx`、`src/App.tsx` |
| **验收** | `npm run dev` 启动；`npm run build` 通过；空白页面显示 |

**目录结构**：
```
src/
├── api/              # Axios 实例 + 各模块 API 函数
│   ├── client.ts     # Axios 实例（baseURL、拦截器）
│   ├── auth.ts       # 登录
│   ├── environments.ts
│   ├── policies.ts
│   ├── audit.ts
│   ├── chat.ts
│   ├── plugins.ts
│   ├── knowledge.ts
│   ├── dataQuery.ts
│   ├── inspections.ts
│   ├── riskAlerts.ts
│   ├── runbooks.ts
│   ├── executor.ts
│   └── helpdesk.ts
├── stores/           # Zustand stores
│   ├── authStore.ts
│   └── chatStore.ts
├── components/       # 全局组件
│   ├── Layout/
│   │   ├── AppLayout.tsx
│   │   ├── Header.tsx
│   │   └── Sidebar.tsx
│   ├── AuthGuard.tsx
│   ├── RoleGuard.tsx
│   ├── EnvironmentSelector.tsx
│   ├── ActionPlanCard.tsx
│   ├── StreamingMessage.tsx
│   └── ErrorBoundary.tsx
├── pages/            # 页面组件
│   ├── Login/
│   ├── Dashboard/
│   ├── Chat/
│   ├── Environments/
│   ├── Policies/
│   ├── Audit/
│   ├── Plugins/
│   ├── Knowledge/
│   ├── DataQuery/
│   ├── Inspections/
│   ├── RiskAlerts/
│   ├── Runbooks/
│   ├── Actions/
│   └── HelpDesk/
├── hooks/            # 自定义 hooks
│   ├── useSSE.ts
│   └── usePagination.ts
├── types/            # TypeScript 类型定义
│   └── index.ts
├── utils/            # 工具函数
│   └── jwt.ts
└── router.tsx        # 路由配置
```

### F-002 路由 + 布局框架

| 项 | 说明 |
|----|------|
| **依赖** | F-001 |
| **实现** | React Router 6 路由配置；`AppLayout`（Header + Sidebar + Content 三栏布局）；面包屑导航；404 页面 |
| **API** | 无 |
| **验收** | 侧边栏菜单展开/收起；路由切换；菜单高亮当前页；刷新保持当前路由 |

**菜单结构**：
```
📊 Dashboard
💬 AI 对话
🌐 环境管理
🛡️ 安全策略
📋 审计日志
🔌 插件列表
📚 知识库
📈 数据查询
🔍 巡检任务
⚠️ 高危预警
📖 自愈手册
⏳ 待确认操作
🎧 HelpDesk
```

### F-003 认证（登录 + JWT + 路由守卫）

| 项 | 说明 |
|----|------|
| **依赖** | F-002 |
| **实现** | 登录页 UI；`POST /api/v1/auth/login`；token 存储 localStorage；`AuthGuard` 路由守卫；Axios 401 拦截器自动跳转登录页；JWT 过期检测；退出登录 |
| **API** | `POST /api/v1/auth/login` |
| **验收** | 未登录访问任意页 → 跳转 login；登录成功 → 跳转 dashboard；401 → 自动登出；token 过期提前提示 |

### F-004 环境管理页

| 项 | 说明 |
|----|------|
| **依赖** | F-003 |
| **实现** | 环境列表表格（名称/slug/描述/时间）；新建 Modal；编辑 Modal；slug 前端校验（小写+连字符）；删除确认 |
| **API** | `GET /api/v1/environments`、`POST /api/v1/environments` |
| **验收** | 列表加载；新建/编辑；slug 不合法提示 |

### F-005 安全策略页（CRUD + YAML 编辑器）

| 项 | 说明 |
|----|------|
| **依赖** | F-003 |
| **实现** | 策略列表表格；启用/禁用开关（Switch 组件）；Monaco Editor 集成（YAML 语法高亮+校验）；策略新建/编辑表单；defaultDecision 下拉选择；导入模板按钮；删除确认 |
| **API** | `GET/POST/PUT /api/v1/security-boundaries`、`enable/disable`、`import-template` |
| **验收** | Monaco 编辑器渲染；YAML 保存后回显一致；启用/禁用开关即时生效（乐观更新） |

### F-006 策略模拟器

| 项 | 说明 |
|----|------|
| **依赖** | F-005 |
| **实现** | 模拟页面（`/policies/:id/simulate`）；左侧 ActionPlan JSON 编辑器（Monaco JSON 模式）；右侧结果卡片（decision + matched_rule_id + message + 颜色标记：ALLOW 绿、DENY 红、ASK 橙）；"运行模拟" 按钮 |
| **API** | `POST /api/v1/security-boundaries/:id/simulate` |
| **验收** | 输入 ActionPlan → 点击模拟 → 右侧显示评估结果带颜色标记 |

### F-007 AI 对话页（核心页面）

| 项 | 说明 |
|----|------|
| **依赖** | F-003 |
| **实现** | 三栏布局；左侧历史会话列表（`GET /api/v1/chat/sessions`）；中间对话区（Markdown 渲染、代码块 SyntaxHighlight）；底部输入框 + 环境选择器（`EnvironmentSelector`）；SSE 流式渲染（打字机效果）；消息 ActionPlan 卡片解析；**确认执行**按钮（ASK 决策 → `POST /api/v1/actions/confirm`）；新建会话 |
| **API** | `POST /api/v1/chat`、`GET /api/v1/chat/stream`、`GET /api/v1/chat/sessions`、`GET /api/v1/chat/sessions/:id/messages`、`POST /api/v1/actions/confirm` |
| **验收** | 输入 "echo hello" → SSE 逐字渲染；ActionPlan 卡片显示风险级别颜色；ASK 决策可确认/拒绝；切换历史会话加载历史消息 |

**消息气泡样式**：
- 用户消息：右对齐，蓝色背景
- AI 文本：左对齐，白色背景，Markdown 渲染
- AI ActionPlan：左边，橙色边框卡片，可操作的确认组
- AI 代码块：Copy 按钮
- 加载中：三个跳动点动画

### F-008 概览仪表盘

| 项 | 说明 |
|----|------|
| **依赖** | F-003 |
| **实现** | 4 个 Statistic 卡片（环境数、策略数、待确认数、今日审计数）；最近操作时间线（最近 20 条审计）；插件在线状态表格 |
| **API** | `GET /api/v1/environments`、`GET /api/v1/security-boundaries`、`GET /api/v1/actions/pending`、`GET /api/v1/audit`、`GET /api/v1/plugins` |
| **验收** | 5 个 API 并发请求；卡片数据正确；时间线可滚动 |

---

## 阶段 B：数据管理 + 插件（5 条）

### F-010 审计日志页

| 项 | 说明 |
|----|------|
| **依赖** | F-003 |
| **实现** | 表格（时间/用户/操作/环境/资源/结果）+ 环境下拉筛选 + 关键词搜索 + 分页 |
| **API** | `GET /api/v1/audit?environment=xxx` |
| **验收** | 列表分页正常；环境筛选生效；搜索防抖 300ms |

### F-011 插件列表页

| 项 | 说明 |
|----|------|
| **依赖** | F-003 |
| **实现** | 卡片网格布局（插件名/类型/描述/状态徽章）；每个插件独立健康检查（绿点/红点） |
| **API** | `GET /api/v1/plugins` |
| **验收** | 卡片展示；在线状态正确 |

### F-012 知识库管理页

| 项 | 说明 |
|----|------|
| **依赖** | F-003 |
| **实现** | 文档列表表格；新建/编辑 Markdown 文档（title + content，使用 react-markdown 预览）；删除确认 |
| **API** | `GET/POST/DELETE /api/v1/kb/documents` |
| **验收** | CRUD 完整；Markdown 编辑实时预览 |

### F-013 数据查询页

| 项 | 说明 |
|----|------|
| **依赖** | F-003 |
| **实现** | 插件下拉选择（Kubernetes/Prometheus/Logs）；环境选择；查询输入框 + 发送按钮；结果区：K8s → Ant Table；Prometheus → ECharts 折线图；Logs → 日志行列表 + 关键词高亮；超限截断提示 |
| **API** | `POST /api/v1/data/query` |
| **验收** | 三种数据源结果渲染正确；超限提示显示 |

### F-014 待确认操作页

| 项 | 说明 |
|----|------|
| **依赖** | F-003 |
| **实现** | 表格（时间/用户/操作摘要/命令预览/环境）；确认/拒绝按钮；高危命令红色标记不可操作；10s 轮询自动刷新；**权限：readonly 隐藏操作按钮** |
| **API** | `GET /api/v1/actions/pending`、`POST /api/v1/actions/confirm` |
| **验收** | 列表轮询刷新；确认后条目消失；高危命令红色禁止 |

---

## 阶段 C：巡检 + 自愈 + 预警（5 条）

### F-020 巡检任务管理页

| 项 | 说明 |
|----|------|
| **依赖** | F-003 |
| **实现** | 巡检列表表格（名称/环境/步骤数/最近执行/状态）；新建巡检表单（名称、环境、步骤 Editor：插件+action+参数 JSON）；删除确认 |
| **API** | `GET/POST /api/v1/inspections` |
| **验收** | 列表加载；新建多步骤；步骤可添加/删除/编辑 |

### F-021 巡检执行 + 报告

| 项 | 说明 |
|----|------|
| **依赖** | F-020 |
| **实现** | [立即执行] 按钮 → `POST /api/v1/inspections/:id/run`；执行中 loading 动画；完成后跳转报告列表；报告表格（步骤/状态/耗时/错误信息） |
| **API** | `POST /api/v1/inspections/:id/run`、`GET /api/v1/inspections/:id/reports` |
| **验收** | 执行后状态更新；报告列表展示；失败步骤红色高亮 |

### F-022 高危预警页

| 项 | 说明 |
|----|------|
| **依赖** | F-003 |
| **实现** | 预警卡片列表（级别颜色标记：Critical 红/High 橙/Medium 黄/Low 蓝）；[立即扫描] 按钮；扫描中 loading |
| **API** | `GET /api/v1/risk-alerts`、`POST /api/v1/risk-alerts/scan` |
| **验收** | 卡片颜色正确；扫描触发后列表更新 |

### F-023 Runbook 自愈手册页

| 项 | 说明 |
|----|------|
| **依赖** | F-003 |
| **实现** | 列表表格（名称/关联策略/步骤数/最近执行/成功率）；新建 Runbook 表单；详情页：查看步骤 + [手动执行] 按钮；执行记录子表格 |
| **API** | `GET/POST /api/v1/runbooks`、`POST /api/v1/runbooks/:id/execute` |
| **验收** | CRUD 完整；执行后结果显示 |

### F-024 巡检/自愈与 Chat 联动（React 版，仅标记 todo）

| 项 | 说明 |
|----|------|
| **依赖** | F-007、F-021、F-023 |
| **实现** | Chat 页面右侧面板新增 Tab："巡检"（显示最近巡检结果）+ "自愈"（显示相关 Runbook）；不使用 WebSocket，通过现有 API 查询 |
| **API** | `GET /api/v1/inspections`（最近一次）+ `GET /api/v1/inspections/:id/reports` + `GET /api/v1/runbooks` |
| **验收** | Chat 右侧面板可查看巡检和 Runbook 信息 |

---

## 阶段 D：HelpDesk + IM + 辅助 + RCA + 打磨（10 条）

### F-031 HelpDesk 页面

| 项 | 说明 |
|----|------|
| **依赖** | F-003 |
| **实现** | 简化聊天界面（无环境选择，固定 helpdesk 上下文）；输入问题 → `POST /api/v1/helpdesk/chat`；Markdown 渲染回复 |
| **API** | `POST /api/v1/helpdesk/chat` |
| **验收** | 提问 → KB+RAG 回答；回复含参考文档链接 |

### F-032 Chat 修复辅助面板

| 项 | 说明 |
|----|------|
| **依赖** | F-007 |
| **实现** | Chat 右侧面板新增 "修复辅助" Tab；输入错误信息 + 上下文 → `POST /api/v1/chat/assist`；展示修复建议（Markdown） |
| **API** | `POST /api/v1/chat/assist` |
| **验收** | 输入 "pod crash" → 展示修复步骤 |

### F-033 Chat RCA 根因分析面板

| 项 | 说明 |
|----|------|
| **依赖** | F-007 |
| **实现** | Chat 右侧面板新增 "RCA" Tab；输入事件描述 + 上下文 → `POST /api/v1/chat/rca`；展示分析结果（时间线/因果链 Markdown） |
| **API** | `POST /api/v1/chat/rca` |
| **验收** | 输入 "高CPU告警" → 展示分析链路 |

### F-034 IM Webhook 配置页

| 项 | 说明 |
|----|------|
| **依赖** | F-003 |
| **实现** | Webhook URL 展示（`https://aiops.tclpv.com/api/v1/im/webhook`）+ 复制按钮；使用说明文档；测试发送表单（输入消息文本 → POST webhook） |
| **API** | `POST /api/v1/im/webhook` |
| **验收** | URL 复制；测试发送成功 |

### F-035 全局错误处理 + Loading 态

| 项 | 说明 |
|----|------|
| **依赖** | 所有已实现页面 |
| **实现** | `ErrorBoundary` 全局错误捕获 + 友好提示页；统一 Loading 骨架屏（表格 → Skeleton Table；卡片 → Skeleton Card）；网络请求 loading 状态统一管理 |
| **验收** | 各页面加载有骨架屏；JS 运行时错误被 ErrorBoundary 捕获 |

### F-036 响应式适配

| 项 | 说明 |
|----|------|
| **依赖** | 所有已实现页面 |
| **实现** | Ant Design Grid 响应式断点：xxl(1600) / xl(1200) / lg(992)；Sidebar 在 ≤lg 时折叠为图标模式；Chat 三栏在 ≤xl 时变为两栏（隐藏右侧详情面板）；表格横向滚动 |
| **验收** | 1920/1440/1366/1024 四个断点布局正常 |

### F-037 Dockerfile + Nginx 配置

| 项 | 说明 |
|----|------|
| **依赖** | F-036 |
| **实现** | `Dockerfile`：多阶段构建（node:20 构建 → nginx:alpine 运行）；`nginx.conf`：SPA history mode fallback、gzip、静态资源缓存策略 |
| **验收** | `docker build` 成功；容器本地运行页面可访问 |

**Dockerfile**：
```dockerfile
FROM node:20-alpine AS builder
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY nginx.conf /etc/nginx/nginx.conf
COPY --from=builder /app/dist /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

### F-038 Helm 部署配置

| 项 | 说明 |
|----|------|
| **依赖** | F-037 |
| **实现** | 在现有 `helm/` 中新增 `aiops-frontend` Deployment + Service + 更新 Ingress 路由；`values.yaml` 新增 frontend 配置段 |
| **验收** | `helm template` 生成正确；Ingress 路由 `/` → frontend，`/api/*` → gateway |

### F-039 类型定义 + API 函数完整性检查

| 项 | 说明 |
|----|------|
| **依赖** | 所有已实现页面 |
| **实现** | `src/types/index.ts` 全量 TypeScript 类型（与后端结构体对齐）；`src/api/*.ts` 所有 API 函数覆盖；ESLint 零 error；`tsc --noEmit` 零 error |
| **验收** | `npm run lint` 通过；`npm run typecheck` 通过 |

---

## 联调阶段 TODO

- [ ] `npm run build` 通过，产物 < 500KB gzip
- [ ] Docker image 构建成功
- [ ] Helm `aiops-frontend` 部署
- [ ] Ingress 配置：`/` → frontend，`/api/*` → gateway
- [ ] 登录 → 各页面 → Chat 流式 → 策略编辑器 → 巡检 → 自愈 全链路验收
- [ ] 权限矩阵验收（readonly/operator/admin）

## 产品 DoD

- [ ] 28 条 REQ `done`
- [ ] `npm run build` 零 error
- [ ] Docker + Helm + K8s 部署运行
- [ ] 白屏 / 控制台无 JS 报错
- [ ] 14 个页面全部功能可用

## 变更记录

| 日期 | 变更 |
|------|------|
| 2026-05-27 | 初始创建：28 条 REQ 分 4 个阶段 |
