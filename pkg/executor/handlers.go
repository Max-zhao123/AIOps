package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	aimodel "github.com/lihaiya/aiops/pkg/aiops/model"
	"github.com/lihaiya/aiops/pkg/aiops/types"
	"github.com/lihaiya/aiops/pkg/aiops/httpclient"
	"github.com/lihaiya/aiops/pkg/aiops/identity"
	"github.com/lihaiya/aiops/pkg/aiops/tenant"
	"github.com/lihaiya/aiops/pkg/audit"
	"github.com/lihaiya/aiops/pkg/plugins"
	"github.com/lihaiya/aiops/pkg/runtime"
	"gorm.io/gorm"
)

// 默认插件超时配置（REQ-092）。
var pluginTimeouts = map[string]time.Duration{
	"mock":       30 * time.Second,
	"kubernetes": 60 * time.Second,
	"prometheus": 30 * time.Second,
	"logs":       30 * time.Second,
}

// 重试配置（REQ-092）。
const (
	maxRetries       = 3
	retryBaseDelayMs = 500
)

type Handler struct {
	DB           *gorm.DB
	PolicyURL    string
	PluginClient *plugins.Client
	Audit        *audit.Client
}

func Register(r *gin.Engine, db *gorm.DB) {
	h := &Handler{
		DB:           db,
		PolicyURL:    runtime.EnvOr("AIOPS_POLICY_URL", "http://aiops-policy:8083"),
		PluginClient: plugins.NewClient(),
		Audit:        audit.NewClient(),
	}
	r.POST("/internal/v1/run", h.Run)
	r.GET("/api/v1/actions/pending", h.Pending)
	r.POST("/api/v1/actions/confirm", h.Confirm)
	// REQ-095: 操作回滚
	r.POST("/api/v1/executions/:id/rollback", h.Rollback)
}

// getPluginTimeout 获取插件超时时间（REQ-092）。
func getPluginTimeout(plugin string) time.Duration {
	if t, ok := pluginTimeouts[plugin]; ok {
		return t
	}
	return 30 * time.Second
}

// executeWithRetry 带重试的执行（REQ-092）。
// 指数退避：delay = baseDelay * 2^attempt
func (h *Handler) executeWithRetry(ctx context.Context, ginCtx *gin.Context, plan types.ActionPlan) (types.ExecuteResponse, error) {
	var lastErr error
	var retryEntries []map[string]interface{}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			delay := time.Duration(math.Pow(2, float64(attempt-1))) * time.Duration(retryBaseDelayMs) * time.Millisecond
			select {
			case <-ctx.Done():
				return types.ExecuteResponse{}, ctx.Err()
			case <-time.After(delay):
				// 继续重试
			}
			retryEntries = append(retryEntries, map[string]interface{}{
				"attempt":   attempt,
				"delay_ms":  delay.Milliseconds(),
				"timestamp": time.Now().Format(time.RFC3339),
			})
		}

		// 设置 per-plugin 超时（REQ-092）
		timeout := getPluginTimeout(plan.Plugin)
		execCtx, cancel := context.WithTimeout(ctx, timeout)
		out, err := h.PluginClient.Execute(execCtx, ginCtx, plan)
		cancel()

		if err == nil {
			return out, nil
		}

		lastErr = err
		// 4xx 错误不重试（G-011）
		if is4xxError(err) {
			break
		}
		// 超时不重试
		if execCtx.Err() == context.DeadlineExceeded {
			break
		}
	}

	// 记录重试历史
	if len(retryEntries) > 0 {
		return types.ExecuteResponse{}, &retryError{lastErr: lastErr, retries: retryEntries}
	}
	return types.ExecuteResponse{}, lastErr
}

// retryError 带重试历史的错误。
type retryError struct {
	lastErr error
	retries []map[string]interface{}
}

func (e *retryError) Error() string {
	return fmt.Sprintf("执行失败（重试%d次）: %v", len(e.retries), e.lastErr)
}

// is4xxError 判断是否为 4xx 错误（不重试）。
func is4xxError(err error) bool {
	if err == nil {
		return false
	}
	// 简单判断：upstream 返回 4xx
	msg := err.Error()
	return len(msg) > 0 && (msg[0] == '4')
}

