package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/lihaiya/aiops/pkg/aiops/bootstrap"
	"github.com/lihaiya/aiops/pkg/gateway"
	"github.com/lihaiya/aiops/pkg/runtime"
	"github.com/lihaiya/aiops/config"
)

func main() {
	os.Setenv("AIOPS_SERVICE", "gateway")
	bootstrap.MustInit()
	runtime.Run("gateway", 8080, registerGateway)
}

func registerGateway(r *gin.Engine) {
	r.Use(gateway.JWTAuth(config.GVA_DB))
	platform := runtime.EnvOr("AIOPS_PLATFORM_URL", "http://aiops-platform:8081")
	chat := runtime.EnvOr("AIOPS_CHAT_URL", "http://aiops-chat:8082")
	executor := runtime.EnvOr("AIOPS_EXECUTOR_URL", "http://aiops-executor:8084")

	runtime.MountReverseProxy(r, "/api/v1/auth", platform)
	runtime.MountReverseProxy(r, "/api/v1/environments", platform)
	runtime.MountReverseProxy(r, "/api/v1/security-boundaries", platform)
	runtime.MountReverseProxy(r, "/api/v1/plugins", platform)
	runtime.MountReverseProxy(r, "/api/v1/audit", platform)
	runtime.MountReverseProxy(r, "/api/v1/chat", chat)
	runtime.MountReverseProxy(r, "/api/v1/actions", executor)
	runtime.MountReverseProxy(r, "/api/v1/data", platform)
	runtime.MountReverseProxy(r, "/api/v1/inspections", platform)
	runtime.MountReverseProxy(r, "/api/v1/risk-alerts", platform)
	runtime.MountReverseProxy(r, "/api/v1/im", platform)
	runtime.MountReverseProxy(r, "/api/v1/helpdesk", platform)
	runtime.MountReverseProxy(r, "/api/v1/runbooks", platform)
	runtime.MountReverseProxy(r, "/api/v1/kb", runtime.EnvOr("AIOPS_MODULE_KB_URL", "http://aiops-module-kb:8085"))
	runtime.MountReverseProxy(r, "/admin", platform)
}
