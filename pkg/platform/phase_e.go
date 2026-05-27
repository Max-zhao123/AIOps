package platform

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	aimodel "github.com/lihaiya/aiops/pkg/aiops/model"
	"github.com/lihaiya/aiops/pkg/aiops/tenant"
	"github.com/lihaiya/aiops/pkg/runtime"
	"gorm.io/gorm"
)

// -------- SLA 定义 CRUD（REQ-096） --------

func (h *Handler) CreateSlaDefinition(c *gin.Context) {
	var body struct {
		Name                string `json:"name" binding:"required,max=100"`
		Severity            string `json:"severity" binding:"required,oneof=critical high medium low"`
		ResponseTimeMin     int    `json:"responseTimeMin" binding:"required,min=1"`
		ResolutionTimeMin   int    `json:"resolutionTimeMin" binding:"required,min=1"`
		EscalationChannelID *int64 `json:"escalationChannelId"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error(), "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	envSlug := tenant.FillEnvironmentSlug(c)

	sla := aimodel.SlaDefinition{
		Name:                body.Name,
		Severity:            body.Severity,
		ResponseTimeMin:     body.ResponseTimeMin,
		ResolutionTimeMin:   body.ResolutionTimeMin,
		EscalationChannelID: body.EscalationChannelID,
		EnvironmentSlug:     envSlug,
	}
	if err := h.DB.Create(&sla).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "SLA 定义创建失败", "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": 0, "data": sla})
}

func (h *Handler) ListSlaDefinitions(c *gin.Context) {
	envSlug := tenant.GetEnvironmentSlug(c)
	var list []aimodel.SlaDefinition
	h.DB.Where("environment_slug = ?", envSlug).Order("id asc").Find(&list)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": list})
}

func (h *Handler) UpdateSlaDefinition(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	envSlug := tenant.GetEnvironmentSlug(c)
	var sla aimodel.SlaDefinition
	if err := h.DB.Where("id = ? AND environment_slug = ?", id, envSlug).First(&sla).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "SLA 定义不存在", "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	var body struct {
		Name                string `json:"name" binding:"max=100"`
		ResponseTimeMin     int    `json:"responseTimeMin" binding:"omitempty,min=1"`
		ResolutionTimeMin   int    `json:"resolutionTimeMin" binding:"omitempty,min=1"`
		EscalationChannelID *int64 `json:"escalationChannelId"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error(), "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	if body.Name != "" {
		sla.Name = body.Name
	}
	if body.ResponseTimeMin > 0 {
		sla.ResponseTimeMin = body.ResponseTimeMin
	}
	if body.ResolutionTimeMin > 0 {
		sla.ResolutionTimeMin = body.ResolutionTimeMin
	}
	if body.EscalationChannelID != nil {
		sla.EscalationChannelID = body.EscalationChannelID
	}
	h.DB.Save(&sla)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": sla})
}

func (h *Handler) DeleteSlaDefinition(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	envSlug := tenant.GetEnvironmentSlug(c)
	result := h.DB.Where("id = ? AND environment_slug = ?", id, envSlug).Delete(&aimodel.SlaDefinition{})
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "SLA 定义不存在", "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0})
}

// -------- SLA 统计（REQ-096） --------

