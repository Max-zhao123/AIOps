package model

import (
	"time"

	basemodel "github.com/lihaiya/aiops/pkg/model"
)

// NotificationChannel 通知通道（REQ-091）。
type NotificationChannel struct {
	basemodel.BaseModel
	Name            string `json:"name" gorm:"column:name;type:varchar(100);not null;uniqueIndex:idx_env_name" binding:"required,max=100"`
	Type            string `json:"type" gorm:"column:type;type:varchar(32);not null" binding:"required,oneof=email webhook sms"`
	Config          string `json:"config" gorm:"column:config;type:text;not null"`
	Enabled         bool   `json:"enabled" gorm:"column:enabled;default:true"`
	EnvironmentSlug string `json:"environmentSlug" gorm:"column:environment_slug;type:varchar(50);not null;default:default;uniqueIndex:idx_env_name"`
}

func (NotificationChannel) TableName() string { return "notification_channels" }

// NotificationPolicy 通知策略（REQ-091）。
type NotificationPolicy struct {
	basemodel.BaseModel
	Name                    string `json:"name" gorm:"column:name;type:varchar(100);not null" binding:"required,max=100"`
	Severity                string `json:"severity" gorm:"column:severity;type:varchar(32);not null;uniqueIndex:idx_env_severity" binding:"required,oneof=critical high medium low"`
	ChannelIDs              string `json:"channelIds" gorm:"column:channel_ids;type:text;not null"`
	SilenceWindowMin        int    `json:"silenceWindowMin" gorm:"column:silence_window_min;not null;default:30"`
	SuppressLowerSeverity   bool   `json:"suppressLowerSeverity" gorm:"column:suppress_lower_severity;default:false"`
	EnvironmentSlug         string `json:"environmentSlug" gorm:"column:environment_slug;type:varchar(50);not null;default:default;uniqueIndex:idx_env_severity"`
}

func (NotificationPolicy) TableName() string { return "notification_policies" }

// NotificationTemplate 通知模板（REQ-091）。
type NotificationTemplate struct {
	basemodel.BaseModel
	ChannelType      string `json:"channelType" gorm:"column:channel_type;type:varchar(32);not null;uniqueIndex:idx_type_event" binding:"required,oneof=email webhook sms"`
	EventType        string `json:"eventType" gorm:"column:event_type;type:varchar(50);not null;uniqueIndex:idx_type_event" binding:"required,max=50"`
	Subject          string `json:"subject" gorm:"column:subject;type:varchar(200)"`
	Body             string `json:"body" gorm:"column:body;type:text;not null"`
	EnvironmentSlug  string `json:"environmentSlug" gorm:"column:environment_slug;type:varchar(50);not null;default:default;uniqueIndex:idx_type_event"`
}

func (NotificationTemplate) TableName() string { return "notification_templates" }

// NotifySilence 通知静默记录（内存缓存，不持久化到 DB，此处仅作逻辑参考）。
// 实际使用 sync.Map 在 notifier 包中实现。

// NotifyRecordExt 扩展通知记录（REQ-091），扩展 existing NotifyRecord。
type NotifyRecordExt struct {
	ID              int64       `json:"id"`
	Source          string      `json:"source"`
	Environment     string      `json:"environment"`
	Level           string      `json:"level"`
	Message         string      `json:"message"`
	ChannelType     string      `json:"channelType"`
	ChannelID       int64       `json:"channelId"`
	Status          string      `json:"status"`
	RetryCount      int         `json:"retryCount"`
	AlertKey        string      `json:"alertKey"`
	NextRetryAt     *time.Time  `json:"nextRetryAt"`
	CreatedAt       time.Time   `json:"createdAt"`
}

// ExpiredCredentialNotification 用于凭证过期提醒通知。
type ExpiredCredentialNotification struct {
	CredentialID   int64     `json:"credentialId"`
	CredentialName string    `json:"credentialName"`
	ExpiresAt      time.Time `json:"expiresAt"`
	DaysRemaining  int       `json:"daysRemaining"`
}
