package model

import basemodel "github.com/lihaiya/aiops/pkg/model"

type Environment struct {
	basemodel.BaseModel
	Slug        string `json:"slug" gorm:"column:slug;type:varchar(64);uniqueIndex;not null"`
	Name        string `json:"name" gorm:"column:name;type:varchar(128);not null"`
	Description string `json:"description" gorm:"column:description;type:varchar(512)"`
}

func (Environment) TableName() string { return "environments" }
