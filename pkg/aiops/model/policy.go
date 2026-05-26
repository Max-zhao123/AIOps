package model

import basemodel "github.com/lihaiya/aiops/pkg/model"

type SecurityBoundaryPolicy struct {
	basemodel.BaseModel
	EnvironmentID int64  `json:"environmentId" gorm:"column:environment_id;index;not null"`
	Name          string `json:"name" gorm:"column:name;type:varchar(128);not null"`
	SpecYAML      string `json:"specYaml" gorm:"column:spec_yaml;type:longtext;not null"`
	Enabled       bool   `json:"enabled" gorm:"column:enabled;default:false"`
}

func (SecurityBoundaryPolicy) TableName() string { return "security_boundary_policies" }
