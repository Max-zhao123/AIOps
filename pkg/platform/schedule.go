package platform

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	aimodel "github.com/lihaiya/aiops/pkg/aiops/model"
	"github.com/lihaiya/aiops/pkg/aiops/tenant"
	"github.com/lihaiya/aiops/pkg/runtime"
	"gorm.io/gorm"
)

// -------- 定时调度（REQ-090） --------

func (h *Handler) ListSchedules(c *gin.Context) {
	envSlug := tenant.GetEnvironmentSlug(c)
	var list []aimodel.Schedule
	q := h.DB.Where("environment_slug = ?", envSlug).Order("id asc")
	if enabled := c.Query("enabled"); enabled != "" {
		q = q.Where("enabled = ?", enabled == "true")
	}
	q.Find(&list)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": list})
}

func (h *Handler) CreateSchedule(c *gin.Context) {
	var body struct {
		Name       string `json:"name" binding:"required,max=100"`
		Cron       string `json:"cron" binding:"required,max=50"`
		Timezone   string `json:"timezone" binding:"max=50"`
		TargetType string `json:"targetType" binding:"required,oneof=inspection runbook"`
		TargetID   int64  `json:"targetId" binding:"required"`
		Enabled    *bool  `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error(), "request_id": runtime.RequestIDFromContext(c)})
		return
	}

	// 校验 cron 表达式
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(body.Cron)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的 cron 表达式: " + err.Error(), "request_id": runtime.RequestIDFromContext(c)})
		return
	}

	envSlug := tenant.FillEnvironmentSlug(c)
	timezone := body.Timezone
	if timezone == "" {
		timezone = "UTC"
	}

	enabled := false
	if body.Enabled != nil {
		enabled = *body.Enabled
	}

	sched := aimodel.Schedule{
		Name:            body.Name,
		Cron:            body.Cron,
		Timezone:        timezone,
		TargetType:      body.TargetType,
		TargetID:        body.TargetID,
		Enabled:         enabled,
		EnvironmentSlug: envSlug,
	}

	// 计算下次执行时间
	loc, locErr := time.LoadLocation(timezone)
	if locErr != nil {
		loc = time.UTC
	}
	next := schedule.Next(time.Now().In(loc))
	sched.NextRunAt = &next

	if err := h.DB.Create(&sched).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "调度创建失败", "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": 0, "data": sched})
}

func (h *Handler) GetSchedule(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	envSlug := tenant.GetEnvironmentSlug(c)
	var sched aimodel.Schedule
	if err := h.DB.Where("id = ? AND environment_slug = ?", id, envSlug).First(&sched).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "调度不存在", "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": sched})
}

func (h *Handler) UpdateSchedule(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	envSlug := tenant.GetEnvironmentSlug(c)
	var sched aimodel.Schedule
	if err := h.DB.Where("id = ? AND environment_slug = ?", id, envSlug).First(&sched).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "调度不存在", "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	var body struct {
		Name       string `json:"name" binding:"max=100"`
		Cron       string `json:"cron" binding:"max=50"`
		Timezone   string `json:"timezone" binding:"max=50"`
		TargetType string `json:"targetType" binding:"omitempty,oneof=inspection runbook"`
		TargetID   *int64 `json:"targetId"`
		Enabled    *bool  `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error(), "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	if body.Name != "" {
		sched.Name = body.Name
	}
	if body.Cron != "" {
		parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
		if _, err := parser.Parse(body.Cron); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的 cron 表达式", "request_id": runtime.RequestIDFromContext(c)})
			return
		}
		sched.Cron = body.Cron
	}
	if body.Timezone != "" {
		sched.Timezone = body.Timezone
	}
	if body.TargetType != "" {
		sched.TargetType = body.TargetType
	}
	if body.TargetID != nil {
		sched.TargetID = *body.TargetID
	}
	if body.Enabled != nil {
		sched.Enabled = *body.Enabled
	}
	// 重新计算下次执行时间
	next := computeNextRun(sched.Cron, sched.Timezone)
	sched.NextRunAt = next
	h.DB.Save(&sched)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": sched})
}

func (h *Handler) DeleteSchedule(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	envSlug := tenant.GetEnvironmentSlug(c)
	result := h.DB.Where("id = ? AND environment_slug = ?", id, envSlug).Delete(&aimodel.Schedule{})
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "调度不存在", "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0})
}

func (h *Handler) ListScheduleExecutions(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var list []aimodel.ScheduleExecution
	h.DB.Where("schedule_id = ?", id).Order("id desc").Limit(50).Find(&list)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": list})
}

// ToggleSchedule 切换调度启用/禁用状态。
func (h *Handler) ToggleSchedule(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	envSlug := tenant.GetEnvironmentSlug(c)
	var sched aimodel.Schedule
	if err := h.DB.Where("id = ? AND environment_slug = ?", id, envSlug).First(&sched).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "调度不存在", "request_id": runtime.RequestIDFromContext(c)})
		return
	}
	sched.Enabled = !sched.Enabled
	// 如果启用，重新计算下次执行时间
	if sched.Enabled {
		if next := computeNextRun(sched.Cron, sched.Timezone); next != nil {
			sched.NextRunAt = next
		}
	}
	h.DB.Save(&sched)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": sched})
}

