package prometheus

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type QueryRequest struct {
	Query  string                 `json:"query"`
	Start  string                 `json:"start"`
	End    string                 `json:"end"`
	Params map[string]interface{} `json:"params"`
}

func ExecuteQuery(req QueryRequest) (map[string]interface{}, error) {
	base := os.Getenv("AIOPS_PROMETHEUS_URL")
	if base == "" {
		base = "http://prometheus:9090"
	}
	promQL := req.Query
	if promQL == "" {
		if p, ok := req.Params["promql"].(string); ok {
			promQL = p
		}
	}
	if promQL == "" {
		return nil, fmt.Errorf("query or promql required")
	}
	u, _ := url.Parse(strings.TrimRight(base, "/") + "/api/v1/query")
	q := u.Query()
	q.Set("query", promQL)
	u.RawQuery = q.Encode()
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(u.String())
	if err != nil {
		return map[string]interface{}{"status": "unavailable", "error": err.Error()}, nil
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	var out map[string]interface{}
	_ = json.Unmarshal(b, &out)
	if out == nil {
		out = map[string]interface{}{"raw": string(b)}
	}
	return out, nil
}
