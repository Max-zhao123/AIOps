package identity

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lihaiya/aiops/pkg/runtime"
)

func UserID(c *gin.Context) int64 {
	v := c.GetHeader(runtime.HeaderUserID)
	if v == "" {
		return 0
	}
	id, _ := strconv.ParseInt(v, 10, 64)
	return id
}

func Role(c *gin.Context) string {
	return c.GetHeader(runtime.HeaderRole)
}

func Environment(c *gin.Context) string {
	return c.GetHeader(runtime.HeaderEnvironment)
}
