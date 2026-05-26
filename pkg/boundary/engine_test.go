package boundary

import (
	"testing"

	"github.com/lihaiya/aiops/pkg/aiops/types"
)

func TestHardDenyKubectlDelete(t *testing.T) {
	plan := types.ActionPlan{
		Plugin:         "kubernetes",
		Action:         "delete",
		CommandPreview: "kubectl delete pod foo",
		Risk:           "write",
	}
	deny, _ := HardDeny(plan)
	if !deny {
		t.Fatal("expected hard deny")
	}
}

func TestEvaluateReadAllow(t *testing.T) {
	spec, err := ParseSpecYAML(prodStandardYAML)
	if err != nil {
		t.Fatal(err)
	}
	plan := types.ActionPlan{
		Plugin: "kubernetes",
		Action: "list",
		Risk:   "read",
	}
	res := Evaluate(spec, plan)
	if res.Decision != types.DecisionAllow {
		t.Fatalf("want ALLOW got %s", res.Decision)
	}
}

func TestEvaluateWriteAsk(t *testing.T) {
	spec, _ := ParseSpecYAML(prodStandardYAML)
	plan := types.ActionPlan{
		Plugin:         "kubernetes",
		Action:         "apply",
		CommandPreview: "kubectl apply -f deployment.yaml",
		Risk:           "write",
	}
	res := Evaluate(spec, plan)
	if res.HardDeny {
		t.Fatalf("unexpected hard deny: %v", res)
	}
}

const prodStandardYAML = `defaultDecision: ASK
rules:
  - id: read-allow
    match: { plugin: "", actions: ["get","list","describe","logs"] }
    decision: ALLOW
  - id: write-ask
    match: { plugin: "", actions: ["delete","apply","patch","create"] }
    decision: ASK
`
