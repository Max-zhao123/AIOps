package model

import basemodel "github.com/lihaiya/aiops/pkg/model"

type ChatSession struct {
	basemodel.BaseModel
	UserID      int64  `json:"userId" gorm:"column:user_id;index"`
	Environment string `json:"environment" gorm:"column:environment;type:varchar(64);index"`
	Title       string `json:"title" gorm:"column:title;type:varchar(256)"`
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
