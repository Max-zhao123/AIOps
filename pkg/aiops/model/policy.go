package model

import basemodel "github.com/lihaiya/aiops/pkg/model"

type SecurityBoundaryPolicy struct {
	basemodel.BaseModel
	EnvironmentID int64  `json:"environmentId" gorm:"column:environment_id;index;not null"`
	Name          string `json:"name" gorm:"column:name;type:varchar(128);not null"`
	SpecYAML      string `json:"specYaml" gorm:"column:spec_yaml;type:longtext;not null"`
	Enabled       bool   `json:"enabled" gorm:"column:enabled;default:false"`
	// REQ-094 审批配置
	ApprovalConfig string `json:"approvalConfig" gorm:"column:approval_config;type:text"`
	// REQ-095 回滚计划
	RollbackPlans  string `json:"rollbackPlans" gorm:"column:rollback_plans;type:text"`
	RollbackPolicy string `json:"rollbackPolicy" gorm:"column:rollback_policy;type:varchar(32);default:manual"`
}

func (SecurityBoundaryPolicy) TableName() string { return "security_boundary_policies" }