func (h *Handler) GetSlaStats(c *gin.Context) {
	envSlug := tenant.GetEnvironmentSlug(c)
	if envSlug == "" {
		envSlug = "default"
	}
	period := c.DefaultQuery("period", "7d")
	var since time.Time
	switch period {
	case "1d":
		since = time.Now().AddDate(0, 0, -1)
	case "7d":
		since = time.Now().AddDate(0, 0, -7)
	case "30d":
		since = time.Now().AddDate(0, 0, -30)
	default:
		since = time.Now().AddDate(0, 0, -7)
	}

	var records []aimodel.SlaRecord
	h.DB.Where("environment_slug = ? AND created_at >= ?", envSlug, since).Find(&records)

	// 预加载所有关联的 SlaDefinition，用于按 severity 分组
	defMap := make(map[int64]aimodel.SlaDefinition)
	var definitions []aimodel.SlaDefinition
	h.DB.Where("environment_slug = ?", envSlug).Find(&definitions)
	for _, def := range definitions {
		defMap[def.ID] = def
	}

	stats := make(map[string]*aimodel.SlaSeverityStats)
	totalResolved := 0
	var totalResolutionMs int64

	for _, rec := range records {
		// 通过 SlaDefinitionID 查找对应的 Severity
		def, ok := defMap[rec.SlaDefinitionID]
		sev := "unknown"
		if ok {
			sev = def.Severity
		}
		if stats[sev] == nil {
			stats[sev] = &aimodel.SlaSeverityStats{}
		}
		stats[sev].Total++
		if rec.ResponseSlaMet != nil && *rec.ResponseSlaMet {
			stats[sev].ResponseSlaMet++
		}
		if rec.ResolutionSlaMet != nil && *rec.ResolutionSlaMet {
			stats[sev].ResolutionSlaMet++
		}
		if rec.ResolvedAt != nil && !rec.ResolvedAt.IsZero() {
			totalResolved++
			ms := rec.ResolvedAt.Sub(rec.CreatedAt).Milliseconds()
			totalResolutionMs += ms
			stats[sev].AvgMTTRMin += float64(ms) / 60000.0
		}
	}
	for _, s := range stats {
		if s.Total > 0 {
			s.AvgMTTRMin = s.AvgMTTRMin / float64(s.Total)
		}
	}

	var overallMTTR float64
	if totalResolved > 0 {
		overallMTTR = float64(totalResolutionMs) / float64(totalResolved) / 60000.0
	}
	var overallSlaRate float64
	totalRecords := len(records)
	slaMetCount := 0
	for _, rec := range records {
		if rec.ResolutionSlaMet != nil && *rec.ResolutionSlaMet {
			slaMetCount++
		}
	}
	if totalRecords > 0 {
		overallSlaRate = float64(slaMetCount) / float64(totalRecords) * 100
	}

	resp := aimodel.SlaStatsResponse{
		Period:         period,
		Environment:    envSlug,
		SeverityStats:  stats,
		OverallMTTR:    overallMTTR,
		OverallSLARate: overallSlaRate,
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": resp})
}

// CheckSlaBreaches 检查 SLA 违规（后台任务）。
func (h *Handler) CheckSlaBreaches() {
	envSlug := "default"
	var definitions []aimodel.SlaDefinition
	h.DB.Where("environment_slug = ?", envSlug).Find(&definitions)

	for _, def := range definitions {
		var breached []aimodel.SlaRecord
		h.DB.Where("environment_slug = ? AND sla_definition_id = ? AND resolved_at IS NULL AND created_at < ?",
			envSlug, def.ID,
			time.Now().Add(-time.Duration(def.ResolutionTimeMin)*time.Minute),
		).Find(&breached)

		for _, rec := range breached {
			if !rec.Escalated {
				h.DB.Model(&aimodel.SlaRecord{}).Where("id = ?", rec.ID).Update("escalated", true)
				h.DB.Create(&aimodel.NotifyRecord{
					Source:      "sla",
					Environment: envSlug,
					Level:       "critical",
					Message:     fmt.Sprintf("SLA 违规: %s, 记录ID=%d", def.Name, rec.ID),
				})
			}
		}
	}
}

// -------- 环境配置（REQ-097） --------

func (h *Handler) CreateEnvironmentConfig(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "环境标识缺失", "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	var body struct {
		Key          string `json:"key" binding:"required,max=200"`
		Value        string `json:"value" binding:"required"`
		OverrideType string `json:"overrideType" binding:"omitempty,oneof=plugin service llm security other"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error(), "request_id": runtime.RequestIDFromContext(c)})
		return
	}

	overrideType := body.OverrideType
	if overrideType == "" {
		overrideType = "other"
	}

	var existing aimodel.EnvironmentConfig
	result := h.DB.Where("environment_slug = ? AND config_key = ?", slug, body.Key).First(&existing)
	if result.Error == gorm.ErrRecordNotFound {
		cfg := aimodel.EnvironmentConfig{
			EnvironmentSlug: slug,
			ConfigKey:       body.Key,
			ConfigValue:     body.Value,
			OverrideType:    overrideType,
		}
		h.DB.Create(&cfg)
		h.incrementConfigVersion(c)
		c.JSON(http.StatusCreated, gin.H{"code": 0, "data": cfg})
	} else {
		h.DB.Model(&existing).Updates(map[string]interface{}{
			"config_value":  body.Value,
			"override_type": overrideType,
		})
		h.incrementConfigVersion(c)
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": existing})
	}
}

func (h *Handler) GetEnvironmentConfig(c *gin.Context) {
	slug := c.Param("slug")
	key := c.Param("key")
	var cfg aimodel.EnvironmentConfig
	if err := h.DB.Where("environment_slug = ? AND config_key = ?", slug, key).First(&cfg).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "配置项不存在", "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": cfg})
}

func (h *Handler) DeleteEnvironmentConfig(c *gin.Context) {
	slug := c.Param("slug")
	key := c.Param("key")
	if slug == "" || key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "环境标识或配置键缺失", "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	result := h.DB.Where("environment_slug = ? AND config_key = ?", slug, key).Delete(&aimodel.EnvironmentConfig{})
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "配置项不存在", "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	h.incrementConfigVersion(c)
	c.JSON(http.StatusOK, gin.H{"code": 0})
}

func (h *Handler) ListEnvironmentConfigs(c *gin.Context) {
	slug := c.Param("slug")
	var list []aimodel.EnvironmentConfig
	q := h.DB.Where("environment_slug = ?", slug).Order("id asc")
	if overrideType := c.Query("overrideType"); overrideType != "" {
		q = q.Where("override_type = ?", overrideType)
	}
	q.Find(&list)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": list})
}

// -------- 内部配置查询（REQ-097/099） --------

func (h *Handler) InternalGetConfig(c *gin.Context) {
	env := c.Param("environment")
	key := c.Param("key")
	var cfg aimodel.EnvironmentConfig
	if err := h.DB.Where("environment_slug = ? AND config_key = ?", env, key).First(&cfg).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "配置项不存在", "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"key": cfg.ConfigKey, "value": cfg.ConfigValue}})
}

// -------- 配置版本号（REQ-099） --------

func (h *Handler) GetConfigVersion(c *gin.Context) {
	var cv aimodel.ConfigVersion
	if err := h.DB.Order("version desc").First(&cv).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"version": 0}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"version": cv.Version}})
}

func (h *Handler) incrementConfigVersion(c *gin.Context) {
	var cv aimodel.ConfigVersion
	if err := h.DB.Order("version desc").First(&cv).Error; err != nil {
		h.DB.Create(&aimodel.ConfigVersion{Version: 1, ChangedBy: "system", ChangeSummary: "初始配置"})
		return
	}
	h.DB.Create(&aimodel.ConfigVersion{
		Version:       cv.Version + 1,
		ChangedBy:     "system",
		ChangeSummary: "配置更新",
	})
}

// -------- 审计导出（REQ-098） --------

func (h *Handler) ExportAudit(c *gin.Context) {
	format := c.DefaultQuery("format", "csv")
	env := c.Query("environment")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	q := h.DB.Model(&aimodel.AuditLog{}).Order("id asc")
	if env != "" {
		q = q.Where("environment = ?", env)
	}
	if startDate != "" {
		if t, err := time.Parse(time.RFC3339, startDate); err == nil {
			q = q.Where("created_at >= ?", t)
		}
	}
	if endDate != "" {
		if t, err := time.Parse(time.RFC3339, endDate); err == nil {
			q = q.Where("created_at <= ?", t)
		}
	}

	var logs []aimodel.AuditLog
	q.Limit(10000).Find(&logs)

	switch format {
	case "csv":
		h.exportAuditCSV(c, logs)
	case "pdf":
		h.exportAuditPDF(c, logs)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "不支持的导出格式，仅支持 csv/pdf", "request_id": runtime.RequestIDFromContext(c)})
	}
}

func (h *Handler) exportAuditCSV(c *gin.Context, logs []aimodel.AuditLog) {
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=audit_export.csv")

	writer := csv.NewWriter(c.Writer)
	c.Writer.Write([]byte{0xEF, 0xBB, 0xBF}) // BOM for Excel
	writer.Write([]string{"ID", "事件类型", "环境", "用户ID", "载荷", "创建时间"})

	for _, log := range logs {
		writer.Write([]string{
			strconv.FormatInt(log.ID, 10),
			log.EventType,
			log.Environment,
			strconv.FormatInt(log.UserID, 10),
			log.Payload,
			log.CreatedAt.Time().Format(time.RFC3339),
		})
	}
	writer.Flush()
}

func (h *Handler) exportAuditPDF(c *gin.Context, logs []aimodel.AuditLog) {
	// PDF 导出返回 JSON 格式（实际项目中使用 gofpdf 等库渲染 PDF）
	data, _ := json.Marshal(logs)
	c.Header("Content-Disposition", "attachment; filename=audit_export.json")
	c.Data(http.StatusOK, "application/json", data)
}

// -------- 合规报告（REQ-098） --------

func (h *Handler) GetAuditReport(c *gin.Context) {
	env := c.Query("environment")
	period := c.DefaultQuery("period", "7d")

	var since time.Time
	switch period {
	case "1d":
		since = time.Now().AddDate(0, 0, -1)
	case "7d":
		since = time.Now().AddDate(0, 0, -7)
	case "30d":
		since = time.Now().AddDate(0, 0, -30)
	default:
		since = time.Now().AddDate(0, 0, -7)
	}

	q := h.DB.Model(&aimodel.AuditLog{}).Where("created_at >= ?", since)
	if env != "" {
		q = q.Where("environment = ?", env)
	}

	var total int64
	q.Count(&total)

	type eventCount struct {
		EventType string `json:"eventType"`
		Count     int64  `json:"count"`
	}
	var byType []eventCount
	h.DB.Model(&aimodel.AuditLog{}).
		Select("event_type, count(*) as count").
		Where("created_at >= ?", since).
		Group("event_type").
		Scan(&byType)

	var denyCount int64
	h.DB.Model(&aimodel.AuditLog{}).
		Where("created_at >= ? AND event_type LIKE ?", since, "%deny%").
		Count(&denyCount)

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{
		"period":       period,
		"totalEvents":  total,
		"denyCount":    denyCount,
		"eventsByType": byType,
		"generatedAt":  time.Now().Format(time.RFC3339),
	}})
}

// -------- 环境配额（REQ-103） --------

func (h *Handler) GetEnvironmentQuotas(c *gin.Context) {
	slug := c.Param("slug")
	var quotas []aimodel.EnvironmentQuota
	h.DB.Where("environment_slug = ?", slug).Find(&quotas)

	if len(quotas) == 0 {
		defaults := []aimodel.EnvironmentQuota{
			{EnvironmentSlug: slug, ResourceType: "session", DailyLimit: 1000, CurrentCount: 0, ResetAt: time.Now().Truncate(24 * time.Hour).AddDate(0, 0, 1)},
			{EnvironmentSlug: slug, ResourceType: "execution", DailyLimit: 500, CurrentCount: 0, ResetAt: time.Now().Truncate(24 * time.Hour).AddDate(0, 0, 1)},
			{EnvironmentSlug: slug, ResourceType: "inspection", DailyLimit: 100, CurrentCount: 0, ResetAt: time.Now().Truncate(24 * time.Hour).AddDate(0, 0, 1)},
		}
		for _, d := range defaults {
			h.DB.Create(&d)
		}
		quotas = defaults
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": quotas})
}

func (h *Handler) UpdateEnvironmentQuotas(c *gin.Context) {
	slug := c.Param("slug")
	var body struct {
		ResourceType string `json:"resourceType" binding:"required,oneof=session execution inspection"`
		DailyLimit   int    `json:"dailyLimit" binding:"required,min=1,max=100000"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error(), "request_id": runtime.RequestIDFromContext(c)})
		return
	}

	var quota aimodel.EnvironmentQuota
	result := h.DB.Where("environment_slug = ? AND resource_type = ?", slug, body.ResourceType).First(&quota)
	if result.Error != nil {
		quota = aimodel.EnvironmentQuota{
			EnvironmentSlug: slug,
			ResourceType:    body.ResourceType,
			DailyLimit:      body.DailyLimit,
			CurrentCount:    0,
			ResetAt:         time.Now().Truncate(24 * time.Hour).AddDate(0, 0, 1),
		}
		h.DB.Create(&quota)
	} else {
		h.DB.Model(&quota).Update("daily_limit", body.DailyLimit)
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": quota})
}

