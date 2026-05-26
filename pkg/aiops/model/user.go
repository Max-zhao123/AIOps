package model

import basemodel "github.com/lihaiya/aiops/pkg/model"

type User struct {
	basemodel.BaseModel
	Username               string `json:"username" gorm:"column:username;type:varchar(64);uniqueIndex;not null"`
	Password               string `json:"-" gorm:"column:password;type:varchar(255);not null"`
	Role                   string `json:"role" gorm:"column:role;type:varchar(32);not null;default:operator"`
	AllowedEnvironmentIDs  string `json:"allowedEnvironmentIds" gorm:"column:allowed_environment_ids;type:text"`
}

func (User) TableName() string { return "users" }
