package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/lihaiya/aiops/config"
	"github.com/lihaiya/aiops/pkg/aiops/bootstrap"
	"github.com/lihaiya/aiops/pkg/aiops/metrics"
	"github.com/lihaiya/aiops/pkg/chat"
	"github.com/lihaiya/aiops/pkg/runtime"
)

func main() {
	os.Setenv("AIOPS_SERVICE", "chat")
	bootstrap.MustInit()
	runtime.Run("chat", 8082, func(r *gin.Engine) {
		chat.Register(r, config.GVA_DB)
		// Prometheus 指标端点（REQ-102）
		metrics.RegisterMetricsEndpoint(r)
	})
}
