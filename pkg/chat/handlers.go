package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	aimodel "github.com/lihaiya/aiops/pkg/aiops/model"
	"github.com/lihaiya/aiops/pkg/aiops/types"
	"github.com/lihaiya/aiops/pkg/aiops/httpclient"
	"github.com/lihaiya/aiops/pkg/aiops/identity"
	"github.com/lihaiya/aiops/pkg/audit"
	"github.com/lihaiya/aiops/pkg/llm"
	"github.com/lihaiya/aiops/pkg/runtime"
	"gorm.io/gorm"
)

type Handler struct {
	DB          *gorm.DB
	PolicyURL   string
	ExecutorURL string
	KbURL       string
	Audit       *audit.Client
	LLM         *llm.Client
}

func Register(r *gin.Engine, db *gorm.DB) {
	h := &Handler{
		DB:          db,
		PolicyURL:   runtime.EnvOr("AIOPS_POLICY_URL", "http://aiops-policy:8083"),
		ExecutorURL: runtime.EnvOr("AIOPS_EXECUTOR_URL", "http://aiops-executor:8084"),
		KbURL:       runtime.EnvOr("AIOPS_MODULE_KB_URL", "http://aiops-module-kb:8085"),
		Audit:       audit.NewClient(),
		LLM:         llm.NewFromEnv(),
	}
	api := r.Group("/api/v1")
	api.POST("/chat", h.PostChat)
	api.GET("/chat/stream", h.StreamChat)
	api.GET("/chat/sessions", h.ListSessions)
	api.GET("/chat/sessions/:id/messages", h.ListMessages)
	api.POST("/chat/assist", h.PostAssist)
	api.POST("/chat/rca", h.PostRCA)
}

func (h *Handler) llmConfigured() bool {
	return h.LLM != nil && h.LLM.Configured()
}

// getLlmClient 根据请求中的 model 参数动态选择 LLM 配置。
// 如果指定了 model 且在 DB 中找到对应配置，使用该配置；否则回退到默认环境变量配置。
func (h *Handler) getLlmClient(c *gin.Context) (*llm.Client, error) {
	modelName := c.Query("model")
	if modelName == "" {
		var body struct{ Model string `json:"model"` }
		_ = c.ShouldBindJSON(&body)
		modelName = body.Model
	}
	if modelName != "" {
		var cfg aimodel.LlmConfig
		if err := h.DB.Where("name = ? AND active = ?", modelName, true).First(&cfg).Error; err == nil {
			return llm.New(llm.Config{
				BaseURL: cfg.BaseURL,
				APIKey:  cfg.APIKey,
				Model:   cfg.Model,
				Timeout: 60e9, // 60s
			}), nil
		}
	}
	if h.LLM == nil || !h.LLM.Configured() {
		return nil, fmt.Errorf("llm not configured")
	}
	return h.LLM, nil
}

func (h *Handler) PostChat(c *gin.Context) {
	client, err := h.getLlmClient(c)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "llm not configured"})
		return
	}
	var body struct {
		Environment string `json:"environment"`
		SessionID   int64  `json:"sessionId"`
		Message     string `json:"message"`
		UseRAG      bool   `json:"useRag"`
		Model       string `json:"model"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Environment == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "environment required"})
		return
	}
	c.Request.Header.Set(runtime.HeaderEnvironment, body.Environment)

	session, err := h.ensureSession(c, body.SessionID, body.Environment, body.Message)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.saveMessage(session.ID, "user", body.Message, "")

	ctx := c.Request.Context()
	messages := h.buildMessages(ctx, session.ID, body.Message, body.UseRAG, SystemPromptOps)
	reply, err := client.ChatCompletion(ctx, messages)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	plans := ParseActionPlans(reply)
	plansJSON, _ := json.Marshal(plans)
	h.saveMessage(session.ID, "assistant", reply, string(plansJSON))
	h.auditEvent(c, "chat.message", body.Environment, gin.H{"sessionId": session.ID})

	execResults := h.processPlans(c, body.Environment, session.ID, plans)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"sessionId":  session.ID,
			"reply":      reply,
			"plans":      plans,
			"executions": execResults,
		},
	})
}

func (h *Handler) StreamChat(c *gin.Context) {
	client, err := h.getLlmClient(c)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "llm not configured"})
		return
	}
	env := c.Query("environment")
	msg := c.Query("message")
	if env == "" || msg == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "environment and message required"})
		return
	}
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	sessionID, _ := strconv.ParseInt(c.Query("sessionId"), 10, 64)
	session, _ := h.ensureSession(c, sessionID, env, msg)
	h.saveMessage(session.ID, "user", msg, "")

	reply, err := client.ChatCompletion(c.Request.Context(), h.buildMessages(c.Request.Context(), session.ID, msg, false, SystemPromptOps))
	if err != nil {
		fmt.Fprintf(c.Writer, "data: %s\n\n", jsonEscape(`{"error":"`+err.Error()+`"}`))
		return
	}
	plans := ParseActionPlans(reply)
	pj, _ := json.Marshal(plans)
	h.saveMessage(session.ID, "assistant", reply, string(pj))

	for _, chunk := range chunkText(reply, 80) {
		payload, _ := json.Marshal(gin.H{"type": "delta", "content": chunk})
		fmt.Fprintf(c.Writer, "data: %s\n\n", string(payload))
		c.Writer.Flush()
	}
	done, _ := json.Marshal(gin.H{"type": "done", "sessionId": session.ID, "plans": plans})
	fmt.Fprintf(c.Writer, "data: %s\n\n", string(done))
}

func (h *Handler) PostAssist(c *gin.Context) {
	h.postMode(c, SystemPromptAssist, false)
}

func (h *Handler) PostRCA(c *gin.Context) {
	var body struct {
		Environment string `json:"environment"`
		IncidentID  string `json:"incidentId"`
	}
	_ = c.ShouldBindJSON(&body)
	// 附加 audit 摘要
	var logs []aimodel.AuditLog
	h.DB.Where("environment = ?", body.Environment).Order("id desc").Limit(50).Find(&logs)
	extra, _ := json.Marshal(logs)
	c.Set("rca_context", string(extra))
	h.postMode(c, SystemPromptRCA, false)
}

func (h *Handler) postMode(c *gin.Context, system string, useRAG bool) {
	client, err := h.getLlmClient(c)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "llm not configured"})
		return
	}
	var body struct {
		Environment string `json:"environment"`
		SessionID   int64  `json:"sessionId"`
		Message     string `json:"message"`
		Model       string `json:"model"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	session, _ := h.ensureSession(c, body.SessionID, body.Environment, body.Message)
	h.saveMessage(session.ID, "user", body.Message, "")
	msgs := h.buildMessages(c.Request.Context(), session.ID, body.Message, useRAG, system)
	if v, ok := c.Get("rca_context"); ok {
		msgs = append(msgs, llm.Message{Role: "user", Content: "审计上下文:\n" + v.(string)})
	}
	reply, err := client.ChatCompletion(c.Request.Context(), msgs)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	h.saveMessage(session.ID, "assistant", reply, "")
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"sessionId": session.ID, "reply": reply}})
}

