package model

import (
	"time"

	basemodel "github.com/lihaiya/aiops/pkg/model"
)

// SlaDefinition SLA 定义（REQ-096）。
type SlaDefinition struct {
	basemodel.BaseModel
	Name                string `json:"name" gorm:"column:name;type:varchar(100);not null;uniqueIndex:idx_env_severity" binding:"required,max=100"`
	Severity            string `json:"severity" gorm:"column:severity;type:varchar(32);not null;uniqueIndex:idx_env_severity" binding:"required,oneof=critical high medium low"`
	ResponseTimeMin     int    `json:"responseTimeMin" gorm:"column:response_time_min;not null" binding:"required,min=1"`
	ResolutionTimeMin   int    `json:"resolutionTimeMin" gorm:"column:resolution_time_min;not null" binding:"required,min=1"`
	EscalationChannelID *int64 `json:"escalationChannelId" gorm:"column:escalation_channel_id"`
	EnvironmentSlug     string `json:"environmentSlug" gorm:"column:environment_slug;type:varchar(50);not null;default:default;uniqueIndex:idx_env_severity"`
}

func (SlaDefinition) TableName() string { return "sla_definitions" }

// SlaRecord SLA 记录（REQ-096）。
type SlaRecord struct {
	basemodel.BaseModel
	AlertID          int64      `json:"alertId" gorm:"column:alert_id;not null;index:idx_alert"`
	SlaDefinitionID  int64      `json:"slaDefinitionId" gorm:"column:sla_definition_id;not null;index:idx_sla_def"`
	CreatedAt        time.Time  `json:"createdAt" gorm:"column:created_at;not null;index:idx_env_time"`
	AcknowledgedAt   *time.Time `json:"acknowledgedAt" gorm:"column:acknowledged_at"`
	ResolvedAt       *time.Time `json:"resolvedAt" gorm:"column:resolved_at"`
	ResponseSlaMet   *bool      `json:"responseSlaMet" gorm:"column:response_sla_met"`
	ResolutionSlaMet *bool      `json:"resolutionSlaMet" gorm:"column:resolution_sla_met"`
	Escalated        bool       `json:"escalated" gorm:"column:escalated;default:false"`
	EnvironmentSlug  string     `json:"environmentSlug" gorm:"column:environment_slug;type:varchar(50);not null;default:default;index:idx_env_time"`
}

func (SlaRecord) TableName() string { return "sla_records" }

// SlaStatsResponse SLA 统计 API 响应（REQ-096）。
type SlaStatsResponse struct {
	Period        string                       `json:"period"`
	Environment   string                       `json:"environment"`
	SeverityStats map[string]*SlaSeverityStats `json:"severityStats"`
	OverallMTTR   float64                      `json:"overallMttrMin"`
	OverallSLARate float64                     `json:"overallSlaRate"`
}

// SlaSeverityStats 单个 severity 的 SLA 统计。
type SlaSeverityStats struct {
	Total            int     `json:"total"`
	ResponseSlaMet   int     `json:"responseSlaMet"`
	ResolutionSlaMet int     `json:"resolutionSlaMet"`
	AvgMTTRMin       float64 `json:"avgMttrMin"`
}
