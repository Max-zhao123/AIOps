package kb

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	aimodel "github.com/lihaiya/aiops/pkg/aiops/model"
	"gorm.io/gorm"
)

type Handler struct {
	DB *gorm.DB
}

func Register(r *gin.Engine, db *gorm.DB) {
	h := &Handler{DB: db}
	api := r.Group("/api/v1/kb")
	api.GET("/documents", h.List)
	api.POST("/documents", h.Create)
	api.GET("/documents/:id", h.Get)
	api.DELETE("/documents/:id", h.Delete)
	r.POST("/internal/v1/kb/search", h.Search)
}

func (h *Handler) List(c *gin.Context) {
	var list []aimodel.KbDocument
	h.DB.Order("id desc").Limit(200).Find(&list)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": list})
}

func (h *Handler) Create(c *gin.Context) {
	var doc aimodel.KbDocument
	if err := c.ShouldBindJSON(&doc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.DB.Create(&doc)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": doc})
}

func (h *Handler) Get(c *gin.Context) {
	var doc aimodel.KbDocument
	if err := h.DB.First(&doc, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": doc})
}

func (h *Handler) Delete(c *gin.Context) {
	h.DB.Delete(&aimodel.KbDocument{}, c.Param("id"))
	c.JSON(http.StatusOK, gin.H{"code": 0})
}

func (h *Handler) Search(c *gin.Context) {
	var body struct {
		Query string `json:"query"`
		Limit int    `json:"limit"`
	}
	_ = c.ShouldBindJSON(&body)
	if body.Limit <= 0 {
		body.Limit = 5
	}
	q := "%" + strings.TrimSpace(body.Query) + "%"
	var list []aimodel.KbDocument
	h.DB.Where("title LIKE ? OR content LIKE ?", q, q).Limit(body.Limit).Find(&list)
	c.JSON(http.StatusOK, gin.H{"data": list})
}
