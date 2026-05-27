package metrics

import (
	"encoding/json"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/gin-gonic/gin"
)

// 简化版 Prometheus 风格指标（REQ-102）。
// 由于环境限制无法安装 prometheus/client_golang，
// 使用内存计数器提供 /metrics JSON 端点。
// 生产部署时替换为 prometheus/client_golang。

var (
	chatRequestsTotal     sync.Map // map[string]*atomic.Int64, key=env:model
	executionsTotal        sync.Map // map[string]*atomic.Int64, key=env:plugin:decision:status
	policyEvaluationsTotal sync.Map // map[string]*atomic.Int64, key=decision:rule
	llmRequestCount        sync.Map // map[string]*atomic.Int64, key=model
	llmRequestDurationMs   sync.Map // map[string]*durationAccumulator, key=model
)

type durationAccumulator struct {
	totalMs int64
	count   int64
}

// RegisterMetricsEndpoint 在 gin Engine 上注册 /metrics 路由（REQ-102）。
func RegisterMetricsEndpoint(r *gin.Engine) {
	r.GET("/metrics", metricsHandler)
}

func metricsHandler(c *gin.Context) {
	metrics := map[string]interface{}{
		"chat_requests_total":      mapSnapshot(&chatRequestsTotal),
		"executions_total":        mapSnapshot(&executionsTotal),
		"policy_evaluations_total": mapSnapshot(&policyEvaluationsTotal),
		"llm_request_count":       mapSnapshot(&llmRequestCount),
		"llm_request_duration_ms": durationMapSnapshot(&llmRequestDurationMs),
	}
	c.JSON(http.StatusOK, metrics)
}

func mapSnapshot(m *sync.Map) map[string]int64 {
	result := make(map[string]int64)
	m.Range(func(key, value interface{}) bool {
		if v, ok := value.(*atomic.Int64); ok {
			result[key.(string)] = v.Load()
		}
		return true
	})
	return result
}

func durationMapSnapshot(m *sync.Map) map[string]map[string]interface{} {
	result := make(map[string]map[string]interface{})
	m.Range(func(key, value interface{}) bool {
		if v, ok := value.(*durationAccumulator); ok {
			count := atomic.LoadInt64(&v.count)
			totalMs := atomic.LoadInt64(&v.totalMs)
			var avgMs float64
			if count > 0 {
				avgMs = float64(totalMs) / float64(count)
			}
			result[key.(string)] = map[string]interface{}{
				"count":    count,
				"total_ms": totalMs,
				"avg_ms":   avgMs,
			}
		}
		return true
	})
	return result
}

func getOrCreateCounter(m *sync.Map, key string) *atomic.Int64 {
	if v, ok := m.Load(key); ok {
		return v.(*atomic.Int64)
	}
	counter := &atomic.Int64{}
	actual, _ := m.LoadOrStore(key, counter)
	return actual.(*atomic.Int64)
}

func getOrCreateDuration(m *sync.Map, key string) *durationAccumulator {
	if v, ok := m.Load(key); ok {
		return v.(*durationAccumulator)
	}
	d := &durationAccumulator{}
	actual, _ := m.LoadOrStore(key, d)
	return actual.(*durationAccumulator)
}

// IncChatRequest 增加 chat 请求计数。
func IncChatRequest(environment, model string) {
	key := environment + ":" + model
	getOrCreateCounter(&chatRequestsTotal, key).Add(1)
}

// IncExecution 增加执行计数。
func IncExecution(environment, plugin, decision, status string) {
	key := environment + ":" + plugin + ":" + decision + ":" + status
	getOrCreateCounter(&executionsTotal, key).Add(1)
}

// IncPolicyEvaluation 增加策略评估计数。
func IncPolicyEvaluation(decision, matchedRuleID string) {
	key := decision + ":" + matchedRuleID
	getOrCreateCounter(&policyEvaluationsTotal, key).Add(1)
}

// ObserveLLMDuration 记录 LLM 请求耗时。
func ObserveLLMDuration(model string, seconds float64) {
	ms := int64(seconds * 1000)
	d := getOrCreateDuration(&llmRequestDurationMs, model)
	atomic.AddInt64(&d.totalMs, ms)
	atomic.AddInt64(&d.count, 1)
}

// MarshalMetrics 返回 JSON 格式的指标数据。
func MarshalMetrics() ([]byte, error) {
	metrics := map[string]interface{}{
		"chat_requests_total":      mapSnapshot(&chatRequestsTotal),
		"executions_total":        mapSnapshot(&executionsTotal),
		"policy_evaluations_total": mapSnapshot(&policyEvaluationsTotal),
		"llm_request_count":       mapSnapshot(&llmRequestCount),
		"llm_request_duration_ms": durationMapSnapshot(&llmRequestDurationMs),
	}
	return json.Marshal(metrics)
}
