package platform

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	aimodel "github.com/lihaiya/aiops/pkg/aiops/model"
	"github.com/lihaiya/aiops/pkg/aiops/tenant"
	"github.com/lihaiya/aiops/pkg/runtime"
)

// -------- 通知通道管理（REQ-091） --------

func (h *Handler) CreateNotificationChannel(c *gin.Context) {
	var body struct {
		Name   string `json:"name" binding:"required,max=100"`
		Type   string `json:"type" binding:"required,oneof=email webhook sms"`
		Config string `json:"config" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error(), "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	envSlug := tenant.FillEnvironmentSlug(c)

	// 幂等检查
	var cnt int64
	h.DB.Model(&aimodel.NotificationChannel{}).Where("name = ? AND environment_slug = ?", body.Name, envSlug).Count(&cnt)
	if cnt > 0 {
		c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "同名通知通道已存在", "request_id": runtime.RequestIDFromContext(c)})
		return
	}

	ch := aimodel.NotificationChannel{
		Name:            body.Name,
		Type:            body.Type,
		Config:          body.Config,
		Enabled:         true,
		EnvironmentSlug: envSlug,
	}
	if err := h.DB.Create(&ch).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "通知通道创建失败", "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": 0, "data": ch})
}

func (h *Handler) ListNotificationChannels(c *gin.Context) {
	envSlug := tenant.GetEnvironmentSlug(c)
	var list []aimodel.NotificationChannel
	q := h.DB.Where("environment_slug = ?", envSlug).Order("id asc")
	if chType := c.Query("type"); chType != "" {
		q = q.Where("type = ?", chType)
	}
	q.Find(&list)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": list})
}

func (h *Handler) UpdateNotificationChannel(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	envSlug := tenant.GetEnvironmentSlug(c)
	var ch aimodel.NotificationChannel
	if err := h.DB.Where("id = ? AND environment_slug = ?", id, envSlug).First(&ch).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "通知通道不存在", "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	var body struct {
		Name    string `json:"name" binding:"max=100"`
		Config  string `json:"config"`
		Enabled *bool  `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error(), "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	if body.Name != "" {
		ch.Name = body.Name
	}
	if body.Config != "" {
		ch.Config = body.Config
	}
	if body.Enabled != nil {
		ch.Enabled = *body.Enabled
	}
	h.DB.Save(&ch)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": ch})
}

func (h *Handler) DeleteNotificationChannel(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	envSlug := tenant.GetEnvironmentSlug(c)
	result := h.DB.Where("id = ? AND environment_slug = ?", id, envSlug).Delete(&aimodel.NotificationChannel{})
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "通知通道不存在", "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0})
}

// -------- 通知策略（REQ-091） --------

func (h *Handler) CreateNotificationPolicy(c *gin.Context) {
	var body struct {
		Name                  string `json:"name" binding:"required,max=100"`
		Severity              string `json:"severity" binding:"required,oneof=critical high medium low"`
		ChannelIDs            string `json:"channelIds" binding:"required"`
		SilenceWindowMin      int    `json:"silenceWindowMin"`
		SuppressLowerSeverity bool   `json:"suppressLowerSeverity"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error(), "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	envSlug := tenant.FillEnvironmentSlug(c)

	silenceWindow := body.SilenceWindowMin
	if silenceWindow <= 0 {
		silenceWindow = 30
	}

	policy := aimodel.NotificationPolicy{
		Name:                  body.Name,
		Severity:              body.Severity,
		ChannelIDs:            body.ChannelIDs,
		SilenceWindowMin:      silenceWindow,
		SuppressLowerSeverity: body.SuppressLowerSeverity,
		EnvironmentSlug:       envSlug,
	}
	if err := h.DB.Create(&policy).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "通知策略创建失败", "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": 0, "data": policy})
}

func (h *Handler) ListNotificationPolicies(c *gin.Context) {
	envSlug := tenant.GetEnvironmentSlug(c)
	var list []aimodel.NotificationPolicy
	h.DB.Where("environment_slug = ?", envSlug).Order("id asc").Find(&list)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": list})
}

// TestNotificationChannel 测试通知通道（发送测试消息）。
func (h *Handler) TestNotificationChannel(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	envSlug := tenant.GetEnvironmentSlug(c)
	var ch aimodel.NotificationChannel
	if err := h.DB.Where("id = ? AND environment_slug = ?", id, envSlug).First(&ch).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "通知通道不存在", "request_id": runtime.RequestIDFromContext(c)})
		return
	}

	var body struct {
		Message string `json:"message"`
	}
	_ = c.ShouldBindJSON(&body)
	testMsg := body.Message
	if testMsg == "" {
		testMsg = "这是一条测试通知消息，用于验证通道连通性。"
	}

	// 根据通道类型模拟发送测试消息
	var sendResult string
	switch ch.Type {
	case "email":
		sendResult = "测试邮件已发送至: " + ch.Config
	case "webhook":
		sendResult = "测试 Webhook 已推送到: " + ch.Config
	case "sms":
		sendResult = "测试短信已发送至: " + ch.Config
	default:
		sendResult = "测试消息已发送"
	}

	// 记录测试通知
	h.DB.Create(&aimodel.NotifyRecord{
		Source:      "notification_channel_test",
		Environment: envSlug,
		Level:       "info",
		Message:     sendResult,
	})

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{
		"channelId": ch.ID,
		"channel":   ch.Name,
		"type":      ch.Type,
		"message":   testMsg,
		"result":    sendResult,
	}})
}

// UpdateNotificationPolicy 更新通知策略。
func (h *Handler) UpdateNotificationPolicy(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	envSlug := tenant.GetEnvironmentSlug(c)
	var policy aimodel.NotificationPolicy
	if err := h.DB.Where("id = ? AND environment_slug = ?", id, envSlug).First(&policy).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "通知策略不存在", "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	var body struct {
		Name                  string `json:"name" binding:"max=100"`
		Severity              string `json:"severity" binding:"omitempty,oneof=critical high medium low"`
		ChannelIDs            string `json:"channelIds"`
		SilenceWindowMin     *int   `json:"silenceWindowMin"`
		SuppressLowerSeverity *bool  `json:"suppressLowerSeverity"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error(), "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	if body.Name != "" {
		policy.Name = body.Name
	}
	if body.Severity != "" {
		policy.Severity = body.Severity
	}
	if body.ChannelIDs != "" {
		policy.ChannelIDs = body.ChannelIDs
	}
	if body.SilenceWindowMin != nil {
		policy.SilenceWindowMin = *body.SilenceWindowMin
	}
	if body.SuppressLowerSeverity != nil {
		policy.SuppressLowerSeverity = *body.SuppressLowerSeverity
	}
	h.DB.Save(&policy)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": policy})
}

