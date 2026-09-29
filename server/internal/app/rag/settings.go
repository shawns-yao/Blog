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
	embedder      *infraai.Embedder
	reranker      *infraai.Reranker
	tuning        domain.Tuning
	chunkSize     int
	overlap       int
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
		if channel.APIKey == "" {
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
	return result, nil
}

func (s *Service) embeddingConfigured() bool {
	p := s.providers
	parsed, err := url.Parse(p.EmbeddingBaseURL)
	return err == nil && parsed.Host != "" && (parsed.Scheme == "https" || parsed.Scheme == "http") &&
		parsed.User == nil && parsed.RawQuery == "" && parsed.Fragment == "" &&
		p.EmbeddingModel != "" && p.EmbeddingAPIKey != "" && p.EmbeddingDimensions >= 0 && p.EmbeddingDimensions <= 16384
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
	p := s.providers
	result.embedder = infraai.NewEmbedder(p.EmbeddingBaseURL, p.EmbeddingAPIKey, p.EmbeddingModel, p.EmbeddingDimensions)
	return result, nil
}

func chatConfigured(channel appconfig.RAGChatConfig) bool {
	_, err := infraai.NewRAGChatClient(channel.BaseURL, channel.APIKey, channel.HeadersJSON, channel.ExtraBodyJSON, channel.SessionHeader, channel.Timeout, channel.Protocol)
	return err == nil && channel.Model != ""
}

func (s *Service) newReranker() (*infraai.Reranker, error) {
	p := s.providers
	return infraai.NewReranker(p.RerankBaseURL, p.RerankModel, p.RerankAPIKey, p.RerankTimeout)
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
		tuning: tuning, chatPriority: priority, chunkSize: tuning.ChunkSize, overlap: tuning.ChunkOverlap, minSimilarity: tuning.MinSimilarity,
	}
	// Credentials are intentionally absent from fingerprints and public responses.
	result.profile = fingerprint(struct {
		Version, Chunker, Model, URL string
		Size, Overlap, Dimensions    int
	}{tuning.IndexVersion, infrarag.ChunkerVersion, s.providers.EmbeddingModel,
		s.providers.EmbeddingBaseURL, tuning.ChunkSize, tuning.ChunkOverlap, s.providers.EmbeddingDimensions})
	return result, nil
}

func fingerprint(value any) string {
	data, _ := json.Marshal(value)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
