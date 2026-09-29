package rag

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/url"
	"strings"

	appconfig "github.com/shawns-yao/shawn-blog/server/internal/config"
	domainconfig "github.com/shawns-yao/shawn-blog/server/internal/domain/config"
	domain "github.com/shawns-yao/shawn-blog/server/internal/domain/rag"
	infraai "github.com/shawns-yao/shawn-blog/server/internal/infra/ai"
	infrarag "github.com/shawns-yao/shawn-blog/server/internal/infra/rag"
)

var errDisabled = errors.New("rag disabled")
var errNotConfigured = errors.New("rag not configured")

type ConfigReader interface {
	ListConfigs(context.Context, []string) ([]domainconfig.SysConfig, error)
}

type settings struct {
	profile       string
	channels      []chatChannel
	chatPriority  []string
	embedder      *infraai.EmbeddingPool
	reranker      *infraai.RerankPool
	tuning        domain.Tuning
	minSimilarity float64
}

type chatChannel struct {
	client  *infraai.RAGChatClient
	model   string
	primary bool
	name    string
}

func (s *Service) loadSettings(ctx context.Context) (settings, error) {
	result, err := s.loadIndexSettings(ctx)
	if err != nil {
		return result, err
	}
	for _, channel := range s.orderedChatChannels(result.chatPriority) {
		if channel.Disabled || channel.APIKey == "" {
			continue
		}
		client, err := infraai.NewRAGChatClient(channel.BaseURL, channel.APIKey, channel.HeadersJSON, channel.ExtraBodyJSON, channel.SessionHeader, channel.Timeout, channel.Protocol)
		if err != nil || channel.Model == "" {
			continue
		}
		result.channels = append(result.channels, chatChannel{client: client, model: channel.Model, primary: len(result.channels) == 0, name: channel.Name})
	}
	if len(result.channels) == 0 {
		return result, errNotConfigured
	}
	if result.tuning.RerankEnabled {
		result.reranker, err = s.newReranker()
		// A reranker configuration error affects retrieval, not general chat.
	}
	result.embedder = s.embedder
	return result, nil
}

func (s *Service) embeddingConfigured() bool {
	p := s.providers
	parsed, err := url.Parse(p.EmbeddingBaseURL)
	return err == nil && parsed.Host != "" && (parsed.Scheme == "https" || parsed.Scheme == "http") &&
		parsed.User == nil && parsed.RawQuery == "" && parsed.Fragment == "" &&
		p.EmbeddingModel != "" && p.EmbeddingAPIKey != "" && s.embedder != nil && s.indexEmbedder != nil
}

// Indexing depends on embedding configuration, independently of chat/rerank availability.
func (s *Service) loadIndexSettings(ctx context.Context) (settings, error) {
	if !s.providers.Enabled {
		return settings{}, errDisabled
	}
	result, err := s.loadTuning(ctx)
	if err != nil {
		return result, err
	}
	if !s.embeddingConfigured() {
		return result, errNotConfigured
	}
	result.embedder = s.indexEmbedder
	return result, nil
}

func chatConfigured(channel appconfig.RAGChatConfig) bool {
	if channel.Disabled {
		return false
	}
	_, err := infraai.NewRAGChatClient(channel.BaseURL, channel.APIKey, channel.HeadersJSON, channel.ExtraBodyJSON, channel.SessionHeader, channel.Timeout, channel.Protocol)
	return err == nil && channel.Model != ""
}

func (s *Service) newReranker() (*infraai.RerankPool, error) {
	if s.reranker == nil {
		return nil, infraai.ErrRerankUnavailable
	}
	return s.reranker, nil
}

func (s *Service) embeddingRoutes() []infraai.ModelRoute {
	p := s.providers
	routes := []infraai.ModelRoute{{Name: p.EmbeddingProvider, BaseURL: p.EmbeddingBaseURL, Model: p.EmbeddingModel,
		APIKey: p.EmbeddingAPIKey, Timeout: p.EmbeddingTimeout}}
	return appendModelRoute(routes, p.EmbeddingFallback)
}

func (s *Service) rerankRoutes() []infraai.ModelRoute {
	p := s.providers
	routes := []infraai.ModelRoute{{Name: p.RerankProvider, BaseURL: p.RerankBaseURL, Model: p.RerankModel,
		APIKey: p.RerankAPIKey, Timeout: p.RerankTimeout}}
	return appendModelRoute(appendModelRoute(routes, p.RerankFallback), p.RerankLastResort)
}

func appendModelRoute(routes []infraai.ModelRoute, config appconfig.RAGModelConfig) []infraai.ModelRoute {
	if config.BaseURL != "" || config.Model != "" || config.APIKey != "" {
		routes = append(routes, infraai.ModelRoute{Name: config.Name, BaseURL: config.BaseURL, Model: config.Model,
			APIKey: config.APIKey, Timeout: config.Timeout})
	}
	return routes
}

func (s *Service) loadTuning(ctx context.Context) (settings, error) {
	var result settings
	items, err := s.config.ListConfigs(ctx, configKeys)
	if err != nil {
		return result, errNotConfigured
	}
	values := make(map[string]string, len(items))
	for _, item := range items {
		values[item.Key] = strings.TrimSpace(item.Value)
	}
	tuning, err := decodeTuning(values)
	if err != nil {
		return result, err
	}
	priority, err := decodeChatPriority(values[chatPriorityKey])
	if err != nil {
		return result, err
	}
	result = settings{
		tuning: tuning, chatPriority: priority, minSimilarity: tuning.MinSimilarity,
	}
	// Credentials are intentionally absent from fingerprints and public responses.
	// Verified routes in one explicit space share indexes. Legacy single-route profiles keep their URL identity.
	indexIdentity := s.providers.EmbeddingBaseURL
	if s.providers.EmbeddingSpaceID != "" {
		indexIdentity = "space:" + s.providers.EmbeddingSpaceID
	}
	result.profile = fingerprint(struct {
		Version, Chunker, Encoding, Model, URL                string
		Target, Minimum, Maximum, Overlap, Parent, Dimensions int
		Adaptive                                              bool
	}{tuning.IndexVersion, infrarag.ChunkerVersion, infrarag.TokenEncoding, s.providers.EmbeddingModel,
		indexIdentity, tuning.ChunkTargetTokens, tuning.ChunkMinTokens, tuning.ChunkMaxTokens,
		tuning.ChunkOverlapTokens, tuning.ParentMaxTokens, s.providers.EmbeddingDimensions, tuning.AdaptiveChunkingEnabled})
	return result, nil
}

func fingerprint(value any) string {
	data, _ := json.Marshal(value)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
