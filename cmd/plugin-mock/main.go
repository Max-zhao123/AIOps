package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/lihaiya/aiops/pkg/aiops/types"
	"github.com/lihaiya/aiops/pkg/plugins/mock"
	"github.com/lihaiya/aiops/pkg/runtime"
)

func main() {
	os.Setenv("AIOPS_SERVICE", "plugin-mock")
	runtime.Run("plugin-mock", 8090, func(r *gin.Engine) {
		r.POST("/internal/v1/execute", func(c *gin.Context) {
			var req types.ExecuteRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, mock.Execute(req.Plan))
		})
	})
}
