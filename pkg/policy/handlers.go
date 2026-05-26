package policy

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lihaiya/aiops/pkg/aiops/types"
	"github.com/lihaiya/aiops/pkg/boundary"
	"github.com/lihaiya/aiops/pkg/platform"
	"gorm.io/gorm"
)

type Handler struct {
	DB *gorm.DB
}

func Register(r *gin.Engine, db *gorm.DB) {
	h := &Handler{DB: db}
	r.POST("/internal/v1/evaluate", h.Evaluate)
}

func (h *Handler) Evaluate(c *gin.Context) {
	var req types.EvaluateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	spec, err := platform.LoadEnabledSpec(h.DB, req.Environment)
	if err != nil {
		res := boundary.FailClose(req.Plan, err)
		c.JSON(http.StatusOK, res)
		return
	}
	res := boundary.Evaluate(spec, req.Plan)
	c.JSON(http.StatusOK, res)
}
