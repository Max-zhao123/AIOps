package model

import basemodel "github.com/lihaiya/aiops/pkg/model"

type ExecutionRecord struct {
	basemodel.BaseModel
	SessionID     int64  `json:"sessionId" gorm:"column:session_id;index"`
	Environment   string `json:"environment" gorm:"column:environment;type:varchar(64);index"`
	UserID        int64  `json:"userId" gorm:"column:user_id;index"`
	Plugin        string `json:"plugin" gorm:"column:plugin;type:varchar(64)"`
	Action        string `json:"action" gorm:"column:action;type:varchar(64)"`
	PlanJSON      string `json:"planJson" gorm:"column:plan_json;type:text"`
	Decision      string `json:"decision" gorm:"column:decision;type:varchar(16)"`
	Status        string `json:"status" gorm:"column:status;type:varchar(32);index"`
	ResultJSON    string `json:"resultJson" gorm:"column:result_json;type:text"`
}

func (ExecutionRecord) TableName() string { return "execution_records" }
