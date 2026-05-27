package model

import (
	basemodel "github.com/lihaiya/aiops/pkg/model"
)

// ApprovalPolicy 审批策略（REQ-094）。
type ApprovalPolicy struct {
	basemodel.BaseModel
	RiskLevel           string `json:"riskLevel" gorm:"column:risk_level;type:varchar(32);not null;uniqueIndex:idx_env_risk" binding:"required,oneof=write notify"`
	MinApprovers        int    `json:"minApprovers" gorm:"column:min_approvers;not null;default:1"`
	TimeoutMin          int    `json:"timeoutMin" gorm:"column:timeout_min;not null;default:30"`
	TimeoutAction       string `json:"timeoutAction" gorm:"column:timeout_action;type:varchar(32);not null;default:REJECT" binding:"oneof=REJECT ESCALATE"`
	RequireDifferentRoles bool  `json:"requireDifferentRoles" gorm:"column:require_different_roles;default:true"`
	EnvironmentSlug     string `json:"environmentSlug" gorm:"column:environment_slug;type:varchar(50);not null;default:default;uniqueIndex:idx_env_risk"`
}

func (ApprovalPolicy) TableName() string { return "approval_policies" }

// ApproverEntry 审批人记录（JSON 内嵌于 ExecutionRecord.Approvers）。
type ApproverEntry struct {
	UserID     int64  `json:"user_id"`
	Role       string `json:"role"`
	ApprovedAt string `json:"approved_at"`
	Reason     string `json:"reason,omitempty"`
}