// TriggerSchedule 手动触发调度执行。
func (h *Handler) TriggerSchedule(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	envSlug := tenant.GetEnvironmentSlug(c)
	var sched aimodel.Schedule
	if err := h.DB.Where("id = ? AND environment_slug = ?", id, envSlug).First(&sched).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "调度不存在", "request_id": runtime.RequestIDFromContext(c)})
		return
	}

	now := time.Now()
	exec := aimodel.ScheduleExecution{
		ScheduleID: sched.ID,
		Status:     "running",
		StartedAt:  now,
	}
	h.DB.Create(&exec)

	// 更新调度最后执行时间
	sched.LastRunAt = &now
	if sched.Enabled {
		if next := computeNextRun(sched.Cron, sched.Timezone); next != nil {
			sched.NextRunAt = next
		}
	}
	h.DB.Save(&sched)

	// 执行任务
	var resultSummary string
	var errMsg string
	var status string

	switch sched.TargetType {
	case "inspection":
		var task aimodel.InspectionTask
		if err := h.DB.First(&task, sched.TargetID).Error; err != nil {
			errMsg = "巡检任务不存在"
			status = "failed"
		} else {
			report, err := h.runInspectionSteps(nil, task, nil)
			if err != nil {
				errMsg = err.Error()
				status = "failed"
			} else {
				resultSummary = report.Summary
				status = "completed"
			}
		}
	case "runbook":
		var rb aimodel.Runbook
		if err := h.DB.First(&rb, sched.TargetID).Error; err != nil {
			errMsg = "Runbook 不存在"
			status = "failed"
		} else {
			resultSummary = "runbook " + rb.Name + " executed"
			status = "completed"
		}
	default:
		errMsg = "未知目标类型: " + sched.TargetType
		status = "failed"
	}

	completedAt := time.Now()
	exec.Status = status
	exec.CompletedAt = &completedAt
	exec.ResultSummary = resultSummary
	exec.ErrorMessage = errMsg
	h.DB.Save(&exec)

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": exec})
}

// -------- 调度执行逻辑（REQ-090） --------

// RunSchedule 执行定时调度任务（worker 内部调用）。
func (h *Handler) RunSchedule(ctx interface{}, scheduleID int64) error {
	var sched aimodel.Schedule
	if err := h.DB.First(&sched, scheduleID).Error; err != nil {
		return err
	}

	now := time.Now()
	exec := aimodel.ScheduleExecution{
		ScheduleID: sched.ID,
		Status:     "running",
		StartedAt:  now,
	}
	h.DB.Create(&exec)

	// 更新调度最后执行时间
	sched.LastRunAt = &now
	if next := computeNextRun(sched.Cron, sched.Timezone); next != nil {
		sched.NextRunAt = next
	}
	h.DB.Save(&sched)

	// 执行任务
	var resultSummary string
	var errMsg string
	var status string

	switch sched.TargetType {
	case "inspection":
		var task aimodel.InspectionTask
		if err := h.DB.First(&task, sched.TargetID).Error; err != nil {
			errMsg = "巡检任务不存在"
			status = "failed"
		} else {
			report, err := h.runInspectionSteps(nil, task, nil)
			if err != nil {
				errMsg = err.Error()
				status = "failed"
			} else {
				resultSummary = report.Summary
				status = "completed"
			}
		}
	case "runbook":
		var rb aimodel.Runbook
		if err := h.DB.First(&rb, sched.TargetID).Error; err != nil {
			errMsg = "Runbook 不存在"
			status = "failed"
		} else {
			resultSummary = "runbook " + rb.Name + " executed"
			status = "completed"
		}
	default:
		errMsg = "未知目标类型: " + sched.TargetType
		status = "failed"
	}

	completedAt := time.Now()
	exec.Status = status
	exec.CompletedAt = &completedAt
	exec.ResultSummary = resultSummary
	exec.ErrorMessage = errMsg
	h.DB.Save(&exec)

	return nil
}

// computeNextRun 计算下次执行时间。
func computeNextRun(cronExpr, timezone string) *time.Time {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(cronExpr)
	if err != nil {
		return nil
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}
	next := schedule.Next(time.Now().In(loc))
	return &next
}

// ScheduleExecutionWorker 后台调度器（REQ-090）。
// 启动后每 30 秒检查是否有需要执行的定时任务。
func ScheduleExecutionWorker(db *gorm.DB) {
	ticker := time.NewTicker(30 * time.Second)
	for range ticker.C {
		var schedules []aimodel.Schedule
		db.Where("enabled = ? AND next_run_at <= ?", true, time.Now()).Find(&schedules)
		for _, sched := range schedules {
			h := &Handler{DB: db}
			_ = h.RunSchedule(nil, sched.ID)
		}
	}
}
