package boundary

import (
	"strings"

	"github.com/lihaiya/aiops/pkg/aiops/types"
)

// Evaluate 策略评估（PRD §3.6）。
func Evaluate(spec types.PolicySpec, plan types.ActionPlan) types.EvaluateResponse {
	if deny, msg := HardDeny(plan); deny {
		return types.EvaluateResponse{
			Decision: types.DecisionDeny,
			Message:  msg,
			HardDeny: true,
		}
	}
	for _, rule := range spec.Rules {
		if !matchRule(plan, rule) {
			continue
		}
		d := normalizeDecision(rule.Decision)
		if d == types.DecisionDeny {
			return types.EvaluateResponse{
				Decision:       types.DecisionDeny,
				Message:        rule.Message,
				MatchedRuleIDs: []string{rule.ID},
			}
		}
		if d == types.DecisionAllow || d == types.DecisionAsk {
			return types.EvaluateResponse{
				Decision:       d,
				Message:        rule.Message,
				MatchedRuleIDs: []string{rule.ID},
			}
		}
	}
	def := normalizeDecision(spec.DefaultDecision)
	if def == "" {
		def = types.DecisionAsk
	}
	return types.EvaluateResponse{Decision: def, Message: "default decision"}
}

// FailClose 评估异常时的 fail-close。
func FailClose(plan types.ActionPlan, err error) types.EvaluateResponse {
	if plan.Risk == "write" {
		return types.EvaluateResponse{Decision: types.DecisionDeny, Message: err.Error()}
	}
	return types.EvaluateResponse{Decision: types.DecisionAsk, Message: err.Error()}
}

func normalizeDecision(d string) string {
	switch strings.ToUpper(strings.TrimSpace(d)) {
	case types.DecisionAllow, types.DecisionAsk, types.DecisionDeny:
		return strings.ToUpper(strings.TrimSpace(d))
	default:
		return types.DecisionAsk
	}
}
