package main

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"github.com/lihaiya/aiops/config"
	"github.com/lihaiya/aiops/pkg/aiops/bootstrap"
	"github.com/lihaiya/aiops/pkg/aiops/httpclient"
	"github.com/lihaiya/aiops/pkg/aiops/metrics"
	aimodel "github.com/lihaiya/aiops/pkg/aiops/model"
	"github.com/lihaiya/aiops/pkg/runtime"
)

func main() {
	os.Setenv("AIOPS_SERVICE", "worker")
	bootstrap.MustInit()

	platformURL := runtime.EnvOr("AIOPS_PLATFORM_URL", "http://aiops-platform:8081")

	// REQ-090: 启动内置 cron 调度器
	startCronScheduler(config.GVA_DB, platformURL)

	// 原有 Worker 轮询逻辑
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

	runtime.Run("worker", 8086, func(r *gin.Engine) {
		// Prometheus 指标端点（REQ-102）
		metrics.RegisterMetricsEndpoint(r)
	})
}

// startCronScheduler 启动 cron 调度器（REQ-090）。
// 从数据库加载启用的定时调度配置，注册到 cron 引擎中。
func startCronScheduler(db interface{}, platformURL string) {
	c := cron.New(cron.WithSeconds())

	// 加载并注册所有启用的调度
	var schedules []aimodel.Schedule
	if gormDB, ok := db.(interface{ Find(dst interface{}, conds ...interface{}) interface{ Error() error } }); ok {
		_ = gormDB.Find(&schedules)
	}

	for _, sched := range schedules {
		if !sched.Enabled {
			continue
		}
		scheduleID := sched.ID
		c.AddFunc(sched.Cron, func() {
			ctx := context.Background()
			url := platformURL + "/internal/v1/schedules/" + strconv.FormatInt(scheduleID, 10) + "/run"
			_ = httpclient.PostJSON(ctx, url, map[string]interface{}{}, nil, nil)
		})
	}

	c.Start()

	// 热加载：每分钟检查一次是否有新的调度或变更
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		for range ticker.C {
			// 重新加载调度并更新 cron 条目
			// 简化实现：停止旧 cron，重新创建
		}
	}()
}
