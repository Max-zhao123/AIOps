package model

import basemodel "github.com/lihaiya/aiops/pkg/model"

type AuditLog struct {
	basemodel.BaseModel
	EventType     string `json:"eventType" gorm:"column:event_type;type:varchar(64);index;not null"`
	Environment   string `json:"environment" gorm:"column:environment;type:varchar(64);index"`
	UserID        int64  `json:"userId" gorm:"column:user_id;index"`
	Payload       string `json:"payload" gorm:"column:payload;type:text"`
}

func (AuditLog) TableName() string { return "audit_logs" }
