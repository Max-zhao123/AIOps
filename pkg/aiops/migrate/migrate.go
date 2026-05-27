package migrate

import (
	"os"

	"github.com/lihaiya/aiops/pkg/aiops/model"
	"github.com/lihaiya/aiops/pkg/aiops/seed"
	"github.com/lihaiya/aiops/pkg/auth"
	"gorm.io/gorm"
)

// RegisterTables 阶段 A+E 控制面表。
func RegisterTables(db *gorm.DB) {
	err := db.Set("gorm:table_options", "CHARSET=utf8mb4").AutoMigrate(
		// Phase A
		model.Environment{},
		model.User{},
		model.AuditLog{},
		model.SecurityBoundaryPolicy{},
		model.ExecutionRecord{},
		model.ChatSession{},
		model.ChatMessage{},
		model.KbDocument{},
		model.InspectionTask{},
		model.InspectionReport{},
		model.WorkerJob{},
		model.NotifyRecord{},
		model.Runbook{},
		model.RiskAlert{},
		model.ImMessage{},
		model.LlmConfig{},
		auth.BaseUser{},
		// Phase E
		model.Schedule{},
		model.ScheduleExecution{},
		model.NotificationChannel{},
		model.NotificationPolicy{},
		model.NotificationTemplate{},
		model.Credential{},
		model.ApprovalPolicy{},
		model.SlaDefinition{},
		model.SlaRecord{},
		model.EnvironmentConfig{},
		model.ConfigVersion{},
		model.EnvironmentQuota{},
	)
	if err != nil {
		os.Exit(1)
	}
	seed.Run(db)
}
