package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lihaiya/aiops/pkg/runtime"
)

// PostJSON 集群内 POST，透传 gateway 身份头。
func PostJSON(ctx context.Context, url string, body interface{}, ginCtx *gin.Context, out interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	copyHeaders(req, ginCtx)
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("upstream %s: %s", resp.Status, string(raw))
	}
	if out != nil {
		return json.Unmarshal(raw, out)
	}
	return nil
}

// GetJSON GET 请求。
func GetJSON(ctx context.Context, url string, ginCtx *gin.Context, out interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	copyHeaders(req, ginCtx)
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("upstream %s: %s", resp.Status, string(raw))
	}
	if out != nil {
		return json.Unmarshal(raw, out)
	}
	return nil
}

func copyHeaders(req *http.Request, c *gin.Context) {
	if c == nil {
		return
	}
	for _, h := range []string{runtime.HeaderRequestID, runtime.HeaderUserID, runtime.HeaderRole, runtime.HeaderEnvironment, "Authorization"} {
		if v := c.GetHeader(h); v != "" {
			req.Header.Set(h, v)
		}
	}
}
