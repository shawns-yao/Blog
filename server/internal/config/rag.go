package config

import (
	"strings"
	"time"
)

// RAGConfig keeps provider credentials in the server environment.
type RAGConfig struct {
	Enabled              bool
	EvaluationTraceDir   string
	GPT                  RAGChatConfig
	Grok                 RAGChatConfig
	Gemini               RAGChatConfig
	Primary              RAGChatConfig
	Fallback             RAGChatConfig
	EmbeddingBaseURL     string
	EmbeddingModel       string
	EmbeddingAPIKey      string
	EmbeddingDimensions  int
	RerankBaseURL        string
	RerankModel          string
	RerankAPIKey         string
	RerankTimeout        time.Duration
	HistoryMaxRounds     int
	HistoryMaxCharacters int
}

type RAGChatConfig struct {
	Name          string
	Protocol      string
	BaseURL       string
	Model         string
	APIKey        string
	HeadersJSON   string
	ExtraBodyJSON string
	SessionHeader string
	Timeout       time.Duration
}

func loadRAG() RAGConfig {
	gpt := loadRAGChat("RAG_CHAT_GPT_", "gpt", "openai", "http://127.0.0.1:8317/v1", "gpt-6-sol", "")
	gpt.ExtraBodyJSON = getEnv("RAG_CHAT_GPT_EXTRA_BODY_JSON", `{"reasoning_effort":"medium"}`)
	return RAGConfig{
		Enabled:              getEnvAsBool("RAG_ENABLED", false),
		EvaluationTraceDir:   strings.TrimSpace(getEnv("RAG_EVALUATION_TRACE_DIR", "")),
		GPT:                  gpt,
		Grok:                 loadRAGChat("RAG_CHAT_GROK_", "grok", "openai", "https://ai.hybgzs.com/v1", "grok-4.7", ""),
		Gemini:               loadRAGChat("RAG_CHAT_GEMINI_", "gemini", "openai", "https://ai.hybgzs.com/v1", "gemini-3.1-flash-lite-preview", ""),
		Primary:              loadRAGChat("RAG_CHAT_PRIMARY_", "opencode_go", "openai", "https://opencode.ai/zen/go/v1", "deepseek-v4.1-flash", "x-opencode-session"),
		Fallback:             loadRAGChat("RAG_CHAT_FALLBACK_", "deepseek", "openai", "https://api.deepseek.com", "deepseek-flash", ""),
		EmbeddingBaseURL:     strings.TrimSpace(getEnv("RAG_EMBEDDING_BASE_URL", "https://router.tumuer.me/v1")),
		EmbeddingModel:       strings.TrimSpace(getEnv("RAG_EMBEDDING_MODEL", "BAAI/bge-m3")),
		EmbeddingAPIKey:      strings.TrimSpace(getEnv("RAG_EMBEDDING_API_KEY", "")),
		EmbeddingDimensions:  getEnvAsInt("RAG_EMBEDDING_DIMENSIONS", 0),
		RerankBaseURL:        strings.TrimRight(strings.TrimSpace(getEnv("RAG_RERANK_BASE_URL", "https://router.tumuer.me/v1")), "/"),
		RerankModel:          strings.TrimSpace(getEnv("RAG_RERANK_MODEL", "Pro/BAAI/bge-reranker-v2-m3")),
		RerankAPIKey:         strings.TrimSpace(getEnv("RAG_RERANK_API_KEY", "")),
		RerankTimeout:        getEnvAsDuration("RAG_RERANK_TIMEOUT", 10*time.Second),
		HistoryMaxRounds:     getEnvAsInt("RAG_HISTORY_MAX_ROUNDS", 5),
		HistoryMaxCharacters: getEnvAsInt("RAG_HISTORY_MAX_CHARACTERS", 20000),
	}
}

func (c RAGConfig) ChatChannels() []RAGChatConfig {
	return []RAGChatConfig{c.GPT, c.Grok, c.Gemini, c.Primary, c.Fallback}
}

func loadRAGChat(prefix, name, protocol, baseURL, model, sessionHeader string) RAGChatConfig {
	return RAGChatConfig{
		Name:          name,
		Protocol:      strings.ToLower(strings.TrimSpace(getEnv(prefix+"PROTOCOL", protocol))),
		BaseURL:       strings.TrimRight(strings.TrimSpace(getEnv(prefix+"BASE_URL", baseURL)), "/"),
		Model:         strings.TrimSpace(getEnv(prefix+"MODEL", model)),
		APIKey:        strings.TrimSpace(getEnv(prefix+"API_KEY", "")),
		HeadersJSON:   getEnv(prefix+"HEADERS_JSON", `{}`),
		ExtraBodyJSON: getEnv(prefix+"EXTRA_BODY_JSON", `{}`),
		SessionHeader: strings.TrimSpace(getEnv(prefix+"SESSION_HEADER", sessionHeader)),
		Timeout:       getEnvAsDuration(prefix+"TIMEOUT", 30*time.Second),
	}
}
