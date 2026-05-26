package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/lihaiya/aiops/pkg/plugins/logs"
	"github.com/lihaiya/aiops/pkg/runtime"
)

func main() {
	os.Setenv("AIOPS_SERVICE", "plugin-logs")
	runtime.Run("plugin-logs", 8093, func(r *gin.Engine) {
		r.POST("/internal/v1/query", func(c *gin.Context) {
			var req logs.QueryRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			out, err := logs.ExecuteQuery(req)
			if err != nil {
				c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, out)
		})
		r.POST("/internal/v1/execute", func(c *gin.Context) {
			var req logs.QueryRequest
			_ = c.ShouldBindJSON(&req)
			out, _ := logs.ExecuteQuery(req)
			c.JSON(http.StatusOK, gin.H{"status": "ok", "output": out})
		})
	})
}
