package main

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/lihaiya/aiops/pkg/aiops/httpclient"
	"github.com/lihaiya/aiops/pkg/runtime"
)

func main() {
	os.Setenv("AIOPS_SERVICE", "worker")
	platformURL := runtime.EnvOr("AIOPS_PLATFORM_URL", "http://aiops-platform:8081")

	go func() {
		ticker := time.NewTicker(15 * time.Second)
		for range ticker.C {
			ctx := context.Background()
			var resp struct {
				Items []struct {
					ID int64 `json:"id"`
				} `json:"items"`
			}
			if err := httpclient.GetJSON(ctx, platformURL+"/internal/v1/worker/jobs/pending", nil, &resp); err != nil {
				continue
			}
			for _, job := range resp.Items {
				processURL := platformURL + "/internal/v1/worker/jobs/" + strconv.FormatInt(job.ID, 10) + "/process"
				if err := httpclient.PostJSON(ctx, processURL, map[string]interface{}{}, nil, nil); err != nil {
					continue
				}
				completeURL := platformURL + "/internal/v1/worker/jobs/" + strconv.FormatInt(job.ID, 10) + "/complete"
				_ = httpclient.PostJSON(ctx, completeURL, map[string]interface{}{"ok": true}, nil, nil)
			}
		}
	}()

	runtime.Run("worker", 8086, nil)
}
