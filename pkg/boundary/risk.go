package boundary

import (
	"regexp"
	"strings"

	"github.com/lihaiya/aiops/pkg/aiops/types"
)

var hardDenyPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)rm\s+-rf`),
	regexp.MustCompile(`(?i)drop\s+database`),
	regexp.MustCompile(`(?i)kubectl\s+delete`),
	regexp.MustCompile(`(?i)truncate\s+table`),
}

// HardDeny 硬禁止扫描（PRD §3.6）。
func HardDeny(plan types.ActionPlan) (bool, string) {
	text := plan.CommandPreview + " " + plan.Summary
	for _, re := range hardDenyPatterns {
		if re.MatchString(text) {
			return true, "hard deny: matched " + re.String()
		}
	}
	if strings.Contains(strings.ToLower(plan.Action), "delete") && plan.Plugin == "kubernetes" {
		return true, "hard deny: kubernetes delete"
	}
	return false, ""
}
