package logs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

// ExecuteQuery 对接 Elasticsearch 简易 search（只读）。
func ExecuteQuery(req QueryRequest) (map[string]interface{}, error) {
	base := os.Getenv("AIOPS_LOGS_URL")
	if base == "" {
		return nil, fmt.Errorf("logs endpoint not configured (set AIOPS_LOGS_URL)")
	}
	index := "_all"
	if v, ok := req.Params["index"].(string); ok && v != "" {
		index = v
	}
	body := map[string]interface{}{
		"size": 100,
		"query": map[string]interface{}{
			"query_string": map[string]interface{}{"query": req.Query},
		},
	}
	if req.Start != "" || req.End != "" {
		body["query"] = map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []interface{}{
					map[string]interface{}{"query_string": map[string]interface{}{"query": req.Query}},
				},
			},
		}
	}
	raw, _ := json.Marshal(body)
	url := strings.TrimRight(base, "/") + "/" + index + "/_search"
	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("logs http %s", resp.Status)
	}
	var out map[string]interface{}
	json.Unmarshal(b, &out)
	return out, nil
}
