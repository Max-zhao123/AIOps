package model

import basemodel "github.com/lihaiya/aiops/pkg/model"

type LlmConfig struct {
	basemodel.BaseModel
	Name    string `json:"name" gorm:"column:name;type:varchar(64);not null"`
	BaseURL string `json:"baseUrl" gorm:"column:base_url;type:varchar(512);not null"`
	APIKey  string `json:"apiKey" gorm:"column:api_key;type:varchar(512);not null"`
	Model   string `json:"model" gorm:"column:model;type:varchar(128);not null"`
	Active  bool   `json:"active" gorm:"column:active;default:true"`
}

func (LlmConfig) TableName() string { return "llm_configs" }
