package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/lihaiya/aiops/config"
	"github.com/lihaiya/aiops/initialize"
	"github.com/lihaiya/aiops/pkg/aiops/bootstrap"
	"github.com/lihaiya/aiops/pkg/platform"
	"github.com/lihaiya/aiops/pkg/runtime"
)

func main() {
	os.Setenv("AIOPS_SERVICE", "platform")
	bootstrap.MustInit()
	runtime.Run("platform", 8081, func(r *gin.Engine) {
		platform.Register(r, config.GVA_DB)
		platform.RegisterExtra(r, config.GVA_DB)
		if config.GVA_CONFIG.Admin.Enable {
			initialize.Admin(r)
		}
	})
}
