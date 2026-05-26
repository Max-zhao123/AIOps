package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/lihaiya/aiops/pkg/aiops/types"
	"github.com/lihaiya/aiops/pkg/plugins/kubernetes"
	"github.com/lihaiya/aiops/pkg/runtime"
)

func main() {
	os.Setenv("AIOPS_SERVICE", "plugin-kubernetes")
	runtime.Run("plugin-kubernetes", 8091, func(r *gin.Engine) {
		r.POST("/internal/v1/execute", func(c *gin.Context) {
			var req types.ExecuteRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			out, err := kubernetes.Execute(req.Plan)
			if err != nil {
				c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, out)
		})
	})
}
