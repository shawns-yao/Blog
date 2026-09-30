package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var ErrRAGChatUnavailable = errors.New("chat provider unavailable")

// RAGChatClient reuses the OpenAI wire format with isolated headers per provider.
type RAGChatClient struct {
	baseURL       string
	apiKey        string
	headers       http.Header
	extraBody     map[string]json.RawMessage
	sessionHeader string
	client        *http.Client
}

func NewRAGChatClient(baseURL, apiKey, headersJSON, extraBodyJSON, sessionHeader string, timeout time.Duration, protocol string) (*RAGChatClient, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") ||
		parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || apiKey == "" || timeout <= 0 || timeout > 30*time.Second {
		return nil, errors.New("invalid chat configuration")
	}
	if protocol != "openai" {
		return nil, errors.New("invalid chat protocol")
	}
	var headers map[string]string
	var extra map[string]json.RawMessage
	if json.Unmarshal([]byte(headersJSON), &headers) != nil || json.Unmarshal([]byte(extraBodyJSON), &extra) != nil {
		return nil, errors.New("invalid chat configuration")
	}
	reserved := map[string]bool{"model": true, "messages": true, "stream": true, "max_tokens": true, "temperature": true}
	for key := range extra {
		if reserved[key] {
			return nil, errors.New("invalid chat configuration")
		}
	}
	result := &RAGChatClient{baseURL: baseURL, apiKey: apiKey, headers: make(http.Header), extraBody: extra, sessionHeader: sessionHeader,
		client: &http.Client{Timeout: timeout, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
	result.headers.Set("User-Agent", "grtblog-rag/1.0")
	for name, value := range headers {
		if strings.EqualFold(name, "Authorization") || strings.EqualFold(name, "Content-Type") ||
			strings.EqualFold(name, "Host") || strings.EqualFold(name, sessionHeader) {
			return nil, errors.New("invalid chat configuration")
		}
		result.headers.Set(name, value)
	}
	return result, nil
}

func (c *RAGChatClient) Chat(ctx context.Context, request ChatRequest, sessionID string) (*ChatResponse, error) {
	body := make(map[string]any, len(c.extraBody)+4)
	for key, value := range c.extraBody {
		body[key] = value
	}
	body["model"], body["messages"], body["stream"] = request.Model, request.Messages, false
	if request.MaxTokens != nil {
		body["max_tokens"] = *request.MaxTokens
	}
	if request.Temperature != nil {
		body["temperature"] = *request.Temperature
	}
	if request.JSONMode {
		if _, configured := body["response_format"]; !configured {
			body["response_format"] = map[string]string{"type": "json_object"}
		}
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, ErrRAGChatUnavailable
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return nil, ErrRAGChatUnavailable
	}
	req.Header = c.headers.Clone()
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	if c.sessionHeader != "" {
		req.Header.Set(c.sessionHeader, sessionID)
	}
	response, err := c.client.Do(req)
	if err != nil {
		return nil, ErrRAGChatUnavailable
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		// Provider bodies may echo secrets or source content; expose only the status.
		return nil, fmt.Errorf("%w (HTTP %d)", ErrRAGChatUnavailable, response.StatusCode)
	}
	var generated openAIResponse
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&generated); err != nil ||
		generated.Error != nil || len(generated.Choices) == 0 || strings.TrimSpace(generated.Choices[0].Message.Content) == "" {
		return nil, ErrRAGChatUnavailable
	}
	return &ChatResponse{Content: generated.Choices[0].Message.Content, Model: generated.Model,
		FinishReason: generated.Choices[0].FinishReason}, nil
}
