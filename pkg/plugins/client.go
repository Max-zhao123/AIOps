package plugins

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/lihaiya/aiops/pkg/aiops/httpclient"
	"github.com/lihaiya/aiops/pkg/aiops/types"
	"github.com/lihaiya/aiops/pkg/runtime"
)

// Client executor 通过 HTTP 调插件（禁止 in-process import 具体插件实现）。
type Client struct {
	Endpoints map[string]string
}

func NewClient() *Client {
	return &Client{
		Endpoints: map[string]string{
			"mock":         runtime.EnvOr("AIOPS_PLUGIN_MOCK_URL", "http://aiops-plugin-mock:8090"),
			"kubernetes":   runtime.EnvOr("AIOPS_PLUGIN_K8S_URL", "http://aiops-plugin-kubernetes:8091"),
			"prometheus":   runtime.EnvOr("AIOPS_PLUGIN_PROM_URL", "http://aiops-plugin-prometheus:8092"),
			"logs":         runtime.EnvOr("AIOPS_PLUGIN_LOGS_URL", "http://aiops-plugin-logs:8093"),
		},
	}
}

func (c *Client) Execute(ctx context.Context, ginCtx *gin.Context, plan types.ActionPlan) (types.ExecuteResponse, error) {
	ep, ok := c.Endpoints[plan.Plugin]
	if !ok || ep == "" {
		return types.ExecuteResponse{}, fmt.Errorf("plugin %s not configured", plan.Plugin)
	}
	var out types.ExecuteResponse
	url := ep + "/internal/v1/execute"
	err := httpclient.PostJSON(ctx, url, types.ExecuteRequest{Plan: plan}, ginCtx, &out)
	return out, err
}
