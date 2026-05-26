package platform

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lihaiya/aiops/pkg/aiops/httpclient"
	"github.com/lihaiya/aiops/pkg/llm"
	"github.com/lihaiya/aiops/pkg/runtime"
)

const helpdeskSystemPrompt = `你是 IT HelpDesk 助手。仅根据提供的知识库内容回答办公常见问题。
不要生成 kubectl、删除、执行类运维命令，不要输出 ActionPlan JSON。`

func (h *Handler) HelpdeskChat(c *gin.Context) {
	var body struct {
		Message string `json:"message"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	kbURL := runtime.EnvOr("AIOPS_MODULE_KB_URL", "http://aiops-module-kb:8085")
	var kbResp struct {
		Data []struct {
			Title   string `json:"title"`
			Content string `json:"content"`
		} `json:"data"`
	}
	_ = httpclient.PostJSON(c.Request.Context(), kbURL+"/internal/v1/kb/search", map[string]interface{}{
		"query": body.Message, "limit": 5,
	}, c, &kbResp)

	var citations []string
	var kbText strings.Builder
	for _, d := range kbResp.Data {
		kbText.WriteString(d.Title + ": " + d.Content + "\n")
		citations = append(citations, d.Title)
	}

	client := llm.NewFromEnv()
	if !client.Configured() {
		c.JSON(503, gin.H{"error": "llm not configured"})
		return
	}
	msgs := []llm.Message{
		{Role: "system", Content: helpdeskSystemPrompt},
		{Role: "system", Content: "知识库:\n" + kbText.String()},
		{Role: "user", Content: body.Message},
	}
	reply, err := client.ChatCompletion(context.Background(), msgs)
	if err != nil {
		c.JSON(502, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{
		"code": 0,
		"data": gin.H{
			"reply":     reply,
			"citations": citations,
		},
	})
}
