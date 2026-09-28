package config

import (
	"strings"
	"time"
)

// RAGConfig keeps provider credentials in the server environment.
type RAGConfig struct {
	Enabled             bool
	Primary             RAGChatConfig
	Fallback            RAGChatConfig
	EmbeddingBaseURL    string
	EmbeddingModel      string
	EmbeddingAPIKey     string
	EmbeddingDimensions int
	RerankBaseURL       string
	RerankModel         string
	RerankAPIKey        string
	RerankTimeout       time.Duration
}

type RAGChatConfig struct {
	BaseURL       string
	Model         string
	APIKey        string
	HeadersJSON   string
	ExtraBodyJSON string
	SessionHeader string
	Timeout       time.Duration
}

func loadRAG() RAGConfig {
	return RAGConfig{
		Enabled:             getEnvAsBool("RAG_ENABLED", false),
		Primary:             loadRAGChat("RAG_CHAT_PRIMARY_", "https://opencode.ai/zen/go/v1", "deepseek-v4.1-flash", "x-opencode-session"),
		Fallback:            loadRAGChat("RAG_CHAT_FALLBACK_", "https://api.deepseek.com", "deepseek-flash", ""),
		EmbeddingBaseURL:    strings.TrimSpace(getEnv("RAG_EMBEDDING_BASE_URL", "https://router.tumuer.me/v1")),
		EmbeddingModel:      strings.TrimSpace(getEnv("RAG_EMBEDDING_MODEL", "BAAI/bge-m3")),
		EmbeddingAPIKey:     strings.TrimSpace(getEnv("RAG_EMBEDDING_API_KEY", "")),
		EmbeddingDimensions: getEnvAsInt("RAG_EMBEDDING_DIMENSIONS", 0),
		RerankBaseURL:       strings.TrimRight(strings.TrimSpace(getEnv("RAG_RERANK_BASE_URL", "https://router.tumuer.me/v1")), "/"),
		RerankModel:         strings.TrimSpace(getEnv("RAG_RERANK_MODEL", "Pro/BAAI/bge-reranker-v2-m3")),
		RerankAPIKey:        strings.TrimSpace(getEnv("RAG_RERANK_API_KEY", "")),
		RerankTimeout:       getEnvAsDuration("RAG_RERANK_TIMEOUT", 10*time.Second),
	}
}

func loadRAGChat(prefix, baseURL, model, sessionHeader string) RAGChatConfig {
	return RAGChatConfig{
		BaseURL:       strings.TrimRight(strings.TrimSpace(getEnv(prefix+"BASE_URL", baseURL)), "/"),
		Model:         strings.TrimSpace(getEnv(prefix+"MODEL", model)),
		APIKey:        strings.TrimSpace(getEnv(prefix+"API_KEY", "")),
		HeadersJSON:   getEnv(prefix+"HEADERS_JSON", `{}`),
		ExtraBodyJSON: getEnv(prefix+"EXTRA_BODY_JSON", `{}`),
		SessionHeader: strings.TrimSpace(getEnv(prefix+"SESSION_HEADER", sessionHeader)),
		Timeout:       getEnvAsDuration(prefix+"TIMEOUT", 30*time.Second),
	}
}