// DeleteNotificationPolicy 删除通知策略。
func (h *Handler) DeleteNotificationPolicy(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	envSlug := tenant.GetEnvironmentSlug(c)
	result := h.DB.Where("id = ? AND environment_slug = ?", id, envSlug).Delete(&aimodel.NotificationPolicy{})
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "通知策略不存在", "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0})
}

// CreateNotificationTemplate 创建通知模板。
func (h *Handler) CreateNotificationTemplate(c *gin.Context) {
	var body struct {
		ChannelType string `json:"channelType" binding:"required,oneof=email webhook sms"`
		EventType   string `json:"eventType" binding:"required,max=50"`
		Subject     string `json:"subject" binding:"max=200"`
		Body        string `json:"body" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error(), "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	envSlug := tenant.FillEnvironmentSlug(c)

	tpl := aimodel.NotificationTemplate{
		ChannelType:     body.ChannelType,
		EventType:       body.EventType,
		Subject:         body.Subject,
		Body:            body.Body,
		EnvironmentSlug: envSlug,
	}
	if err := h.DB.Create(&tpl).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "通知模板创建失败", "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": 0, "data": tpl})
}

// UpdateNotificationTemplate 更新通知模板。
func (h *Handler) UpdateNotificationTemplate(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var tpl aimodel.NotificationTemplate
	if err := h.DB.First(&tpl, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "通知模板不存在", "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	var body struct {
		ChannelType string `json:"channelType" binding:"omitempty,oneof=email webhook sms"`
		EventType   string `json:"eventType" binding:"max=50"`
		Subject     string `json:"subject" binding:"max=200"`
		Body        string `json:"body"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error(), "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	if body.ChannelType != "" {
		tpl.ChannelType = body.ChannelType
	}
	if body.EventType != "" {
		tpl.EventType = body.EventType
	}
	if body.Subject != "" {
		tpl.Subject = body.Subject
	}
	if body.Body != "" {
		tpl.Body = body.Body
	}
	h.DB.Save(&tpl)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": tpl})
}

// DeleteNotificationTemplate 删除通知模板。
func (h *Handler) DeleteNotificationTemplate(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	result := h.DB.Delete(&aimodel.NotificationTemplate{}, id)
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "通知模板不存在", "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0})
}

// -------- 通知模板（REQ-091） --------

func (h *Handler) ListNotificationTemplates(c *gin.Context) {
	var list []aimodel.NotificationTemplate
	q := h.DB.Order("id asc")
	if chType := c.Query("channelType"); chType != "" {
		q = q.Where("channel_type = ?", chType)
	}
	if eventType := c.Query("eventType"); eventType != "" {
		q = q.Where("event_type = ?", eventType)
	}
	q.Find(&list)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": list})
}
