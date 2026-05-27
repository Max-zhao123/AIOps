package platform

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	aimodel "github.com/lihaiya/aiops/pkg/aiops/model"
	"github.com/lihaiya/aiops/pkg/aiops/types"
	"github.com/lihaiya/aiops/pkg/aiops/dataquery"
	"github.com/lihaiya/aiops/pkg/aiops/httpclient"
	"github.com/lihaiya/aiops/pkg/boundary"
	"github.com/lihaiya/aiops/pkg/runtime"
	"gorm.io/gorm"
)

// RegisterExtra 阶段 B/C/D API（在 Register 之后调用）。
func RegisterExtra(r *gin.Engine, db *gorm.DB) {
	h := &Handler{DB: db}
	api := r.Group("/api/v1")
	api.POST("/environments", h.CreateEnvironment)
	api.POST("/security-boundaries/:id/simulate", h.SimulatePolicy)
	api.POST("/data/query", h.DataQuery)
	api.GET("/risk-alerts", h.ListRiskAlerts)
	api.POST("/risk-alerts/scan", h.ScanRiskAlerts)

	api.GET("/inspections", h.ListInspections)
	api.POST("/inspections", h.CreateInspection)
	api.POST("/inspections/:id/run", h.RunInspection)
	api.GET("/inspections/:id/reports", h.ListInspectionReports)

	api.POST("/im/webhook", h.ImWebhook)
	api.POST("/helpdesk/chat", h.HelpdeskChat)

	api.GET("/runbooks", h.ListRunbooks)
	api.POST("/runbooks/:id/execute", h.ExecuteRunbook)

	api.POST("/internal/v1/worker/jobs", h.EnqueueWorkerJob)
	api.GET("/internal/v1/worker/jobs/pending", h.PendingWorkerJobs)
	api.POST("/internal/v1/worker/jobs/:id/process", h.ProcessWorkerJobHTTP)
	api.POST("/internal/v1/worker/jobs/:id/complete", h.CompleteWorkerJob)

	api.POST("/runbooks", h.CreateRunbook)

	api.GET("/llm/config", h.ListLlmConfigs)
	api.POST("/llm/config", h.CreateLlmConfig)
	api.PUT("/llm/config/:id", h.UpdateLlmConfig)
	api.DELETE("/llm/config/:id", h.DeleteLlmConfig)
}

// RegisterExtraE 阶段 E API（REQ-090~103）。
func RegisterExtraE(r *gin.Engine, db *gorm.DB) {
	h := &Handler{DB: db}
	api := r.Group("/api/v1")
	{
		// REQ-090: 定时调度
		api.GET("/schedules", h.ListSchedules)
		api.POST("/schedules", h.CreateSchedule)
		api.GET("/schedules/:id", h.GetSchedule)
		api.PUT("/schedules/:id", h.UpdateSchedule)
		api.DELETE("/schedules/:id", h.DeleteSchedule)
		api.GET("/schedules/:id/executions", h.ListScheduleExecutions)
		api.PATCH("/schedules/:id/toggle", h.ToggleSchedule)
		api.POST("/schedules/:id/trigger", h.TriggerSchedule)

		// REQ-091: 告警通知通道
		api.POST("/notification-channels", h.CreateNotificationChannel)
		api.GET("/notification-channels", h.ListNotificationChannels)
		api.PUT("/notification-channels/:id", h.UpdateNotificationChannel)
		api.DELETE("/notification-channels/:id", h.DeleteNotificationChannel)
		api.POST("/notification-channels/:id/test", h.TestNotificationChannel)

		// REQ-091: 通知策略
		api.POST("/notification-policies", h.CreateNotificationPolicy)
		api.GET("/notification-policies", h.ListNotificationPolicies)
		api.PUT("/notification-policies/:id", h.UpdateNotificationPolicy)
		api.DELETE("/notification-policies/:id", h.DeleteNotificationPolicy)

		// REQ-091: 通知模板
		api.GET("/notification-templates", h.ListNotificationTemplates)
		api.POST("/notification-templates", h.CreateNotificationTemplate)
		api.PUT("/notification-templates/:id", h.UpdateNotificationTemplate)
		api.DELETE("/notification-templates/:id", h.DeleteNotificationTemplate)

		// REQ-093: 凭证管理
		api.GET("/credentials", h.ListCredentials)
		api.POST("/credentials", h.CreateCredential)
		api.GET("/credentials/:id", h.GetCredential)
		api.PUT("/credentials/:id", h.UpdateCredential)
		api.DELETE("/credentials/:id", h.DeleteCredential)
		api.POST("/credentials/:id/rotate", h.RotateCredential)
		api.POST("/credentials/re-encrypt", h.ReEncryptCredentials)

		// REQ-094: 审批策略（通过安全边界配置管理）
		// 审批策略由 approval_policies 表独立管理
		api.GET("/approval-policies", h.ListApprovalPolicies)
		api.POST("/approval-policies", h.CreateApprovalPolicy)
		api.PUT("/approval-policies/:id", h.UpdateApprovalPolicy)

		// REQ-096: SLA
		api.POST("/sla-definitions", h.CreateSlaDefinition)
		api.GET("/sla-definitions", h.ListSlaDefinitions)
		api.PUT("/sla-definitions/:id", h.UpdateSlaDefinition)
		api.DELETE("/sla-definitions/:id", h.DeleteSlaDefinition)
		api.GET("/sla/stats", h.GetSlaStats)

		// REQ-097/099: 环境配置
		api.POST("/environments/:slug/config", h.CreateEnvironmentConfig)
		api.GET("/environments/:slug/configs", h.ListEnvironmentConfigs)
		api.GET("/environments/:slug/config/:key", h.GetEnvironmentConfig)
		api.DELETE("/environments/:slug/config/:key", h.DeleteEnvironmentConfig)

		// REQ-098: 审计导出
		api.GET("/audit/export", h.ExportAudit)
		api.GET("/audit/report", h.GetAuditReport)

		// REQ-100: 会话管理
		api.POST("/chat/sessions/:id/archive", h.ArchiveChatSession)
		api.DELETE("/chat/sessions/:id", h.DeleteChatSession)

		// REQ-101: 用户改密
		api.PUT("/users/:id/password", h.ChangePassword)

		// REQ-103: 环境配额
		api.GET("/environments/:slug/quotas", h.GetEnvironmentQuotas)
		api.PUT("/environments/:slug/quotas", h.UpdateEnvironmentQuotas)
	}

	// 内部 API（集群内调用）
	internal := r.Group("/internal/v1")
	{
		// REQ-093: 内部凭证解密
		internal.GET("/credentials/:id/value", h.GetCredentialValue)

		// REQ-097/099: 内部配置查询
		internal.GET("/config/:environment/:key", h.InternalGetConfig)
		internal.GET("/config/version", h.GetConfigVersion)
	}
}

