package tenant

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lihaiya/aiops/pkg/runtime"
	"gorm.io/gorm"
)

// EnvironmentScope 返回 GORM scope，按 environment_slug 过滤（REQ-103）。
// 从 gin.Context 中的 X-AIOps-Environment 头获取环境标识。
func EnvironmentScope(c *gin.Context) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		envSlug := c.GetHeader(runtime.HeaderEnvironment)
		if envSlug == "" {
			return db
		}
		return db.Where("environment_slug = ?", envSlug)
	}
}

// EnvironmentScopeWithValue 返回指定 environment_slug 的 GORM scope。
func EnvironmentScopeWithValue(envSlug string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if envSlug == "" {
			return db
		}
		return db.Where("environment_slug = ?", envSlug)
	}
}

// GetEnvironmentSlug 从 gin.Context 获取当前环境标识（REQ-103）。
func GetEnvironmentSlug(c *gin.Context) string {
	return c.GetHeader(runtime.HeaderEnvironment)
}

// RequireEnvironment 中间件：强制要求 X-AIOps-Environment 头（REQ-103）。
func RequireEnvironment() gin.HandlerFunc {
	return func(c *gin.Context) {
		env := c.GetHeader(runtime.HeaderEnvironment)
		if env == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"code":       400,
				"message":    "缺少 X-AIOps-Environment 头",
				"request_id": runtime.RequestIDFromContext(c),
			})
			return
		}
		c.Next()
	}
}

// FillEnvironmentSlug 在写入时自动填充 environment_slug（REQ-103）。
// 以请求头 X-AIOps-Environment 为准，不信任 body 中的值。
func FillEnvironmentSlug(c *gin.Context) string {
	return c.GetHeader(runtime.HeaderEnvironment)
}
