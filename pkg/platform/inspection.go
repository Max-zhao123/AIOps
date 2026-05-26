package platform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	aimodel "github.com/lihaiya/aiops/pkg/aiops/model"
	"github.com/lihaiya/aiops/pkg/aiops/types"
	"github.com/lihaiya/aiops/pkg/aiops/httpclient"
	"github.com/lihaiya/aiops/pkg/runtime"
)

type inspectionStepResult struct {
	Plan   types.ActionPlan       `json:"plan"`
	Result map[string]interface{} `json:"result"`
	Error  string                 `json:"error,omitempty"`
}

// runInspectionSteps 执行巡检步骤（只读 ActionPlan 经 executor）。
func (h *Handler) runInspectionSteps(ctx context.Context, task aimodel.InspectionTask, ginCtx *gin.Context) (aimodel.InspectionReport, error) {
	var steps []types.ActionPlan
	if task.StepsJSON != "" {
		if err := json.Unmarshal([]byte(task.StepsJSON), &steps); err != nil {
			return aimodel.InspectionReport{}, err
		}
	}
	execURL := runtime.EnvOr("AIOPS_EXECUTOR_URL", "http://aiops-executor:8084")
	var results []inspectionStepResult
	failed := false
	for _, plan := range steps {
		if plan.Risk == "" {
			plan.Risk = "read"
		}
		var out map[string]interface{}
		err := httpclient.PostJSON(ctx, execURL+"/internal/v1/run", types.RunRequest{
			Environment: task.Environment,
			Plan:        plan,
		}, ginCtx, &out)
		sr := inspectionStepResult{Plan: plan, Result: out}
		if err != nil {
			sr.Error = err.Error()
			failed = true
		} else if st, ok := out["status"].(string); ok && st == "failed" {
			failed = true
			sr.Error = "execution failed"
		}
		results = append(results, sr)
	}
	detail, _ := json.Marshal(results)
	status := "ok"
	summary := fmt.Sprintf("inspection %s: %d steps", task.Name, len(steps))
	if failed {
		status = "failed"
		summary = fmt.Sprintf("inspection %s failed", task.Name)
		h.NotifyFailure(task.Environment, summary)
	}
	report := aimodel.InspectionReport{
		TaskID:     task.ID,
		Status:     status,
		Summary:    summary,
		DetailJSON: string(detail),
	}
	if err := h.DB.Create(&report).Error; err != nil {
		return report, err
	}
	return report, nil
}

// ProcessWorkerJob 处理 worker 任务（REQ-060/061/062）。
func (h *Handler) ProcessWorkerJob(ctx context.Context, jobID int64) error {
	var job aimodel.WorkerJob
	if err := h.DB.First(&job, jobID).Error; err != nil {
		return err
	}
	switch job.JobType {
	case "inspection":
		var payload struct {
			TaskID int64 `json:"taskId"`
		}
		_ = json.Unmarshal([]byte(job.PayloadJSON), &payload)
		var task aimodel.InspectionTask
		if err := h.DB.First(&task, payload.TaskID).Error; err != nil {
			return err
		}
		_, err := h.runInspectionSteps(ctx, task, nil)
		return err
	default:
		return fmt.Errorf("unknown job type: %s", job.JobType)
	}
}
