package audit

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lihaiya/aiops/pkg/aiops/httpclient"
	"github.com/lihaiya/aiops/pkg/aiops/types"
	"github.com/lihaiya/aiops/pkg/runtime"
)

// Client 向 platform 写审计（失败不挡主流程）。
type Client struct {
	PlatformURL string
}

func NewClient() *Client {
	return &Client{PlatformURL: runtime.EnvOr("AIOPS_PLATFORM_URL", "http://aiops-platform:8081")}
}

func (c *Client) Log(ctx context.Context, ginCtx *gin.Context, req types.AuditRequest) {
	url := c.PlatformURL + "/internal/v1/audit"
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := httpclient.PostJSON(ctx, url, req, ginCtx, nil); err != nil {
		_ = json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
			"level":      "warn",
			"msg":        "audit_failed",
			"error":      err.Error(),
			"event_type": req.EventType,
		})
	}
}
