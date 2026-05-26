package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Config LLM 连接配置（OpenAI 兼容）。
type Config struct {
	BaseURL string
	APIKey  string
	Model   string
	Timeout time.Duration
}

func ConfigFromEnv() Config {
	base := os.Getenv("AIOPS_LLM_BASE_URL")
	if base == "" {
		base = os.Getenv("LLM_BASE_URL")
	}
	if base == "" {
		base = "https://api.openai.com/v1"
	}
	key := os.Getenv("AIOPS_LLM_API_KEY")
	if key == "" {
		key = os.Getenv("LLM_API_KEY")
	}
	model := os.Getenv("AIOPS_LLM_MODEL")
	if model == "" {
		model = os.Getenv("LLM_MODEL")
	}
	if model == "" {
		model = "gpt-4o-mini"
	}
	return Config{
		BaseURL: strings.TrimRight(base, "/"),
		APIKey:  key,
		Model:   model,
		Timeout: 60 * time.Second,
	}
}

func (c Config) OK() bool {
	return c.APIKey != ""
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Client struct {
	cfg Config
}

func New(cfg Config) *Client {
	return &Client{cfg: cfg}
}

func NewFromEnv() *Client {
	return New(ConfigFromEnv())
}

// Configured 是否已配置 API Key。
func (c *Client) Configured() bool {
	return c != nil && c.cfg.OK()
}

// ChatCompletion 调用 chat/completions。
func (c *Client) ChatCompletion(ctx context.Context, messages []Message) (string, error) {
	if !c.cfg.OK() {
		return "", fmt.Errorf("llm not configured")
	}
	body := map[string]interface{}{
		"model":    c.cfg.Model,
		"messages": messages,
	}
	raw, _ := json.Marshal(body)
	url := c.cfg.BaseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	client := &http.Client{Timeout: c.cfg.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("llm http %s: %s", resp.Status, string(b))
	}
	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return "", err
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("empty llm response")
	}
	return out.Choices[0].Message.Content, nil
}
