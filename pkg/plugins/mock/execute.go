package mock

import (
	"github.com/lihaiya/aiops/pkg/aiops/types"
)

func Execute(plan types.ActionPlan) types.ExecuteResponse {
	return types.ExecuteResponse{
		Status:  "ok",
		Output:  map[string]interface{}{"plugin": "mock", "action": plan.Action},
		Message: "mock executed",
	}
}