func (h *Handler) CreateEnvironment(c *gin.Context) {
	var body struct {
		Slug, Name, Description string
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	e := aimodel.Environment{Slug: body.Slug, Name: body.Name, Description: body.Description}
	h.DB.Create(&e)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": e})
}

func (h *Handler) UpdateEnvironment(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var body struct {
		Name, Description string
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var e aimodel.Environment
	if err := h.DB.First(&e, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if body.Name != "" { e.Name = body.Name }
	if body.Description != "" { e.Description = body.Description }
	h.DB.Save(&e)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": e})
}

func (h *Handler) DeleteEnvironment(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.DB.Delete(&aimodel.Environment{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0})
}

func (h *Handler) SimulatePolicy(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var p aimodel.SecurityBoundaryPolicy
	if err := h.DB.First(&p, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var body struct {
		Environment string           `json:"environment"`
		Plan        types.ActionPlan `json:"plan"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var spec types.PolicySpec
	var err error
	if body.Environment != "" {
		spec, err = LoadEnabledSpec(h.DB, body.Environment)
	} else {
		spec, err = boundary.ParseSpecYAML(p.SpecYAML)
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res := boundary.Evaluate(spec, body.Plan)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": res})
}

func (h *Handler) DataQuery(c *gin.Context) {
	var body struct {
		Plugin    string                 `json:"plugin"`
		Query     string                 `json:"query"`
		Start     string                 `json:"start"`
		End       string                 `json:"end"`
		Params    map[string]interface{} `json:"params"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	maxRange := 24 * time.Hour
	if body.Start != "" && body.End != "" {
		st, e1 := time.Parse(time.RFC3339, body.Start)
		en, e2 := time.Parse(time.RFC3339, body.End)
		if e1 == nil && e2 == nil && en.Sub(st) > maxRange {
			c.JSON(http.StatusBadRequest, gin.H{"error": "time range exceeds 24h"})
			return
		}
	}
	var endpoint string
	switch body.Plugin {
	case "prometheus":
		endpoint = runtime.EnvOr("AIOPS_PLUGIN_PROM_URL", "http://aiops-plugin-prometheus:8092")
	case "logs":
		endpoint = runtime.EnvOr("AIOPS_PLUGIN_LOGS_URL", "http://aiops-plugin-logs:8093")
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown plugin"})
		return
	}
	var out map[string]interface{}
	err := httpclient.PostJSON(c.Request.Context(), endpoint+"/internal/v1/query", body, c, &out)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	rowData, rowsTruncated := dataquery.TruncateRows(out, dataquery.DefaultMaxRows)
	truncatedData, truncated := dataquery.TruncateResult(rowData, dataquery.DefaultMaxBytes)
	resp := gin.H{"code": 0, "data": truncatedData}
	if rowsTruncated {
		resp["rowsTruncated"] = true
	}
	if truncated {
		resp["truncated"] = true
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) ListRiskAlerts(c *gin.Context) {
	var list []aimodel.RiskAlert
	h.DB.Where("status = ?", "open").Order("id desc").Limit(100).Find(&list)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": list})
}

func (h *Handler) ScanRiskAlerts(c *gin.Context) {
	var body struct {
		Environment string           `json:"environment"`
		Plan        types.ActionPlan `json:"plan"`
	}
	_ = c.ShouldBindJSON(&body)
	if deny, msg := boundary.HardDeny(body.Plan); deny {
		raw, _ := json.Marshal(body.Plan)
		a := aimodel.RiskAlert{Environment: body.Environment, PlanJSON: string(raw), Reason: msg, Status: "open"}
		h.DB.Create(&a)
	}
	c.JSON(http.StatusOK, gin.H{"code": 0})
}

func (h *Handler) ListInspections(c *gin.Context) {
	var list []aimodel.InspectionTask
	h.DB.Find(&list)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": list})
}

func (h *Handler) CreateInspection(c *gin.Context) {
	var body aimodel.InspectionTask
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.DB.Create(&body)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": body})
}

func (h *Handler) RunInspection(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var task aimodel.InspectionTask
	if err := h.DB.First(&task, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	report, err := h.runInspectionSteps(c.Request.Context(), task, c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": report})
}

func (h *Handler) ListInspectionReports(c *gin.Context) {
	id := c.Param("id")
	var list []aimodel.InspectionReport
	h.DB.Where("task_id = ?", id).Order("id desc").Find(&list)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": list})
}

func (h *Handler) ImWebhook(c *gin.Context) {
	var body struct {
		Channel string `json:"channel"`
		UserKey string `json:"userKey"`
		Text    string `json:"text"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	chatURL := runtime.EnvOr("AIOPS_CHAT_URL", "http://aiops-chat:8082")
	var chatOut map[string]interface{}
	_ = httpclient.PostJSON(c.Request.Context(), chatURL+"/api/v1/chat", map[string]interface{}{
		"environment": "production",
		"message":     body.Text,
	}, c, &chatOut)
	reply, _ := json.Marshal(chatOut)
	rec := aimodel.ImMessage{Channel: body.Channel, UserKey: body.UserKey, Content: body.Text, ReplyJSON: string(reply)}
	h.DB.Create(&rec)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": chatOut})
}

func (h *Handler) ListRunbooks(c *gin.Context) {
	var list []aimodel.Runbook
	h.DB.Find(&list)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": list})
}

func (h *Handler) ExecuteRunbook(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var rb aimodel.Runbook
	if err := h.DB.First(&rb, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var steps []types.ActionPlan
	_ = json.Unmarshal([]byte(rb.StepsJSON), &steps)
	execURL := runtime.EnvOr("AIOPS_EXECUTOR_URL", "http://aiops-executor:8084")
	var results []interface{}
	for _, plan := range steps {
		var out map[string]interface{}
		_ = httpclient.PostJSON(c.Request.Context(), execURL+"/internal/v1/run", types.RunRequest{
			Environment: rb.Environment, Plan: plan,
		}, c, &out)
		results = append(results, out)
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": results})
}

func (h *Handler) EnqueueWorkerJob(c *gin.Context) {
	var body aimodel.WorkerJob
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	body.Status = "pending"
	body.NextRunAt = time.Now().Unix()
	h.DB.Create(&body)
	c.JSON(http.StatusOK, gin.H{"ok": true, "id": body.ID})
}

func (h *Handler) PendingWorkerJobs(c *gin.Context) {
	var list []aimodel.WorkerJob
	h.DB.Where("status = ?", "pending").Order("id asc").Limit(20).Find(&list)
	c.JSON(http.StatusOK, gin.H{"items": list})
}

func (h *Handler) ProcessWorkerJobHTTP(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.ProcessWorkerJob(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) CompleteWorkerJob(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	h.DB.Model(&aimodel.WorkerJob{}).Where("id = ?", id).Updates(map[string]interface{}{"status": "done"})
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) CreateRunbook(c *gin.Context) {
	var body aimodel.Runbook
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.DB.Create(&body)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": body})
}

// NotifyFailure 巡检失败通知（REQ-062）。
func (h *Handler) NotifyFailure(env, msg string) {
	h.DB.Create(&aimodel.NotifyRecord{Source: "inspection", Environment: env, Level: "error", Message: msg})
}

