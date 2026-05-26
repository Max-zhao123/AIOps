package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/lihaiya/aiops/config"
	"github.com/lihaiya/aiops/pkg/aiops/bootstrap"
	"github.com/lihaiya/aiops/pkg/module/kb"
	"github.com/lihaiya/aiops/pkg/runtime"
)

func main() {
	os.Setenv("AIOPS_SERVICE", "module-kb")
	bootstrap.MustInit()
	runtime.Run("module-kb", 8085, func(r *gin.Engine) {
		kb.Register(r, config.GVA_DB)
	})
}
