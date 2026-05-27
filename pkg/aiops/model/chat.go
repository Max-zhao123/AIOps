package model

import (
	"time"

	basemodel "github.com/lihaiya/aiops/pkg/model"
)

type ChatSession struct {
	basemodel.BaseModel
	UserID      int64  `json:"userId" gorm:"column:user_id;index"`
	Environment string `json:"environment" gorm:"column:environment;type:varchar(64);index"`
	Title       string `json:"title" gorm:"column:title;type:varchar(256)"`
	// REQ-100 会话管理
	TTLDays        int        `json:"ttlDays" gorm:"column:ttl_days;not null;default:90"`
	LastActiveAt   time.Time  `json:"lastActiveAt" gorm:"column:last_active_at;not null;default:CURRENT_TIMESTAMP;index:idx_session_env_status"`
	Status         string     `json:"status" gorm:"column:status;type:varchar(32);not null;default:active;index:idx_session_env_status"`
	EnvironmentSlug string    `json:"environmentSlug" gorm:"column:environment_slug;type:varchar(50);not null;default:default;index:idx_session_env_status"`
}

func (ChatSession) TableName() string { return "chat_sessions" }

type ChatMessage struct {
	basemodel.BaseModel
	SessionID int64  `json:"sessionId" gorm:"column:session_id;index;not null"`
	Role      string `json:"role" gorm:"column:role;type:varchar(16);not null"`
	Content   string `json:"content" gorm:"column:content;type:longtext;not null"`
	PlansJSON string `json:"plansJson" gorm:"column:plans_json;type:text"`
}

func (ChatMessage) TableName() string { return "chat_messages" }
