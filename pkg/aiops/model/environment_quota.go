package model

import (
	"time"

	basemodel "github.com/lihaiya/aiops/pkg/model"
)

// EnvironmentQuota 环境配额（REQ-103）。
type EnvironmentQuota struct {
	basemodel.BaseModel
	EnvironmentSlug string    `json:"environmentSlug" gorm:"column:environment_slug;type:varchar(50);not null;uniqueIndex:idx_env_resource" binding:"required,max=50"`
	ResourceType    string    `json:"resourceType" gorm:"column:resource_type;type:varchar(32);not null;uniqueIndex:idx_env_resource" binding:"required,oneof=session execution inspection"`
	DailyLimit      int       `json:"dailyLimit" gorm:"column:daily_limit;not null;default:1000"`
	CurrentCount    int       `json:"currentCount" gorm:"column:current_count;not null;default:0"`
	ResetAt         time.Time `json:"resetAt" gorm:"column:reset_at;type:date;not null"`
}

func (EnvironmentQuota) TableName() string { return "environment_quotas" }

// QuotaCheckResult 配额检查结果。
type QuotaCheckResult struct {
	Allowed    bool   `json:"allowed"`
	Reason     string `json:"reason,omitempty"`
	CurrentCount int  `json:"currentCount"`
	DailyLimit   int  `json:"dailyLimit"`
}