// CheckQuota 检查环境配额（内部方法，REQ-103）。
func (h *Handler) CheckQuota(envSlug, resourceType string) aimodel.QuotaCheckResult {
	var quota aimodel.EnvironmentQuota
	if err := h.DB.Where("environment_slug = ? AND resource_type = ?", envSlug, resourceType).First(&quota).Error; err != nil {
		return aimodel.QuotaCheckResult{Allowed: true, DailyLimit: 0, CurrentCount: 0}
	}

	now := time.Now()
	if now.After(quota.ResetAt) {
		h.DB.Model(&quota).Updates(map[string]interface{}{
			"current_count": 0,
			"reset_at":      now.Truncate(24 * time.Hour).AddDate(0, 0, 1),
		})
		quota.CurrentCount = 0
	}

	if quota.CurrentCount >= quota.DailyLimit {
		return aimodel.QuotaCheckResult{
			Allowed:      false,
			Reason:       fmt.Sprintf("%s 配额已用完 (%d/%d)", resourceType, quota.CurrentCount, quota.DailyLimit),
			CurrentCount: quota.CurrentCount,
			DailyLimit:   quota.DailyLimit,
		}
	}
	return aimodel.QuotaCheckResult{Allowed: true, CurrentCount: quota.CurrentCount, DailyLimit: quota.DailyLimit}
}