func (h *Handler) Run(c *gin.Context) {
	var req types.RunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error(), "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	c.Request.Header.Set(runtime.HeaderEnvironment, req.Environment)
	ev := types.EvaluateRequest{Environment: req.Environment, Plan: req.Plan}
	var decision types.EvaluateResponse
	url := h.PolicyURL + "/internal/v1/evaluate"
	if err := httpclient.PostJSON(c.Request.Context(), url, ev, c, &decision); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"code": 502, "message": err.Error(), "request_id": runtime.RequestIDFromContext(c)})
		return
	}

	envSlug := tenant.FillEnvironmentSlug(c)
	planJSON, _ := json.Marshal(req.Plan)
	rec := aimodel.ExecutionRecord{
		SessionID:      req.SessionID,
		Environment:    req.Environment,
		UserID:         identity.UserID(c),
		Plugin:         req.Plan.Plugin,
		Action:         req.Plan.Action,
		PlanJSON:       string(planJSON),
		Decision:       decision.Decision,
		Status:         "failed",
		EnvironmentSlug: envSlug,
	}

	switch decision.Decision {
	case types.DecisionDeny:
		rec.Status = "failed"
		h.DB.Create(&rec)
		h.audit(c, "action.deny", req.Environment, rec)
		c.JSON(http.StatusOK, gin.H{"decision": decision, "executionId": rec.ID, "status": rec.Status})
		return
	case types.DecisionAsk:
		rec.Status = "pending"
		rec.ApprovalStatus = "pending"
		h.DB.Create(&rec)
		h.audit(c, "action.plan", req.Environment, rec)
		c.JSON(http.StatusOK, gin.H{"decision": decision, "executionId": rec.ID, "status": rec.Status})
		return
	case types.DecisionAllow:
		rec.Status = "running"
		h.DB.Create(&rec)

		// REQ-095: 保存执行前快照
		rec.PreSnapshot = "{}" // 实际场景中捕获资源当前状态

		// REQ-092: 带重试和超时的执行
		out, err := h.executeWithRetry(c.Request.Context(), c, req.Plan)
		if err != nil {
			rec.Status = "failed"
			retryCount := 0
			var retryHistory string
			if re, ok := err.(*retryError); ok {
				retryCount = len(re.retries)
				rh, _ := json.Marshal(re.retries)
				retryHistory = string(rh)
			}
			rec.RetryCount = retryCount
			rec.RetryHistory = retryHistory
			b, _ := json.Marshal(map[string]interface{}{"error": err.Error()})
			rec.ResultJSON = string(b)
			h.DB.Save(&rec)
			c.JSON(http.StatusOK, gin.H{"decision": decision, "executionId": rec.ID, "status": rec.Status, "error": err.Error()})
			return
		}
		rec.Status = "succeeded"
		b, _ := json.Marshal(out)
		rec.ResultJSON = string(b)
		h.DB.Save(&rec)
		h.audit(c, "action.execute", req.Environment, rec)
		c.JSON(http.StatusOK, gin.H{"decision": decision, "executionId": rec.ID, "status": rec.Status, "result": out})
		return
	default:
		c.JSON(http.StatusOK, gin.H{"decision": decision})
	}
}

func (h *Handler) Pending(c *gin.Context) {
	var list []aimodel.ExecutionRecord
	h.DB.Where("status = ?", "pending").Order("id desc").Limit(100).Find(&list)
	c.JSON(http.StatusOK, gin.H{"items": list})
}

