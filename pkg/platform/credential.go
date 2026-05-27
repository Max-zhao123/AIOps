package platform

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	aimodel "github.com/lihaiya/aiops/pkg/aiops/model"
	"github.com/lihaiya/aiops/pkg/aiops/crypto"
	"github.com/lihaiya/aiops/pkg/aiops/tenant"
	"github.com/lihaiya/aiops/pkg/runtime"
)

// -------- 凭证管理（REQ-093） --------

func (h *Handler) ListCredentials(c *gin.Context) {
	envSlug := tenant.GetEnvironmentSlug(c)
	var list []aimodel.Credential
	q := h.DB.Where("environment_slug = ?", envSlug).Order("id asc")
	if credType := c.Query("type"); credType != "" {
		q = q.Where("type = ?", credType)
	}
	q.Find(&list)

	var resp []aimodel.CredentialResponse
	for _, cred := range list {
		resp = append(resp, aimodel.CredentialResponse{
			ID:            cred.ID,
			Name:          cred.Name,
			Type:          cred.Type,
			Value:         "******",
			ExpiresAt:     cred.ExpiresAt,
			LastRotatedAt: cred.LastRotatedAt,
			CreatedAt:     cred.CreatedAt.Time().Format(time.RFC3339),
			UpdatedAt:     cred.UpdatedAt.Time().Format(time.RFC3339),
		})
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": resp})
}

func (h *Handler) CreateCredential(c *gin.Context) {
	var body struct {
		Name      string     `json:"name" binding:"required,max=100"`
		Type      string     `json:"type" binding:"required,oneof=smtp api_key database kubeconfig other"`
		Value     string     `json:"value" binding:"required,min=1,max=4096"`
		ExpiresAt *time.Time `json:"expiresAt"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error(), "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	envSlug := tenant.FillEnvironmentSlug(c)

	// 幂等检查：同环境下同名凭证
	var cnt int64
	h.DB.Model(&aimodel.Credential{}).Where("name = ? AND environment_slug = ?", body.Name, envSlug).Count(&cnt)
	if cnt > 0 {
		c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "同名凭证已存在", "request_id": runtime.RequestIDFromContext(c)})
		return
	}

	encrypted, err := crypto.Encrypt(body.Value)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "凭证加密失败", "request_id": runtime.RequestIDFromContext(c)})
		return
	}

	cred := aimodel.Credential{
		Name:            body.Name,
		Type:            body.Type,
		EncryptedValue:  encrypted,
		ExpiresAt:       body.ExpiresAt,
		EnvironmentSlug: envSlug,
	}
	if err := h.DB.Create(&cred).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "凭证创建失败", "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": 0, "data": aimodel.CredentialResponse{
		ID:        cred.ID,
		Name:      cred.Name,
		Type:      cred.Type,
		Value:     "******",
		ExpiresAt: cred.ExpiresAt,
		CreatedAt: cred.CreatedAt.Time().Format(time.RFC3339),
		UpdatedAt: cred.UpdatedAt.Time().Format(time.RFC3339),
	}})
}

func (h *Handler) GetCredential(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	envSlug := tenant.GetEnvironmentSlug(c)
	var cred aimodel.Credential
	if err := h.DB.Where("id = ? AND environment_slug = ?", id, envSlug).First(&cred).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "凭证不存在", "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": aimodel.CredentialResponse{
		ID:            cred.ID,
		Name:          cred.Name,
		Type:          cred.Type,
		Value:         "******",
		ExpiresAt:     cred.ExpiresAt,
		LastRotatedAt: cred.LastRotatedAt,
		CreatedAt:     cred.CreatedAt.Time().Format(time.RFC3339),
		UpdatedAt:     cred.UpdatedAt.Time().Format(time.RFC3339),
	}})
}

func (h *Handler) UpdateCredential(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	envSlug := tenant.GetEnvironmentSlug(c)
	var cred aimodel.Credential
	if err := h.DB.Where("id = ? AND environment_slug = ?", id, envSlug).First(&cred).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "凭证不存在", "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	var body struct {
		Name      string     `json:"name" binding:"max=100"`
		Value     string     `json:"value" binding:"max=4096"`
		ExpiresAt *time.Time `json:"expiresAt"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error(), "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	if body.Name != "" {
		cred.Name = body.Name
	}
	if body.Value != "" {
		encrypted, err := crypto.Encrypt(body.Value)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "凭证加密失败", "request_id": runtime.RequestIDFromContext(c)})
			return
		}
		cred.EncryptedValue = encrypted
		now := time.Now()
		cred.LastRotatedAt = &now
	}
	if body.ExpiresAt != nil {
		cred.ExpiresAt = body.ExpiresAt
	}
	h.DB.Save(&cred)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": aimodel.CredentialResponse{
		ID:            cred.ID,
		Name:          cred.Name,
		Type:          cred.Type,
		Value:         "******",
		ExpiresAt:     cred.ExpiresAt,
		LastRotatedAt: cred.LastRotatedAt,
		CreatedAt:     cred.CreatedAt.Time().Format(time.RFC3339),
		UpdatedAt:     cred.UpdatedAt.Time().Format(time.RFC3339),
	}})
}