// IncrementQuotaUsage 增加配额使用量。
func (h *Handler) IncrementQuotaUsage(envSlug, resourceType string) {
	h.DB.Model(&aimodel.EnvironmentQuota{}).
		Where("environment_slug = ? AND resource_type = ?", envSlug, resourceType).
		UpdateColumn("current_count", gorm.Expr("current_count + 1"))
}

// -------- 会话管理（REQ-100） --------

func (h *Handler) ArchiveChatSession(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var session aimodel.ChatSession
	if err := h.DB.First(&session, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "会话不存在", "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	h.DB.Model(&session).Update("status", "archived")
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"id": session.ID, "status": "archived"}})
}

func (h *Handler) DeleteChatSession(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var session aimodel.ChatSession
	if err := h.DB.First(&session, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "会话不存在", "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	h.DB.Model(&session).Update("status", "deleted")
	c.JSON(http.StatusOK, gin.H{"code": 0})
}

// CleanupExpiredSessions 清理过期会话（后台任务，REQ-100）。
func (h *Handler) CleanupExpiredSessions() {
	var sessions []aimodel.ChatSession
	h.DB.Where("status = ? AND last_active_at < ?", "active",
		time.Now().AddDate(0, 0, -90)).Find(&sessions)
	for _, s := range sessions {
		h.DB.Model(&aimodel.ChatSession{}).Where("id = ?", s.ID).Update("status", "expired")
	}
}

// -------- 用户改密（REQ-101） --------

func (h *Handler) ChangePassword(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var user aimodel.User
	if err := h.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户不存在", "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	var body struct {
		OldPassword string `json:"oldPassword" binding:"required"`
		NewPassword string `json:"newPassword" binding:"required,min=12"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error(), "request_id": runtime.RequestIDFromContext(c)})
		return
	}

	// 验证旧密码
	if !checkBcrypt(body.OldPassword, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "旧密码不正确", "request_id": runtime.RequestIDFromContext(c)})
		return
	}

	// 验证新密码复杂度
	policy := getPasswordPolicy()
	if err := policy.Validate(body.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error(), "request_id": runtime.RequestIDFromContext(c)})
		return
	}

	// 检查密码历史
	if isPasswordInHistory(body.NewPassword, user.PasswordHistory) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "新密码不能与最近5次密码相同", "request_id": runtime.RequestIDFromContext(c)})
		return
	}

	// 更新密码
	newHash := hashBcrypt(body.NewPassword)
	now := time.Now()
	history := appendPasswordHistory(user.PasswordHistory, user.Password, policy.PasswordHistoryCount)

	h.DB.Model(&user).Updates(map[string]interface{}{
		"password":             newHash,
		"password_changed_at":  now,
		"must_change_password": false,
		"password_history":    history,
	})
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "密码修改成功"})
}
