package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/lihaiya/aiops/config"
	"github.com/lihaiya/aiops/pkg/aiops/bootstrap"
	"github.com/lihaiya/aiops/pkg/policy"
	"github.com/lihaiya/aiops/pkg/runtime"
)

func main() {
	os.Setenv("AIOPS_SERVICE", "policy")
	bootstrap.MustInit()
	runtime.Run("policy", 8083, func(r *gin.Engine) {
		policy.Register(r, config.GVA_DB)
	})
}
