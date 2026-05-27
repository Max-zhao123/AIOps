// ============ 用户 & 认证 ============
export interface User {
  id: number
  username: string
  role: 'admin' | 'operator' | 'readonly' | 'auditor'
}

export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  token: string
  user: User
}

// ============ 环境 ============
export interface Environment {
  id: number
  name: string
  slug: string
  description: string
  created_at: string
}

// ============ 安全策略 ============
export interface SecurityPolicy {
  id: number
  name: string
  spec_yaml: string
  enabled: boolean
  created_at: string
  updated_at: string
}

export interface PolicyRule {
  id: string
  match: {
    plugin: string
    actions: string[]
    commandPattern?: string
    resourceLabels?: Record<string, string>
  }
  decision: 'ALLOW' | 'ASK' | 'DENY'
  message: string
}

export interface PolicySpec {
  defaultDecision: 'ALLOW' | 'ASK' | 'DENY'
  rules: PolicyRule[]
}

export interface PolicyEvaluateResult {
  decision: 'ALLOW' | 'ASK' | 'DENY'
  matched_rule_id?: string
  message: string
}

export interface SimulateRequest {
  plan: ActionPlan
  securitySpecYAML: string
}

// ============ ActionPlan ============
export interface ActionPlan {
  plugin: string
  action: string
  parameters: Record<string, unknown>
  risk?: 'read' | 'write' | 'notify'
  summary?: string
  commandPreview?: string
}

// ============ 执行 ============
export interface ExecutionRecord {
  id: number
  sessionId: number
  userId: number
  plugin: string
  action: string
  planJson: string
  decision: string
  status: 'pending' | 'running' | 'succeeded' | 'failed' | 'rejected'
  resultJson?: string
  environment: string
  createdAt: string
}

export interface ConfirmRequest {
  executionId: number
  approved: boolean
}

// ============ 审计 ============
export interface AuditLog {
  id: number
  user_id: number
  username: string
  action: string
  resource_type: string
  resource_id: string
  detail: string
  environment: string
  result: string
  created_at: string
}

// ============ 插件 ============
export interface Plugin {
  name: string
  type: string
  description: string
  online: boolean
}

// ============ 知识库 ============
export interface KBDocument {
  id: number
  title: string
  content: string
  created_at: string
  updated_at: string
}

// ============ 数据查询 ============
export interface DataQueryRequest {
  plugin: string
  query: string
  environment?: string
  timeRange?: { start: string; end: string }
}

export interface DataQueryResponse {
  plugin: string
  data: unknown[]
  truncated: boolean
  total_rows: number
  message?: string
}

// ============ 巡检 ============
export interface InspectionStep {
  plugin: string
  action: string
  parameters: Record<string, unknown>
}

export interface Inspection {
  id: number
  name: string
  environment: string
  steps: InspectionStep[]
  last_run_at?: string
  status: string
  created_at: string
}

export interface InspectionReport {
  id: number
  inspection_id: number
  step_index: number
  status: 'success' | 'failed'
  result: string
  duration_ms: number
  error?: string
  executed_at: string
}

// ============ 高危预警 ============
export interface RiskAlert {
  id: number
  level: 'critical' | 'high' | 'medium' | 'low'
  title: string
  description: string
  source: string
  created_at: string
}

// ============ Runbook ============
export interface RunbookStep {
  plugin: string
  action: string
  parameters: Record<string, unknown>
}

export interface Runbook {
  id: number
  name: string
  policy_id?: number
  steps: RunbookStep[]
  success_count: number
  fail_count: number
  last_run_at?: string
  created_at: string
}

export interface RunbookExecution {
  id: number
  runbook_id: number
  status: 'success' | 'failed' | 'running'
  result: string
  executed_at: string
}

// ============ Chat ============
export interface ChatRequest {
  environment: string
  message: string
  stream?: boolean
}

export interface ChatSession {
  id: number
  title: string
  environment: string
  created_at: string
  updated_at: string
}

export interface ChatMessage {
  id: number
  session_id: number
  role: 'user' | 'assistant' | 'system'
  content: string
  action_plans?: ActionPlan[]
  created_at: string
}

// ============ HelpDesk ============
export interface HelpDeskRequest {
  query: string
}

// ============ IM ============
export interface IMWebhookRequest {
  text: string
}

// ============ 通用 ============
export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data: T
}

// ============ LLM 配置 ============
export interface LlmConfig {
  id: number
  name: string
  baseUrl: string
  apiKey: string
  model: string
  active: boolean
  createdAt: string
}
