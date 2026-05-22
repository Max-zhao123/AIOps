package initialize

import (
	"github.com/gin-gonic/gin"
	"github.com/lihaiya/aiops/config"
	"github.com/lihaiya/aiops/internal/app"
	"github.com/lihaiya/aiops/pkg/admin"
)

func Admin(r *gin.Engine) {
	if !config.GVA_CONFIG.Admin.Enable {
		return
	}
	admin.Init(r, nil)
	app.Admin()
}
