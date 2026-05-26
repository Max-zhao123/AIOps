package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/lihaiya/aiops/config"
	"github.com/lihaiya/aiops/pkg/aiops/bootstrap"
	"github.com/lihaiya/aiops/pkg/executor"
	"github.com/lihaiya/aiops/pkg/runtime"
)

func main() {
	os.Setenv("AIOPS_SERVICE", "executor")
	bootstrap.MustInit()
	runtime.Run("executor", 8084, func(r *gin.Engine) {
		executor.Register(r, config.GVA_DB)
	})
}