func (h *Handler) DeleteCredential(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	envSlug := tenant.GetEnvironmentSlug(c)
	result := h.DB.Where("id = ? AND environment_slug = ?", id, envSlug).Delete(&aimodel.Credential{})
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "凭证不存在", "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0})
}

// RotateCredential 单条凭证密钥轮转（区别于 PUT 整体更新）。
func (h *Handler) RotateCredential(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	envSlug := tenant.GetEnvironmentSlug(c)
	var cred aimodel.Credential
	if err := h.DB.Where("id = ? AND environment_slug = ?", id, envSlug).First(&cred).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "凭证不存在", "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	var body struct {
		NewValue string `json:"newValue" binding:"required,min=1,max=4096"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error(), "request_id": runtime.RequestIDFromContext(c)})
		return
	}

	encrypted, err := crypto.Encrypt(body.NewValue)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "凭证加密失败", "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	now := time.Now()
	h.DB.Model(&cred).Updates(map[string]interface{}{
		"encrypted_value":  encrypted,
		"last_rotated_at":  now,
	})

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": aimodel.CredentialResponse{
		ID:            cred.ID,
		Name:          cred.Name,
		Type:          cred.Type,
		Value:         "******",
		ExpiresAt:     cred.ExpiresAt,
		LastRotatedAt: &now,
		CreatedAt:     cred.CreatedAt.Time().Format(time.RFC3339),
		UpdatedAt:     now.Format(time.RFC3339),
	}})
}

// GetCredentialValue 内部 API：解密凭证值（REQ-093）。
func (h *Handler) GetCredentialValue(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var cred aimodel.Credential
	if err := h.DB.First(&cred, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "凭证不存在", "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	plaintext, err := crypto.Decrypt(cred.EncryptedValue)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "凭证解密失败", "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": aimodel.CredentialValueResponse{
		ID:    cred.ID,
		Value: plaintext,
		Type:  cred.Type,
	}})
}

// ReEncryptCredentials 密钥轮转（REQ-093）。
func (h *Handler) ReEncryptCredentials(c *gin.Context) {
	var body struct {
		NewKeyBase64 string `json:"newKeyBase64" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error(), "request_id": runtime.RequestIDFromContext(c)})
		return
	}

	var credentials []aimodel.Credential
	h.DB.Find(&credentials)

	var rotated, failed int
	for _, cred := range credentials {
		reEncrypted, err := crypto.ReEncrypt(cred.EncryptedValue, body.NewKeyBase64)
		if err != nil {
			failed++
			continue
		}
		h.DB.Model(&aimodel.Credential{}).Where("id = ?", cred.ID).Update("encrypted_value", reEncrypted)
		rotated++
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"rotated": rotated, "failed": failed}})
}

// NotifyExpiringCredentials 凭证过期提醒（内部逻辑）。
func (h *Handler) NotifyExpiringCredentials() {
	var expiring []aimodel.Credential
	threshold := time.Now().Add(7 * 24 * time.Hour)
	h.DB.Where("expires_at IS NOT NULL AND expires_at <= ?", threshold).Find(&expiring)

	for _, cred := range expiring {
		days := int(time.Until(*cred.ExpiresAt).Hours() / 24)
		data, _ := json.Marshal(aimodel.ExpiredCredentialNotification{
			CredentialID:   cred.ID,
			CredentialName: cred.Name,
			ExpiresAt:      *cred.ExpiresAt,
			DaysRemaining:  days,
		})
		h.DB.Create(&aimodel.NotifyRecord{
			Source:      "credential",
			Environment: cred.EnvironmentSlug,
			Level:       "warning",
			Message:     string(data),
		})
	}
}
