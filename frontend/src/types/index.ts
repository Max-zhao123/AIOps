// ============ 用户 & 认证 ============
export interface User {
  id: number
  username: string
  role: 'admin' | 'operator' | 'readonly' | 'auditor'
  password_expired?: boolean
}

export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  token: string
  user: User
}

export interface ChangePasswordRequest {
  oldPassword: string
  newPassword: string
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
  status: 'pending' | 'running' | 'succeeded' | 'failed' | 'rejected' | 'rolled_back' | 'rolling_back'
  resultJson?: string
  environment: string
  createdAt: string
  approvers?: Approver[]
  approval_timeout_at?: string
  rollback_available?: boolean
  rollback_from?: number
}

export interface ConfirmRequest {
  executionId: number
  approved: boolean
}

export interface RollbackRequest {
  reason?: string
}

export interface RollbackResponse {
  rollback_execution_id: number
  status: string
}

// ============ 审批 ============
export interface Approver {
  user_id: string
  username: string
  role: string
  approved_at?: string
  status: 'pending' | 'approved' | 'rejected'
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

export interface ComplianceReport {
  period: string
  total_events: number
  by_action: Record<string, number>
  by_result: Record<string, number>
  by_environment: Record<string, number>
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
  status?: 'active' | 'archived'
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

// ============ 定时调度 ============
export interface Schedule {
  id: number
  name: string
  cron: string
  timezone: string
  target_type: string
  target_id: number
  enabled: boolean
  last_run_at?: string
  next_run_at?: string
  created_at?: string
  updated_at?: string
}

export interface ScheduleExecution {
  id: number
  schedule_id: number
  status: 'success' | 'failed' | 'running'
  started_at: string
  completed_at?: string
  result_summary?: string
  error_message?: string
}

export interface CreateScheduleRequest {
  name: string
  cron: string
  timezone?: string
  target_type: string
  target_id: number
  enabled?: boolean
}

// ============ 通知通道 ============
export interface NotificationChannel {
  id: number
  name: string
  type: 'email' | 'webhook' | 'sms'
  config: Record<string, unknown>
  enabled: boolean
  created_at?: string
  updated_at?: string
}

export interface NotificationPolicy {
  id: number
  name: string
  channel_ids: string
  severity: string
  silence_window_min: number
  suppress_lower_severity: boolean
  enabled: boolean
  created_at?: string
}

export interface NotificationTemplate {
  id: number
  name: string
  channel_type: 'email' | 'webhook' | 'sms'
  subject?: string
  body: string
  created_at?: string
}

export interface TestNotificationRequest {
  recipient?: string
  message?: string
}

// ============ 凭证管理 ============
export interface Credential {
  id: number
  name: string
  type: 'smtp' | 'api_key' | 'database' | 'kubeconfig' | 'other'
  value: string
  expires_at?: string
  last_rotated_at?: string
  created_at?: string
  updated_at?: string
}

export interface CreateCredentialRequest {
  name: string
  type: 'smtp' | 'api_key' | 'database' | 'kubeconfig' | 'other'
  value: string
  expires_at?: string
}

export interface RotateCredentialRequest {
  new_value: string
  expires_at?: string
}

// ============ SLA ============
export interface SLADefinition {
  id: number
  name: string
  severity: string
  response_time_min: number
  resolution_time_min: number
  escalation_channel_id: number
  created_at?: string
  updated_at?: string
}

export interface SLASeverityStats {
  total: number
  response_sla_met: number
  resolution_sla_met: number
  avg_mttr_min: number
}

export interface SLAStats {
  period: string
  environment: string
  severity_stats: Record<string, SLASeverityStats>
  overall_mttr_min: number
  overall_sla_rate: number
}

export interface CreateSLARequest {
  name: string
  severity: string
  response_time_min: number
  resolution_time_min: number
  escalation_channel_id?: number
}

// ============ 环境配置 ============
export interface EnvironmentConfig {
  environment_slug: string
  key: string
  value: string
  override_type: string
  created_at?: string
  updated_at?: string
}

export interface UpsertEnvironmentConfigRequest {
  key: string
  value: string
  override_type?: string
}

// ============ 环境配额 ============
export interface QuotaItem {
  daily_limit: number
  current_count: number
  reset_at: string
}

export interface EnvironmentQuota {
  environment: string
  quotas: Record<string, QuotaItem>
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

// ============ 通用 ============
export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data: T
}
