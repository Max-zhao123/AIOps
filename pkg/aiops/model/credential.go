package model

import (
	"time"

	basemodel "github.com/lihaiya/aiops/pkg/model"
)

// Credential 凭证管理（REQ-093）。
type Credential struct {
	basemodel.BaseModel
	Name            string     `json:"name" gorm:"column:name;type:varchar(100);not null;uniqueIndex:idx_env_name" binding:"required,max=100"`
	Type            string     `json:"type" gorm:"column:type;type:varchar(32);not null" binding:"required,oneof=smtp api_key database kubeconfig other"`
	EncryptedValue  string     `json:"-" gorm:"column:encrypted_value;type:text;not null"`
	ExpiresAt       *time.Time `json:"expiresAt" gorm:"column:expires_at;index:idx_expires"`
	LastRotatedAt   *time.Time `json:"lastRotatedAt" gorm:"column:last_rotated_at"`
	EnvironmentSlug string     `json:"environmentSlug" gorm:"column:environment_slug;type:varchar(50);not null;default:default;uniqueIndex:idx_env_name"`
}

func (Credential) TableName() string { return "credentials" }

// CredentialResponse 凭证 API 响应（脱敏）。
type CredentialResponse struct {
	ID              int64      `json:"id"`
	Name            string     `json:"name"`
	Type            string     `json:"type"`
	Value           string     `json:"value"`
	ExpiresAt       *time.Time `json:"expiresAt,omitempty"`
	LastRotatedAt   *time.Time `json:"lastRotatedAt,omitempty"`
	CreatedAt       string     `json:"createdAt"`
	UpdatedAt       string     `json:"updatedAt"`
}

// CredentialValueResponse 凭证解密值响应（内部 API）。
type CredentialValueResponse struct {
	ID    int64  `json:"id"`
	Value string `json:"value"`
	Type  string `json:"type"`
}
