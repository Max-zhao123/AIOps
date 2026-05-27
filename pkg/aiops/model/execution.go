package model

import basemodel "github.com/lihaiya/aiops/pkg/model"

type ExecutionRecord struct {
	basemodel.BaseModel
	SessionID          int64  `json:"sessionId" gorm:"column:session_id;index"`
	Environment       string `json:"environment" gorm:"column:environment;type:varchar(64);index"`
	UserID            int64  `json:"userId" gorm:"column:user_id;index"`
	Plugin            string `json:"plugin" gorm:"column:plugin;type:varchar(64)"`
	Action            string `json:"action" gorm:"column:action;type:varchar(64)"`
	PlanJSON          string `json:"planJson" gorm:"column:plan_json;type:text"`
	Decision          string `json:"decision" gorm:"column:decision;type:varchar(16)"`
	Status            string `json:"status" gorm:"column:status;type:varchar(32);index"`
	ResultJSON        string `json:"resultJson" gorm:"column:result_json;type:text"`
	// REQ-092 执行超时与重试
	Timeout     bool   `json:"timeout" gorm:"column:timeout;default:false"`
	RetryCount  int    `json:"retryCount" gorm:"column:retry_count;default:0"`
	RetryHistory string `json:"retryHistory" gorm:"column:retry_history;type:text"`
	Progress    string `json:"progress" gorm:"column:progress;type:text"`
	// REQ-094 变更审批流
	Approvers         string `json:"approvers" gorm:"column:approvers;type:text"`
	ApprovalStatus    string `json:"approvalStatus" gorm:"column:approval_status;type:varchar(32);index:idx_exec_approval"`
	ApprovalTimeoutAt string `json:"approvalTimeoutAt" gorm:"column:approval_timeout_at;type:varchar(32);index:idx_exec_approval"`
	// REQ-095 操作回滚
	PreSnapshot       string `json:"preSnapshot" gorm:"column:pre_snapshot;type:text"`
	ParentExecutionID *int64 `json:"parentExecutionId" gorm:"column:parent_execution_id"`
	RollbackPolicy    string `json:"rollbackPolicy" gorm:"column:rollback_policy;type:varchar(32)"`
	// REQ-103 多租户数据隔离
	EnvironmentSlug string `json:"environmentSlug" gorm:"column:environment_slug;type:varchar(50);not null;default:default;index:idx_exec_env"`
}

func (ExecutionRecord) TableName() string { return "execution_records" }
