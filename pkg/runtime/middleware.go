package runtime

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestContext 解析或生成 X-Request-Id，并写入 context。
func RequestContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader(HeaderRequestID)
		if rid == "" {
			rid = NewRequestID()
			c.Request.Header.Set(HeaderRequestID, rid)
		}
		c.Set(ContextRequestID, rid)
		c.Writer.Header().Set(HeaderRequestID, rid)

		if env := c.GetHeader(HeaderEnvironment); env != "" {
			c.Set(ContextEnvironment, env)
		}
		c.Next()
	}
}

// Recovery panic 恢复。
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				_ = json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
					"level":      "error",
					"service":    os.Getenv("AIOPS_SERVICE"),
					"request_id": RequestIDFromContext(c),
					"msg":        fmt.Sprint(err),
					"stack":      string(debug.Stack()),
				})
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}

// AccessLog 结构化访问日志（PRD §3.11）。
func AccessLog(service string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		entry := map[string]interface{}{
			"level":      "info",
			"service":    service,
			"request_id": RequestIDFromContext(c),
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"status":     c.Writer.Status(),
			"latency_ms": time.Since(start).Milliseconds(),
		}
		if env, ok := c.Get(ContextEnvironment); ok {
			entry["environment"] = env
		}
		_ = json.NewEncoder(os.Stdout).Encode(entry)
	}
}
