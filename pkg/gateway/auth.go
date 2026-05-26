package gateway

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lihaiya/aiops/pkg/aiops/model"
	"github.com/lihaiya/aiops/pkg/aiops/types"
	"github.com/lihaiya/aiops/pkg/runtime"
	"github.com/lihaiya/aiops/utils"
	"gorm.io/gorm"
)

// JWTAuth 解析 Bearer JWT，设置 X-AIOps-* 头。
func JWTAuth(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if skipAuth(c.Request.URL.Path) {
			c.Next()
			return
		}
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		token := strings.TrimPrefix(auth, "Bearer ")
		j := utils.NewJWT()
		claims, err := j.ParseToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		var user model.User
		if db != nil {
			_ = db.Where("username = ?", claims.Username).First(&user).Error
		}
		role := user.Role
		if role == "" {
			role = types.RoleOperator
		}
		uid := claims.BaseClaims.ID
		if user.ID > 0 {
			uid = user.ID
		}
		c.Request.Header.Set(runtime.HeaderUserID, strconv.FormatInt(uid, 10))
		c.Request.Header.Set(runtime.HeaderRole, role)

		if c.Request.Method == http.MethodPost && strings.HasSuffix(c.Request.URL.Path, "/confirm") {
			if role == types.RoleReadonly {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "readonly cannot confirm"})
				return
			}
		}
		c.Next()
	}
}

func skipAuth(path string) bool {
	if path == "/healthz" || path == "/readyz" {
		return true
	}
	if strings.HasPrefix(path, "/admin/login") || strings.HasPrefix(path, "/admin/register") {
		return true
	}
	if path == "/api/v1/auth/login" {
		return true
	}
	return false
}
