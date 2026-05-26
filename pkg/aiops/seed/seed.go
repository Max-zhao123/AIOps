package seed

import (
	"github.com/lihaiya/aiops/pkg/aiops/model"
	"github.com/lihaiya/aiops/pkg/aiops/types"
	"github.com/lihaiya/aiops/utils"
	"gorm.io/gorm"
)

// Run 初始化环境、管理员与策略模板（幂等）。
func Run(db *gorm.DB) {
	envs := []model.Environment{
		{Slug: "development", Name: "开发环境", Description: "开发测试"},
		{Slug: "staging", Name: "预发环境", Description: "预发布验证"},
		{Slug: "production", Name: "生产环境", Description: "生产运维"},
	}
	for _, e := range envs {
		var cnt int64
		db.Model(&model.Environment{}).Where("slug = ?", e.Slug).Count(&cnt)
		if cnt == 0 {
			db.Create(&e)
		}
	}

	var userCnt int64
	db.Model(&model.User{}).Where("username = ?", "admin").Count(&userCnt)
	if userCnt == 0 {
		db.Create(&model.User{
			Username: "admin",
			Password: utils.BcryptHash("admin@123"),
			Role:     types.RoleAdmin,
		})
	}

	var prod model.Environment
	if err := db.Where("slug = ?", "production").First(&prod).Error; err != nil {
		return
	}
	templates := []struct {
		name string
		spec string
	}{
		{"prod-readonly", prodReadonlySpec},
		{"prod-standard", prodStandardSpec},
		{"dev-relaxed", devRelaxedSpec},
	}
	var kbCnt int64
	db.Model(&model.KbDocument{}).Count(&kbCnt)
	if kbCnt == 0 {
		db.Create(&model.KbDocument{Title: "加域", Content: "Windows 加域：设置-账户-访问工作单位或学校", Tags: "helpdesk"})
		db.Create(&model.KbDocument{Title: "邮箱申请", Content: "通过 IT 门户提交邮箱开通申请", Tags: "helpdesk"})
	}

	for _, t := range templates {
		var cnt int64
		db.Model(&model.SecurityBoundaryPolicy{}).Where("name = ?", t.name).Count(&cnt)
		if cnt == 0 {
			db.Create(&model.SecurityBoundaryPolicy{
				EnvironmentID: prod.ID,
				Name:          t.name,
				SpecYAML:      t.spec,
				Enabled:       false,
			})
		}
	}
}

const prodReadonlySpec = `defaultDecision: ASK
rules:
  - id: read-allow
    match: { plugin: "", actions: ["get","list","describe","logs"], commandPattern: "" }
    decision: ALLOW
    message: read allowed
  - id: write-deny
    match: { plugin: "", actions: ["delete","apply","patch","create"], commandPattern: "" }
    decision: DENY
    message: write denied in prod-readonly
`

const prodStandardSpec = `defaultDecision: ASK
rules:
  - id: read-allow
    match: { plugin: "", actions: ["get","list","describe","logs"], commandPattern: "" }
    decision: ALLOW
    message: read allowed
  - id: write-ask
    match: { plugin: "", actions: ["delete","apply","patch","create"], commandPattern: "" }
    decision: ASK
    message: write needs confirm
`

const devRelaxedSpec = `defaultDecision: ASK
rules:
  - id: read-allow
    match: { plugin: "", actions: ["get","list","describe","logs"], commandPattern: "" }
    decision: ALLOW
    message: read allowed
  - id: write-ask
    match: { plugin: "", actions: ["delete","apply","patch","create"], commandPattern: "" }
    decision: ASK
    message: write needs confirm
`
