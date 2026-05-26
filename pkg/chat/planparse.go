package chat

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/lihaiya/aiops/pkg/aiops/types"
)

var jsonBlockRe = regexp.MustCompile("(?s)```json\\s*(\\[.*?\\])\\s*```")

// ParseActionPlans 从 LLM 回复提取 ActionPlan 数组。
func ParseActionPlans(content string) []types.ActionPlan {
	content = strings.TrimSpace(content)
	if block := jsonBlockRe.FindStringSubmatch(content); len(block) > 1 {
		var plans []types.ActionPlan
		if json.Unmarshal([]byte(block[1]), &plans) == nil {
			return plans
		}
	}
	// 尝试裸 JSON 数组
	idx := strings.Index(content, "[")
	if idx >= 0 {
		end := strings.LastIndex(content, "]")
		if end > idx {
			var plans []types.ActionPlan
			if json.Unmarshal([]byte(content[idx:end+1]), &plans) == nil {
				return plans
			}
		}
	}
	return nil
}

const SystemPromptOps = `你是 AIOps 运维助手。禁止把自然语言命令直接当 Shell 执行。
若用户需要运维操作，在回复末尾用 JSON 代码块输出 ActionPlan 数组，格式：
` + "```json\n[{\"plugin\":\"kubernetes\",\"action\":\"list\",\"parameters\":{\"namespace\":\"default\",\"resource\":\"pods\"},\"risk\":\"read\",\"summary\":\"...\",\"commandPreview\":\"kubectl get pods -n default\"}]\n```" + `
risk 取值 read|write|notify。plugin 可用 kubernetes 或 mock。无运维动作时不要输出 ActionPlan 块。`

const SystemPromptAssist = `你是故障修复辅助助手。只输出排查建议与步骤，不要生成可自动执行的 ActionPlan JSON。`

const SystemPromptRCA = `你是 RCA 分析助手。根据提供的审计与日志摘要，输出事件时间线与可能根因。不要生成 ActionPlan。`

const SystemPromptHelpdesk = `你是 IT HelpDesk。仅根据知识库片段回答办公常见问题。不要生成 ActionPlan 或 kubectl 命令。`
