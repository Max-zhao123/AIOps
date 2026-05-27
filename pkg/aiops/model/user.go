package model

import (
	"time"

	basemodel "github.com/lihaiya/aiops/pkg/model"
)

type User struct {
	basemodel.BaseModel
	Username               string `json:"username" gorm:"column:username;type:varchar(64);uniqueIndex;not null"`
	Password               string `json:"-" gorm:"column:password;type:varchar(255);not null"`
	Role                   string `json:"role" gorm:"column:role;type:varchar(32);not null;default:operator"`
	AllowedEnvironmentIDs  string `json:"allowedEnvironmentIds" gorm:"column:allowed_environment_ids;type:text"`
	// REQ-101 密码策略
	PasswordChangedAt  *time.Time `json:"passwordChangedAt" gorm:"column:password_changed_at"`
	FailedLoginCount   int        `json:"failedLoginCount" gorm:"column:failed_login_count;not null;default:0"`
	LockedUntil        *time.Time `json:"lockedUntil" gorm:"column:locked_until;index:idx_user_locked"`
	MustChangePassword bool       `json:"mustChangePassword" gorm:"column:must_change_password;default:false"`
	PasswordHistory    string     `json:"-" gorm:"column:password_history;type:text"`
}

func (User) TableName() string { return "users" }
