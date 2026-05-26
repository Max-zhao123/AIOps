package boundary

import (
	"fmt"
	"testing"

	"github.com/lihaiya/aiops/pkg/aiops/types"
)

func BenchmarkEvaluate100Rules(b *testing.B) {
	var rules []types.Rule
	for i := 0; i < 100; i++ {
		rules = append(rules, types.Rule{
			ID: fmt.Sprintf("r%d", i),
			Match: types.RuleMatch{
				Actions: []string{"get", "list"},
			},
			Decision: types.DecisionAsk,
		})
	}
	spec := types.PolicySpec{DefaultDecision: types.DecisionAsk, Rules: rules}
	plan := types.ActionPlan{Plugin: "kubernetes", Action: "list", Risk: "read"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Evaluate(spec, plan)
	}
}
