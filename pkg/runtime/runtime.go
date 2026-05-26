package runtime

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)

// Run 启动 HTTP 服务：/healthz、/readyz + register 自定义路由。
func Run(serviceName string, defaultPort int, register func(*gin.Engine)) {
	port := defaultPort
	if v := os.Getenv("PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			port = p
		}
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(Recovery(), RequestContext(), AccessLog(serviceName))
	r.GET("/healthz", healthHandler(serviceName))
	r.GET("/readyz", healthHandler(serviceName))
	if register != nil {
		register(r)
	}

	addr := fmt.Sprintf(":%d", port)
	s := endless.NewServer(addr, r)
	s.ReadHeaderTimeout = 20 * time.Second
	s.WriteTimeout = 20 * time.Second
	s.MaxHeaderBytes = 1 << 20
	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}

func healthHandler(service string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": service,
		})
	}
}

// RequestIDFromContext 返回当前请求的 X-Request-Id。
func RequestIDFromContext(c *gin.Context) string {
	if v, ok := c.Get(ContextRequestID); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// NewRequestID 生成新的请求 ID。
func NewRequestID() string {
	return uuid.NewV4().String()
}