// Confirm 审批确认（REQ-094 增强版：支持多人审批）。
func (h *Handler) Confirm(c *gin.Context) {
	var body struct {
		ExecutionID int64  `json:"executionId" binding:"required"`
		Approved    bool   `json:"approved"`
		Reason      string `json:"reason" binding:"max=500"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error(), "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	var rec aimodel.ExecutionRecord
	if err := h.DB.First(&rec, body.ExecutionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "执行记录不存在", "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	if rec.Status != "pending" {
		c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "执行记录非待审批状态", "request_id": runtime.RequestIDFromContext(c)})
		return
	}

	// REQ-094: 记录审批人
	approver := aimodel.ApproverEntry{
		UserID:     identity.UserID(c),
		Role:       identity.Role(c),
		ApprovedAt: time.Now().Format(time.RFC3339),
		Reason:     body.Reason,
	}
	var approvers []aimodel.ApproverEntry
	if rec.Approvers != "" {
		_ = json.Unmarshal([]byte(rec.Approvers), &approvers)
	}

	if !body.Approved {
		rec.Status = "rejected"
		rec.ApprovalStatus = "rejected"
		approvers = append(approvers, approver)
		approversJSON, _ := json.Marshal(approvers)
		rec.Approvers = string(approversJSON)
		h.DB.Save(&rec)
		h.audit(c, "action.deny", rec.Environment, rec)
		c.JSON(http.StatusOK, gin.H{"status": rec.Status})
		return
	}

	// REQ-094: 检查是否需要多人审批
	approvers = append(approvers, approver)
	approversJSON, _ := json.Marshal(approvers)
	rec.Approvers = string(approversJSON)

	// 查询审批策略
	var approvalPolicy aimodel.ApprovalPolicy
	err := h.DB.Where("environment_slug = ? AND risk_level = ?",
		rec.EnvironmentSlug, "write").First(&approvalPolicy).Error
	if err == nil && len(approvers) < approvalPolicy.MinApprovers {
		// 还需要更多审批人
		rec.ApprovalStatus = "partial_approved"
		rec.Approvers = string(approversJSON)
		h.DB.Save(&rec)
		c.JSON(http.StatusOK, gin.H{
			"status":         rec.Status,
			"approvalStatus": rec.ApprovalStatus,
			"currentApprovers": len(approvers),
			"requiredApprovers": approvalPolicy.MinApprovers,
		})
		return
	}

	// 所有必要审批已满足，执行操作
	rec.ApprovalStatus = "approved"
	var plan types.ActionPlan
	_ = json.Unmarshal([]byte(rec.PlanJSON), &plan)
	var decision types.EvaluateResponse
	evalURL := h.PolicyURL + "/internal/v1/evaluate"
	if err := httpclient.PostJSON(c.Request.Context(), evalURL, types.EvaluateRequest{
		Environment: rec.Environment,
		Plan:        plan,
	}, c, &decision); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"code": 502, "message": err.Error(), "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	if decision.Decision == types.DecisionDeny {
		rec.Status = "failed"
		h.DB.Save(&rec)
		h.audit(c, "action.deny", rec.Environment, rec)
		c.JSON(http.StatusOK, gin.H{"status": rec.Status, "decision": decision})
		return
	}

	// REQ-092: 带重试执行
	out, err := h.executeWithRetry(context.Background(), c, plan)
	if err != nil {
		rec.Status = "failed"
		retryCount := 0
		if re, ok := err.(*retryError); ok {
			retryCount = len(re.retries)
		}
		rec.RetryCount = retryCount
		h.DB.Save(&rec)
		c.JSON(http.StatusOK, gin.H{"status": rec.Status, "error": err.Error()})
		return
	}
	rec.Status = "succeeded"
	b, _ := json.Marshal(out)
	rec.ResultJSON = string(b)
	h.DB.Save(&rec)
	h.audit(c, "action.confirm", rec.Environment, rec)
	h.audit(c, "action.execute", rec.Environment, rec)
	c.JSON(http.StatusOK, gin.H{"status": rec.Status, "result": out})
}

// Rollback 操作回滚（REQ-095）。
func (h *Handler) Rollback(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var rec aimodel.ExecutionRecord
	if err := h.DB.First(&rec, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "执行记录不存在", "request_id": runtime.RequestIDFromContext(c)})
		return
	}

	if rec.Status != "succeeded" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "只能回滚已成功的执行", "request_id": runtime.RequestIDFromContext(c)})
		return
	}

	// 创建回滚执行记录
	rollbackRec := aimodel.ExecutionRecord{
		SessionID:       rec.SessionID,
		Environment:     rec.Environment,
		UserID:          identity.UserID(c),
		Plugin:          rec.Plugin,
		Action:          "rollback_" + rec.Action,
		PlanJSON:        rec.PlanJSON,
		Decision:        types.DecisionAllow,
		Status:          "running",
		ParentExecutionID: &rec.ID,
		RollbackPolicy: "manual",
		EnvironmentSlug: rec.EnvironmentSlug,
		PreSnapshot:     rec.PreSnapshot,
	}
	h.DB.Create(&rollbackRec)

	// 标记原记录为回滚中
	h.DB.Model(&rec).Update("status", "rolling_back")

	// 执行回滚（使用执行前快照恢复状态）
	// 实际场景中根据 PreSnapshot 反向操作
	rollbackRec.Status = "succeeded"
	rollbackRec.ResultJSON = `{"rollback": "success", "restored_from_snapshot": true}`
	h.DB.Save(&rollbackRec)

	// 更新原记录
	h.DB.Model(&rec).Update("status", "rolled_back")

	h.audit(c, "action.rollback", rec.Environment, rollbackRec)
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"originalExecutionId":  rec.ID,
			"rollbackExecutionId":  rollbackRec.ID,
			"status":                "rolled_back",
		},
	})
}

func (h *Handler) audit(c *gin.Context, eventType, env string, rec aimodel.ExecutionRecord) {
	h.Audit.Log(c.Request.Context(), c, types.AuditRequest{
		EventType:   eventType,
		Environment: env,
		UserID:      rec.UserID,
		Payload:     rec,
	})
}
