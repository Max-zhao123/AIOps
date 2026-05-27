package model

import (
	basemodel "github.com/lihaiya/aiops/pkg/model"
)

// EnvironmentConfig 环境配置（REQ-097）。
type EnvironmentConfig struct {
	basemodel.BaseModel
	EnvironmentSlug string `json:"environmentSlug" gorm:"column:environment_slug;type:varchar(50);not null;uniqueIndex:idx_env_key" binding:"required,max=50"`
	ConfigKey       string `json:"key" gorm:"column:config_key;type:varchar(200);not null;uniqueIndex:idx_env_key" binding:"required,max=200"`
	ConfigValue     string `json:"value" gorm:"column:config_value;type:text;not null"`
	OverrideType    string `json:"overrideType" gorm:"column:override_type;type:varchar(32);not null;default:other" binding:"oneof=plugin service llm security other"`
}

func (EnvironmentConfig) TableName() string { return "environment_configs" }

// ConfigVersion 配置版本号（REQ-099）。
type ConfigVersion struct {
	basemodel.BaseModel
	Version       int    `json:"version" gorm:"column:version;not null;uniqueIndex:idx_version"`
	ChangedBy     string `json:"changedBy" gorm:"column:changed_by;type:varchar(100)"`
	ChangeSummary string `json:"changeSummary" gorm:"column:change_summary;type:varchar(500)"`
}

func (ConfigVersion) TableName() string { return "config_versions" }
