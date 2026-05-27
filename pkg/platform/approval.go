package platform

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	aimodel "github.com/lihaiya/aiops/pkg/aiops/model"
	"github.com/lihaiya/aiops/pkg/aiops/tenant"
	"github.com/lihaiya/aiops/pkg/runtime"
)

// -------- 审批策略（REQ-094） --------

func (h *Handler) ListApprovalPolicies(c *gin.Context) {
	envSlug := tenant.GetEnvironmentSlug(c)
	var list []aimodel.ApprovalPolicy
	h.DB.Where("environment_slug = ?", envSlug).Order("id asc").Find(&list)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": list})
}

func (h *Handler) CreateApprovalPolicy(c *gin.Context) {
	var body struct {
		RiskLevel             string `json:"riskLevel" binding:"required,oneof=write notify"`
		MinApprovers          int    `json:"minApprovers" binding:"min=1,max=10"`
		TimeoutMin            int    `json:"timeoutMin" binding:"min=1,max=1440"`
		TimeoutAction         string `json:"timeoutAction" binding:"oneof=REJECT ESCALATE"`
		RequireDifferentRoles bool   `json:"requireDifferentRoles"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error(), "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	envSlug := tenant.FillEnvironmentSlug(c)

	minApprovers := body.MinApprovers
	if minApprovers <= 0 {
		minApprovers = 1
	}
	timeoutMin := body.TimeoutMin
	if timeoutMin <= 0 {
		timeoutMin = 30
	}
	timeoutAction := body.TimeoutAction
	if timeoutAction == "" {
		timeoutAction = "REJECT"
	}

	policy := aimodel.ApprovalPolicy{
		RiskLevel:             body.RiskLevel,
		MinApprovers:          minApprovers,
		TimeoutMin:            timeoutMin,
		TimeoutAction:         timeoutAction,
		RequireDifferentRoles: body.RequireDifferentRoles,
		EnvironmentSlug:       envSlug,
	}
	if err := h.DB.Create(&policy).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "审批策略创建失败", "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": 0, "data": policy})
}

func (h *Handler) UpdateApprovalPolicy(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	envSlug := tenant.GetEnvironmentSlug(c)
	var policy aimodel.ApprovalPolicy
	if err := h.DB.Where("id = ? AND environment_slug = ?", id, envSlug).First(&policy).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "审批策略不存在", "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	var body struct {
		MinApprovers          int    `json:"minApprovers" binding:"omitempty,min=1,max=10"`
		TimeoutMin            int    `json:"timeoutMin" binding:"omitempty,min=1,max=1440"`
		TimeoutAction         string `json:"timeoutAction" binding:"omitempty,oneof=REJECT ESCALATE"`
		RequireDifferentRoles *bool  `json:"requireDifferentRoles"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error(), "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	if body.MinApprovers > 0 {
		policy.MinApprovers = body.MinApprovers
	}
	if body.TimeoutMin > 0 {
		policy.TimeoutMin = body.TimeoutMin
	}
	if body.TimeoutAction != "" {
		policy.TimeoutAction = body.TimeoutAction
	}
	if body.RequireDifferentRoles != nil {
		policy.RequireDifferentRoles = *body.RequireDifferentRoles
	}
	h.DB.Save(&policy)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": policy})
}
