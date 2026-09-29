package rag

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"

	appsysconfig "github.com/shawns-yao/shawn-blog/server/internal/app/sysconfig"
	domainconfig "github.com/shawns-yao/shawn-blog/server/internal/domain/config"
	domain "github.com/shawns-yao/shawn-blog/server/internal/domain/rag"
)

var tuningKeys = []string{
	"rag.chunkSize", "rag.chunkOverlap", "rag.indexVersion", "rag.minSimilarity",
	"rag.vectorTopK", "rag.keywordTopK", "rag.topK", "rag.rrfK", "rag.rrfVectorWeight", "rag.rrfKeywordWeight",
	"rag.rerankEnabled", "rag.rerankCandidateTopK", "rag.rerankThreshold", "rag.rerankFallback",
}

var configKeys = append(append([]string{}, tuningKeys...), chatPriorityKey)

func defaultTuning() domain.Tuning {
	return domain.Tuning{ChunkSize: 1200, ChunkOverlap: 120, IndexVersion: "1", MinSimilarity: 0.35,
		VectorTopK: 20, KeywordTopK: 20, TopK: 6, RRFK: 60, RRFVectorWeight: 0.7, RRFKeywordWeight: 0.3,
		RerankEnabled: true, RerankCandidateTopK: 40, RerankThreshold: 0.2, RerankFallback: true}
}

func validateTuning(t domain.Tuning) error {
	if t.ChunkSize < 200 || t.ChunkSize > 2000 || t.ChunkOverlap < 0 || t.ChunkOverlap >= t.ChunkSize {
		return fmt.Errorf("分块大小需为 200–2000 个字符，重叠大小需小于分块大小。")
	}
	if strings.TrimSpace(t.IndexVersion) == "" || utf8.RuneCountInString(t.IndexVersion) > 64 {
		return fmt.Errorf("索引版本需为 1–64 个字符。")
	}
	if t.VectorTopK < 1 || t.VectorTopK > 100 || t.KeywordTopK < 1 || t.KeywordTopK > 100 ||
		t.TopK < 1 || t.TopK > 20 || t.RerankCandidateTopK < t.TopK || t.RerankCandidateTopK > 100 ||
		t.RerankCandidateTopK > t.VectorTopK+t.KeywordTopK {
		return fmt.Errorf("召回 TopK 需为 1–100，最终 TopK 需为 1–20；融合候选数需介于最终 TopK 与召回总数之间，最多 100。")
	}
	if t.RRFK < 1 || t.RRFK > 200 || t.RRFVectorWeight < 0 || t.RRFKeywordWeight < 0 ||
		math.Abs(t.RRFVectorWeight+t.RRFKeywordWeight-1) > 0.000001 {
		return fmt.Errorf("RRF K 需为 1–200，两路权重需非负且合计为 1。")
	}
	for _, n := range []float64{t.MinSimilarity, t.RRFVectorWeight, t.RRFKeywordWeight, t.RerankThreshold} {
		if math.IsNaN(n) || math.IsInf(n, 0) {
			return fmt.Errorf("阈值和权重必须为有限数值。")
		}
	}
	if t.MinSimilarity < 0 || t.MinSimilarity > 1 || t.RerankThreshold < -10 || t.RerankThreshold > 10 {
		return fmt.Errorf("向量阈值需为 0–1，重排序阈值需为 -10–10。")
	}
	return nil
}

func decodeTuning(values map[string]string) (domain.Tuning, error) {
	t := defaultTuning()
	data, _ := json.Marshal(t)
	fields := make(map[string]json.RawMessage)
	_ = json.Unmarshal(data, &fields)
	for _, key := range tuningKeys {
		value := strings.TrimSpace(values[key])
		if value == "" {
			continue
		}
		field := strings.TrimPrefix(key, "rag.")
		if field == "indexVersion" {
			fields[field], _ = json.Marshal(value)
		} else {
			fields[field] = json.RawMessage(value)
		}
	}
	data, err := json.Marshal(fields)
	if err != nil {
		return t, errNotConfigured
	}
	if json.Unmarshal(data, &t) != nil || validateTuning(t) != nil {
		return t, errNotConfigured
	}
	return t, nil
}

func (s *Service) AdminSettings(ctx context.Context) (domain.AdminSettings, error) {
	settings, err := s.loadTuning(ctx)
	if err != nil {
		return domain.AdminSettings{}, err
	}
	p := s.providers
	channels := make([]domain.ChatChannel, 0, 5)
	primary := p.GPT
	foundPrimary := false
	for i, channel := range s.orderedChatChannels(settings.chatPriority) {
		configured := chatConfigured(channel)
		var extra struct {
			ReasoningEffort string `json:"reasoning_effort"`
		}
		_ = json.Unmarshal([]byte(channel.ExtraBodyJSON), &extra)
		switch extra.ReasoningEffort {
		case "none", "minimal", "low", "medium", "high", "xhigh":
		default:
			extra.ReasoningEffort = ""
		}
		channels = append(channels, domain.ChatChannel{Name: channel.Name, Model: channel.Model, Protocol: channel.Protocol,
			Priority: i + 1, Configured: configured, Default: channel.Name == p.Fallback.Name, ReasoningEffort: extra.ReasoningEffort})
		if configured && !foundPrimary {
			primary, foundPrimary = channel, true
		}
	}
	_, rerankErr := s.newReranker()
	return domain.AdminSettings{Tuning: settings.tuning, Enabled: p.Enabled, ChatChannels: channels,
		PrimaryModel: primary.Model, FallbackModel: p.Fallback.Model, EmbeddingModel: p.EmbeddingModel, RerankModel: p.RerankModel,
		PrimaryConfigured: foundPrimary, FallbackConfigured: chatConfigured(p.Fallback),
		EmbeddingConfigured: s.embeddingConfigured(), RerankConfigured: rerankErr == nil}, nil
}

func (s *Service) UpdateTuning(ctx context.Context, tuning domain.Tuning) (domain.AdminSettings, error) {
	if err := validateTuning(tuning); err != nil {
		return domain.AdminSettings{}, err
	}
	data, _ := json.Marshal(tuning)
	fields := make(map[string]json.RawMessage)
	_ = json.Unmarshal(data, &fields)
	writer, ok := s.config.(interface {
		UpdateConfigs(context.Context, []appsysconfig.UpdateItem) ([]domainconfig.SysConfig, error)
	})
	if !ok {
		return domain.AdminSettings{}, fmt.Errorf("配置服务不支持写入。")
	}
	items := make([]appsysconfig.UpdateItem, 0, len(tuningKeys))
	for _, key := range tuningKeys {
		field := strings.TrimPrefix(key, "rag.")
		value := fields[field]
		valueType := "number"
		if field == "indexVersion" {
			valueType = "string"
		} else if field == "rerankEnabled" || field == "rerankFallback" {
			valueType = "bool"
		} else if field == "minSimilarity" || field == "rrfVectorWeight" || field == "rrfKeywordWeight" || field == "rerankThreshold" {
			valueType = "string"
			number, _ := strconv.ParseFloat(string(value), 64)
			value, _ = json.Marshal(strconv.FormatFloat(number, 'f', -1, 64))
		}
		group := "rag"
		items = append(items, appsysconfig.UpdateItem{Key: key, Value: &value, ValueType: &valueType, GroupPath: &group})
	}
	if _, err := writer.UpdateConfigs(ctx, items); err != nil {
		return domain.AdminSettings{}, fmt.Errorf("检索配置保存失败。")
	}
	return s.AdminSettings(ctx)
}
