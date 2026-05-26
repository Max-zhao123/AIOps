package bootstrap

import (
	"github.com/lihaiya/aiops/config"
	"github.com/lihaiya/aiops/initialize"
	"go.uber.org/zap"
)

// MustInit 加载配置与 MySQL（platform/policy 等需要 DB 的服务调用）。
func MustInit() {
	if config.GVA_VP == nil {
		config.GVA_VP = initialize.Viper()
	}
	if config.GVA_LOG == nil {
		config.GVA_LOG = initialize.Zap()
		zap.ReplaceGlobals(config.GVA_LOG)
	}
	if config.GVA_DB == nil {
		config.GVA_DB = initialize.Gorm()
	}
}
