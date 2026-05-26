package platform

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	aimodel "github.com/lihaiya/aiops/pkg/aiops/model"
	"github.com/lihaiya/aiops/pkg/aiops/types"
	"github.com/lihaiya/aiops/utils"
	"github.com/lihaiya/aiops/pkg/aiops/identity"
	"github.com/lihaiya/aiops/pkg/boundary"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

type Handler struct {
	DB *gorm.DB
}

func Register(r *gin.Engine, db *gorm.DB) {
	h := &Handler{DB: db}
	r.POST("/api/v1/auth/login", h.Login)
	api := r.Group("/api/v1")
	{
		api.GET("/environments", h.ListEnvironments)
		api.GET("/security-boundaries", h.ListPolicies)
		api.POST("/security-boundaries", h.CreatePolicy)
		api.GET("/security-boundaries/:id", h.GetPolicy)
		api.PUT("/security-boundaries/:id", h.UpdatePolicy)
		api.POST("/security-boundaries/:id/enable", h.EnablePolicy)
		api.POST("/security-boundaries/:id/disable", h.DisablePolicy)
		api.GET("/security-boundaries/import-template", h.ImportTemplate)
		api.GET("/plugins", h.ListPlugins)
		api.GET("/audit", h.ListAudit)
	}
	r.POST("/internal/v1/audit", h.InternalAudit)
}

func (h *Handler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var user aimodel.User
	if err := h.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	if !utils.BcryptCheck(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	j := utils.NewJWT()
	token, err := j.CreateToken(j.CreateClaims(utils.BaseClaims{ID: user.ID, Username: user.Username}))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"token": token, "role": user.Role}})
}

func (h *Handler) ListEnvironments(c *gin.Context) {
	var list []aimodel.Environment
	h.DB.Order("id asc").Find(&list)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": list, "message": "success"})
}

func (h *Handler) ListPolicies(c *gin.Context) {
	var list []aimodel.SecurityBoundaryPolicy
	h.DB.Order("id desc").Find(&list)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": list, "message": "success"})
}

func (h *Handler) CreatePolicy(c *gin.Context) {
	var body struct {
		EnvironmentID int64  `json:"environmentId"`
		Name          string `json:"name"`
		SpecYAML      string `json:"specYaml"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p := aimodel.SecurityBoundaryPolicy{
		EnvironmentID: body.EnvironmentID,
		Name:          body.Name,
		SpecYAML:      body.SpecYAML,
	}
	h.DB.Create(&p)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": p})
}

func (h *Handler) GetPolicy(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var p aimodel.SecurityBoundaryPolicy
	if err := h.DB.First(&p, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": p})
}

func (h *Handler) UpdatePolicy(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var p aimodel.SecurityBoundaryPolicy
	if err := h.DB.First(&p, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var body struct {
		SpecYAML string `json:"specYaml"`
		Name     string `json:"name"`
	}
	_ = c.ShouldBindJSON(&body)
	if body.SpecYAML != "" {
		p.SpecYAML = body.SpecYAML
	}
	if body.Name != "" {
		p.Name = body.Name
	}
	h.DB.Save(&p)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": p})
}

func (h *Handler) EnablePolicy(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var p aimodel.SecurityBoundaryPolicy
	if err := h.DB.First(&p, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	h.DB.Model(&aimodel.SecurityBoundaryPolicy{}).
		Where("environment_id = ?", p.EnvironmentID).
		Update("enabled", false)
	p.Enabled = true
	h.DB.Save(&p)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": p})
}

func (h *Handler) DisablePolicy(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	h.DB.Model(&aimodel.SecurityBoundaryPolicy{}).Where("id = ?", id).Update("enabled", false)
	c.JSON(http.StatusOK, gin.H{"code": 0})
}

func (h *Handler) ImportTemplate(c *gin.Context) {
	name := c.Query("name")
	var spec string
	switch name {
	case "prod-readonly":
		spec = `defaultDecision: ASK
rules:
  - id: read-allow
    match: { actions: ["get","list","describe","logs"] }
    decision: ALLOW
  - id: write-deny
    match: { actions: ["delete","apply","patch","create"] }
    decision: DENY`
	case "prod-standard", "dev-relaxed":
		spec = `defaultDecision: ASK
rules:
  - id: read-allow
    match: { actions: ["get","list","describe","logs"] }
    decision: ALLOW
  - id: write-ask
    match: { actions: ["delete","apply","patch","create"] }
    decision: ASK`
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown template"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"specYaml": spec}})
}

func (h *Handler) ListPlugins(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": []gin.H{
		{"name": "mock", "enabled": true, "actions": []string{"echo"}},
		{"name": "kubernetes", "enabled": true, "actions": []string{"get", "list", "describe", "logs"}},
		{"name": "prometheus", "enabled": true, "actions": []string{"query"}},
		{"name": "logs", "enabled": true, "actions": []string{"search"}},
	}})
}

func (h *Handler) ListAudit(c *gin.Context) {
	var list []aimodel.AuditLog
	q := h.DB.Order("id desc").Limit(200)
	if env := c.Query("environment"); env != "" {
		q = q.Where("environment = ?", env)
	}
	q.Find(&list)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": list})
}

func (h *Handler) InternalAudit(c *gin.Context) {
	var req types.AuditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	raw, _ := json.Marshal(req.Payload)
	if len(raw) > 4096 {
		raw = raw[:4096]
	}
	log := aimodel.AuditLog{
		EventType:   req.EventType,
		Environment: req.Environment,
		UserID:      req.UserID,
		Payload:     string(raw),
	}
	if log.UserID == 0 {
		log.UserID = identity.UserID(c)
	}
	if err := h.DB.Create(&log).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "audit failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// LoadEnabledSpec 读取环境当前 enabled 策略。
func LoadEnabledSpec(db *gorm.DB, environment string) (types.PolicySpec, error) {
	var env aimodel.Environment
	if err := db.Where("slug = ?", environment).First(&env).Error; err != nil {
		return types.PolicySpec{}, err
	}
	var p aimodel.SecurityBoundaryPolicy
	if err := db.Where("environment_id = ? AND enabled = ?", env.ID, true).First(&p).Error; err != nil {
		return types.PolicySpec{DefaultDecision: types.DecisionAsk}, nil
	}
	return boundary.ParseSpecYAML(p.SpecYAML)
}

// ParseSpecFromYAML API 用。
func ParseSpecFromYAML(raw string) (types.PolicySpec, error) {
	var spec types.PolicySpec
	err := yaml.Unmarshal([]byte(raw), &spec)
	return spec, err
}
