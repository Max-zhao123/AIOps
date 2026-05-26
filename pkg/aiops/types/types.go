package types

// ActionPlan 运维动作计划（PRD §3.5）。
type ActionPlan struct {
	Plugin          string                 `json:"plugin"`
	Action          string                 `json:"action"`
	Parameters      map[string]interface{} `json:"parameters"`
	Risk            string                 `json:"risk"`
	Summary         string                 `json:"summary"`
	CommandPreview  string                 `json:"commandPreview"`
}

// PolicySpec 安全边界（PRD §3.6）。
type PolicySpec struct {
	DefaultDecision string `json:"defaultDecision" yaml:"defaultDecision"`
	Rules           []Rule `json:"rules" yaml:"rules"`
}

type Rule struct {
	ID       string    `json:"id" yaml:"id"`
	Match    RuleMatch `json:"match" yaml:"match"`
	Decision string    `json:"decision" yaml:"decision"`
	Message  string    `json:"message" yaml:"message"`
}

type RuleMatch struct {
	Plugin           string            `json:"plugin" yaml:"plugin"`
	Actions          []string          `json:"actions" yaml:"actions"`
	CommandPattern   string            `json:"commandPattern" yaml:"commandPattern"`
	ResourceLabels   map[string]string `json:"resourceLabels" yaml:"resourceLabels"`
}

// EvaluateRequest policy 内部请求。
type EvaluateRequest struct {
	Environment string     `json:"environment"`
	Plan        ActionPlan `json:"plan"`
}

// EvaluateResponse policy 评估结果。
type EvaluateResponse struct {
	Decision       string   `json:"decision"`
	Message        string   `json:"message"`
	MatchedRuleIDs []string `json:"matchedRuleIds"`
	HardDeny       bool     `json:"hardDeny"`
}

// RunRequest executor 内部请求。
type RunRequest struct {
	Environment string     `json:"environment"`
	SessionID   int64      `json:"sessionId,omitempty"`
	Plan        ActionPlan `json:"plan"`
}

// ExecuteRequest 插件内部请求。
type ExecuteRequest struct {
	Plan ActionPlan `json:"plan"`
}

// ExecuteResponse 插件执行结果。
type ExecuteResponse struct {
	Status  string      `json:"status"`
	Output  interface{} `json:"output,omitempty"`
	Message string      `json:"message,omitempty"`
}

// AuditRequest 审计写入。
type AuditRequest struct {
	EventType   string      `json:"eventType"`
	Environment string      `json:"environment"`
	UserID      int64       `json:"userId"`
	Payload     interface{} `json:"payload"`
}

const (
	DecisionAllow = "ALLOW"
	DecisionAsk   = "ASK"
	DecisionDeny  = "DENY"
)

const (
	RoleAdmin    = "admin"
	RoleOperator = "operator"
	RoleReadonly = "readonly"
	RoleAuditor  = "auditor"
)
