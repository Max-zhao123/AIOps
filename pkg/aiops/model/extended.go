package model

import basemodel "github.com/lihaiya/aiops/pkg/model"

type KbDocument struct {
	basemodel.BaseModel
	Title    string `json:"title" gorm:"column:title;size:256;not null"`
	Content  string `json:"content" gorm:"column:content;type:longtext"`
	Tags     string `json:"tags" gorm:"column:tags;size:512"`
	Source   string `json:"source" gorm:"column:source;size:128"`
}

func (KbDocument) TableName() string { return "kb_documents" }

type InspectionTask struct {
	basemodel.BaseModel
	Name        string `json:"name" gorm:"column:name;size:128;not null"`
	Environment string `json:"environment" gorm:"column:environment;size:64;index"`
	CronExpr    string `json:"cronExpr" gorm:"column:cron_expr;size:64"`
	StepsJSON   string `json:"stepsJson" gorm:"column:steps_json;type:text"`
	Enabled     bool   `json:"enabled" gorm:"column:enabled;default:true"`
}

func (InspectionTask) TableName() string { return "inspection_tasks" }

type InspectionReport struct {
	basemodel.BaseModel
	TaskID    int64  `json:"taskId" gorm:"column:task_id;index"`
	Status    string `json:"status" gorm:"column:status;size:32"`
	Summary   string `json:"summary" gorm:"column:summary;type:text"`
	DetailJSON string `json:"detailJson" gorm:"column:detail_json;type:longtext"`
}

func (InspectionReport) TableName() string { return "inspection_reports" }

type WorkerJob struct {
	basemodel.BaseModel
	JobType     string `json:"jobType" gorm:"column:job_type;size:64;index"`
	PayloadJSON string `json:"payloadJson" gorm:"column:payload_json;type:text"`
	Status      string `json:"status" gorm:"column:status;size:32;index"`
	NextRunAt   int64  `json:"nextRunAt" gorm:"column:next_run_at"`
}

func (WorkerJob) TableName() string { return "worker_jobs" }

type NotifyRecord struct {
	basemodel.BaseModel
	Source      string `json:"source" gorm:"column:source;size:64"`
	Environment string `json:"environment" gorm:"column:environment;size:64"`
	Level       string `json:"level" gorm:"column:level;size:16"`
	Message     string `json:"message" gorm:"column:message;type:text"`
}

func (NotifyRecord) TableName() string { return "notify_records" }

type Runbook struct {
	basemodel.BaseModel
	Name        string `json:"name" gorm:"column:name;size:128"`
	Environment string `json:"environment" gorm:"column:environment;size:64"`
	StepsJSON   string `json:"stepsJson" gorm:"column:steps_json;type:text"`
}

func (Runbook) TableName() string { return "runbooks" }

type RiskAlert struct {
	basemodel.BaseModel
	Environment string `json:"environment" gorm:"column:environment;size:64;index"`
	PlanJSON    string `json:"planJson" gorm:"column:plan_json;type:text"`
	Reason      string `json:"reason" gorm:"column:reason;type:text"`
	Status      string `json:"status" gorm:"column:status;size:32;default:open"`
}

func (RiskAlert) TableName() string { return "risk_alerts" }

type ImMessage struct {
	basemodel.BaseModel
	Channel   string `json:"channel" gorm:"column:channel;size:32"`
	UserKey   string `json:"userKey" gorm:"column:user_key;size:128"`
	Content   string `json:"content" gorm:"column:content;type:text"`
	ReplyJSON string `json:"replyJson" gorm:"column:reply_json;type:text"`
}

func (ImMessage) TableName() string { return "im_messages" }
