package executor

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	aimodel "github.com/lihaiya/aiops/pkg/aiops/model"
	"github.com/lihaiya/aiops/pkg/aiops/types"
	"github.com/lihaiya/aiops/pkg/aiops/httpclient"
	"github.com/lihaiya/aiops/pkg/aiops/identity"
	"github.com/lihaiya/aiops/pkg/audit"
	"github.com/lihaiya/aiops/pkg/plugins"
	"github.com/lihaiya/aiops/pkg/runtime"
	"gorm.io/gorm"
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
}

func (h *Handler) Run(c *gin.Context) {
	var req types.RunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Request.Header.Set(runtime.HeaderEnvironment, req.Environment)
	ev := types.EvaluateRequest{Environment: req.Environment, Plan: req.Plan}
	var decision types.EvaluateResponse
	url := h.PolicyURL + "/internal/v1/evaluate"
	if err := httpclient.PostJSON(c.Request.Context(), url, ev, c, &decision); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	planJSON, _ := json.Marshal(req.Plan)
	rec := aimodel.ExecutionRecord{
		SessionID:   req.SessionID,
		Environment: req.Environment,
		UserID:      identity.UserID(c),
		Plugin:      req.Plan.Plugin,
		Action:      req.Plan.Action,
		PlanJSON:    string(planJSON),
		Decision:    decision.Decision,
		Status:      "failed",
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
		h.DB.Create(&rec)
		h.audit(c, "action.plan", req.Environment, rec)
		c.JSON(http.StatusOK, gin.H{"decision": decision, "executionId": rec.ID, "status": rec.Status})
		return
	case types.DecisionAllow:
		rec.Status = "running"
		h.DB.Create(&rec)
		out, err := h.PluginClient.Execute(c.Request.Context(), c, req.Plan)
		if err != nil {
			rec.Status = "failed"
			b, _ := json.Marshal(gin.H{"error": err.Error()})
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

func (h *Handler) Confirm(c *gin.Context) {
	var body struct {
		ExecutionID int64 `json:"executionId"`
		Approved    bool  `json:"approved"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var rec aimodel.ExecutionRecord
	if err := h.DB.First(&rec, body.ExecutionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if rec.Status != "pending" {
		c.JSON(http.StatusConflict, gin.H{"error": "not pending"})
		return
	}
	if !body.Approved {
		rec.Status = "rejected"
		h.DB.Save(&rec)
		h.audit(c, "action.deny", rec.Environment, rec)
		c.JSON(http.StatusOK, gin.H{"status": rec.Status})
		return
	}
	var plan types.ActionPlan
	_ = json.Unmarshal([]byte(rec.PlanJSON), &plan)
	var decision types.EvaluateResponse
	evalURL := h.PolicyURL + "/internal/v1/evaluate"
	if err := httpclient.PostJSON(c.Request.Context(), evalURL, types.EvaluateRequest{
		Environment: rec.Environment,
		Plan:        plan,
	}, c, &decision); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	if decision.Decision == types.DecisionDeny {
		rec.Status = "failed"
		h.DB.Save(&rec)
		h.audit(c, "action.deny", rec.Environment, rec)
		c.JSON(http.StatusOK, gin.H{"status": rec.Status, "decision": decision})
		return
	}
	out, err := h.PluginClient.Execute(context.Background(), c, plan)
	if err != nil {
		rec.Status = "failed"
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

func (h *Handler) audit(c *gin.Context, eventType, env string, rec aimodel.ExecutionRecord) {
	h.Audit.Log(c.Request.Context(), c, types.AuditRequest{
		EventType:   eventType,
		Environment: env,
		UserID:      rec.UserID,
		Payload:     rec,
	})
}
