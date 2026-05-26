package boundary

import (
	"strings"

	"github.com/lihaiya/aiops/pkg/aiops/types"
	"gopkg.in/yaml.v3"
)

func ParseSpecYAML(raw string) (types.PolicySpec, error) {
	var spec types.PolicySpec
	err := yaml.Unmarshal([]byte(raw), &spec)
	return spec, err
}

func matchRule(plan types.ActionPlan, rule types.Rule) bool {
	m := rule.Match
	if m.Plugin != "" && m.Plugin != plan.Plugin {
		return false
	}
	if len(m.Actions) > 0 {
		ok := false
		for _, a := range m.Actions {
			if strings.EqualFold(a, plan.Action) {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	if m.CommandPattern != "" && plan.CommandPreview != "" {
		if !matchPattern(m.CommandPattern, plan.CommandPreview) {
			return false
		}
	}
	return true
}

func matchPattern(pattern, preview string) bool {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return true
	}
	return strings.Contains(strings.ToLower(preview), strings.ToLower(pattern))
}