func (h *Handler) ListSessions(c *gin.Context) {
	var list []aimodel.ChatSession
	h.DB.Where("user_id = ?", identity.UserID(c)).Order("id desc").Limit(50).Find(&list)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": list})
}

func (h *Handler) ListMessages(c *gin.Context) {
	id := c.Param("id")
	var list []aimodel.ChatMessage
	h.DB.Where("session_id = ?", id).Order("id asc").Find(&list)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": list})
}

func (h *Handler) ensureSession(c *gin.Context, sessionID int64, env, title string) (aimodel.ChatSession, error) {
	var session aimodel.ChatSession
	if sessionID > 0 {
		if err := h.DB.First(&session, sessionID).Error; err != nil {
			return session, err
		}
		return session, nil
	}
	session = aimodel.ChatSession{
		UserID:      identity.UserID(c),
		Environment: env,
		Title:       truncate(title, 40),
	}
	return session, h.DB.Create(&session).Error
}

func (h *Handler) saveMessage(sessionID int64, role, content, plansJSON string) {
	h.DB.Create(&aimodel.ChatMessage{SessionID: sessionID, Role: role, Content: content, PlansJSON: plansJSON})
}

func (h *Handler) buildMessages(ctx context.Context, sessionID int64, userMsg string, useRAG bool, system string) []llm.Message {
	msgs := []llm.Message{{Role: "system", Content: system}}
	if useRAG && h.KbURL != "" {
		var kbResp struct {
			Data []struct {
				Title   string `json:"title"`
				Content string `json:"content"`
			} `json:"data"`
		}
		_ = httpclient.PostJSON(ctx, h.KbURL+"/internal/v1/kb/search", map[string]interface{}{
			"query": userMsg, "limit": 3,
		}, nil, &kbResp)
		var cite string
		for _, d := range kbResp.Data {
			cite += d.Title + ": " + d.Content + "\n"
		}
		if cite != "" {
			msgs = append(msgs, llm.Message{Role: "system", Content: "知识库:\n" + cite})
		}
	}
	var history []aimodel.ChatMessage
	h.DB.Where("session_id = ?", sessionID).Order("id asc").Limit(20).Find(&history)
	for _, m := range history {
		msgs = append(msgs, llm.Message{Role: m.Role, Content: m.Content})
	}
	msgs = append(msgs, llm.Message{Role: "user", Content: userMsg})
	return msgs
}

func (h *Handler) processPlans(c *gin.Context, env string, sessionID int64, plans []types.ActionPlan) []interface{} {
	if len(plans) == 0 {
		return nil
	}
	var results []interface{}
	for _, plan := range plans {
		h.auditEvent(c, "action.plan", env, plan)
		var preview types.EvaluateResponse
		_ = httpclient.PostJSON(c.Request.Context(), h.PolicyURL+"/internal/v1/evaluate",
			types.EvaluateRequest{Environment: env, Plan: plan}, c, &preview)
		if preview.Decision == types.DecisionAllow {
			var runOut map[string]interface{}
			_ = httpclient.PostJSON(c.Request.Context(), h.ExecutorURL+"/internal/v1/run",
				types.RunRequest{Environment: env, SessionID: sessionID, Plan: plan}, c, &runOut)
			results = append(results, runOut)
		} else {
			results = append(results, gin.H{"decision": preview.Decision, "preview": preview})
		}
	}
	return results
}

func (h *Handler) auditEvent(c *gin.Context, eventType, env string, payload interface{}) {
	h.Audit.Log(c.Request.Context(), c, types.AuditRequest{
		EventType: eventType, Environment: env, UserID: identity.UserID(c), Payload: payload,
	})
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func chunkText(s string, size int) []string {
	var out []string
	for len(s) > 0 {
		if len(s) <= size {
			out = append(out, s)
			break
		}
		out = append(out, s[:size])
		s = s[size:]
	}
	return out
}

func jsonEscape(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
