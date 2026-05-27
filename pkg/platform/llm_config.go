package platform

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	aimodel "github.com/lihaiya/aiops/pkg/aiops/model"
)

// -------- LLM 配置 CRUD --------

func (h *Handler) ListLlmConfigs(c *gin.Context) {
	var list []aimodel.LlmConfig
	h.DB.Order("id asc").Find(&list)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": list})
}

func (h *Handler) CreateLlmConfig(c *gin.Context) {
	var body struct {
		Name    string `json:"name"`
		BaseURL string `json:"baseUrl"`
		APIKey  string `json:"apiKey"`
		Model   string `json:"model"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cfg := aimodel.LlmConfig{
		Name:    body.Name,
		BaseURL: body.BaseURL,
		APIKey:  body.APIKey,
		Model:   body.Model,
		Active:  true,
	}
	h.DB.Create(&cfg)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": cfg})
}

func (h *Handler) UpdateLlmConfig(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var body struct {
		Name    string `json:"name"`
		BaseURL string `json:"baseUrl"`
		APIKey  string `json:"apiKey"`
		Model   string `json:"model"`
		Active  *bool  `json:"active"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var cfg aimodel.LlmConfig
	if err := h.DB.First(&cfg, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if body.Name != "" { cfg.Name = body.Name }
	if body.BaseURL != "" { cfg.BaseURL = body.BaseURL }
	if body.APIKey != "" { cfg.APIKey = body.APIKey }
	if body.Model != "" { cfg.Model = body.Model }
	if body.Active != nil { cfg.Active = *body.Active }
	h.DB.Save(&cfg)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": cfg})
}

func (h *Handler) DeleteLlmConfig(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	h.DB.Delete(&aimodel.LlmConfig{}, id)
	c.JSON(http.StatusOK, gin.H{"code": 0})
}
