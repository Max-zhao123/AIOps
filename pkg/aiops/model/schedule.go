package model

import (
	"time"

	basemodel "github.com/lihaiya/aiops/pkg/model"
)

// Schedule 定时调度配置（REQ-090）。
type Schedule struct {
	basemodel.BaseModel
	Name            string     `json:"name" gorm:"column:name;type:varchar(100);not null" binding:"required,max=100"`
	Cron            string     `json:"cron" gorm:"column:cron;type:varchar(50);not null" binding:"required,max=50"`
	Timezone        string     `json:"timezone" gorm:"column:timezone;type:varchar(50);not null;default:UTC"`
	TargetType      string     `json:"targetType" gorm:"column:target_type;type:varchar(32);not null" binding:"required,oneof=inspection runbook"`
	TargetID        int64      `json:"targetId" gorm:"column:target_id;not null" binding:"required"`
	Enabled         bool       `json:"enabled" gorm:"column:enabled;default:false"`
	EnvironmentSlug string     `json:"environmentSlug" gorm:"column:environment_slug;type:varchar(50);not null;default:default;index:idx_env_enabled"`
	LastRunAt       *time.Time `json:"lastRunAt" gorm:"column:last_run_at"`
	NextRunAt       *time.Time `json:"nextRunAt" gorm:"column:next_run_at;index:idx_next_run"`
}

func (Schedule) TableName() string { return "schedules" }

// ScheduleExecution 调度执行历史（REQ-090）。
type ScheduleExecution struct {
	basemodel.BaseModel
	ScheduleID    int64      `json:"scheduleId" gorm:"column:schedule_id;not null;index:idx_schedule"`
	Status        string     `json:"status" gorm:"column:status;type:varchar(32);not null;default:running" binding:"oneof=running completed failed skipped timeout"`
	StartedAt     time.Time  `json:"startedAt" gorm:"column:started_at;not null"`
	CompletedAt   *time.Time `json:"completedAt" gorm:"column:completed_at"`
	ResultSummary string     `json:"resultSummary" gorm:"column:result_summary;type:varchar(500)"`
	ErrorMessage  string     `json:"errorMessage" gorm:"column:error_message;type:text"`
}

func (ScheduleExecution) TableName() string { return "schedule_executions" }
