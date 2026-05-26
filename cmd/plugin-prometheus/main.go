package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/lihaiya/aiops/pkg/plugins/prometheus"
	"github.com/lihaiya/aiops/pkg/runtime"
)

func main() {
	os.Setenv("AIOPS_SERVICE", "plugin-prometheus")
	runtime.Run("plugin-prometheus", 8092, func(r *gin.Engine) {
		r.POST("/internal/v1/query", func(c *gin.Context) {
			var req prometheus.QueryRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			out, err := prometheus.ExecuteQuery(req)
			if err != nil {
				c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, out)
		})
		r.POST("/internal/v1/execute", func(c *gin.Context) {
			var req prometheus.QueryRequest
			_ = c.ShouldBindJSON(&req)
			out, _ := prometheus.ExecuteQuery(req)
			c.JSON(http.StatusOK, gin.H{"status": "ok", "output": out})
		})
	})
}
