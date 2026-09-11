package svc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"pet/backend/internal/common"
)

// AIMessage 与 OpenAI 兼容接口的 message 结构
type AIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type aiChatRequest struct {
	Model       string      `json:"model"`
	Messages    []AIMessage `json:"messages"`
	Temperature float64     `json:"temperature"`
}

type aiChatResponse struct {
	Choices []struct {
		Message AIMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

var aiHTTP = &http.Client{}

// AIChat 调用 OpenAI 兼容 chat/completions，返回模型回复文本
func (sc *ServiceContext) AIChat(ctx context.Context, messages []AIMessage) (string, error) {
	c := sc.Config.AI
	if !sc.Config.AIEnabled() {
		return "", common.ErrAIService
	}
	timeout := time.Duration(c.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	callCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	body, _ := json.Marshal(aiChatRequest{Model: c.Model, Messages: messages, Temperature: 0.7})
	req, err := http.NewRequestWithContext(callCtx, http.MethodPost,
		trimSlash(c.BaseURL)+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.ApiKey)

	resp, err := aiHTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var out aiChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("解析 AI 响应失败(http %d): %w", resp.StatusCode, err)
	}
	if out.Error != nil && out.Error.Message != "" {
		return "", fmt.Errorf("AI 接口返回错误: %s", out.Error.Message)
	}
	if len(out.Choices) == 0 || out.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("AI 响应无内容(http %d)", resp.StatusCode)
	}
	return out.Choices[0].Message.Content, nil
}

func trimSlash(s string) string {
	for len(s) > 0 && s[len(s)-1] == '/' {
		s = s[:len(s)-1]
	}
	return s
}
