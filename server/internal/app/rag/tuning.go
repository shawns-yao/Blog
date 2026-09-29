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
	infrarag "github.com/shawns-yao/shawn-blog/server/internal/infra/rag"
)

var tuningKeys = []string{
	"rag.chunkSize", "rag.chunkOverlap", "rag.indexVersion", "rag.minSimilarity",
	"rag.vectorTopK", "rag.keywordTopK", "rag.topK", "rag.rrfK", "rag.rrfVectorWeight", "rag.rrfKeywordWeight",
	"rag.rerankEnabled", "rag.rerankCandidateTopK", "rag.rerankThreshold", "rag.rerankFallback",
	"rag.chunkTargetTokens", "rag.chunkMinTokens", "rag.chunkMaxTokens", "rag.chunkOverlapTokens", "rag.parentMaxTokens",
	"rag.contextMaxTokens", "rag.historyMaxTokens", "rag.multiQueryEnabled", "rag.multiQueryMax", "rag.bm25K1", "rag.bm25B",
	"rag.dynamicTopKEnabled", "rag.dynamicTopKMin", "rag.dynamicTopKMax",
	"rag.adaptiveChunkingEnabled", "rag.adaptiveRetrievalEnabled", "rag.evidenceSelectionEnabled", "rag.evidenceDiversityWeight",
}

var configKeys = append(append([]string{}, tuningKeys...), chatPriorityKey)

func defaultTuning() domain.Tuning {
	return domain.Tuning{ChunkSize: 1200, ChunkOverlap: 120, IndexVersion: "1", MinSimilarity: 0.35,
		ChunkTargetTokens: 500, ChunkMinTokens: 180, ChunkMaxTokens: 800, ChunkOverlapTokens: 60, ParentMaxTokens: 1600,
		ContextMaxTokens: 6000, HistoryMaxTokens: 3000, MultiQueryEnabled: true, MultiQueryMax: 3, BM25K1: 1.2, BM25B: 0.75,
		VectorTopK: 25, KeywordTopK: 25, TopK: 6, RRFK: 60, RRFVectorWeight: 0.7, RRFKeywordWeight: 0.3,
		DynamicTopKEnabled: true, DynamicTopKMin: 2, DynamicTopKMax: 12,
		AdaptiveChunkingEnabled: true, AdaptiveRetrievalEnabled: true, EvidenceSelectionEnabled: true, EvidenceDiversityWeight: 0.2,
		RerankEnabled: true, RerankCandidateTopK: 40, RerankThreshold: 0.2, RerankFallback: true}
}

func validateTuning(t domain.Tuning) error {
	if t.ChunkMinTokens < 1 || t.ChunkTargetTokens < 100 || t.ChunkMaxTokens > 4000 ||
		t.ChunkMinTokens > t.ChunkTargetTokens || t.ChunkTargetTokens > t.ChunkMaxTokens ||
		t.ChunkOverlapTokens < 0 || t.ChunkOverlapTokens >= t.ChunkTargetTokens ||
		t.ParentMaxTokens < t.ChunkMaxTokens || t.ParentMaxTokens > 8000 {
		return fmt.Errorf("子块需满足最小 ≤ 目标 ≤ 上限，目标至少 100 token，上限最多 4000；重叠小于目标，父块上限介于子块上限与 8000 之间。")
	}
	if t.ContextMaxTokens < t.ChunkMaxTokens || t.ContextMaxTokens > 16000 || t.HistoryMaxTokens < 0 || t.HistoryMaxTokens > 8000 ||
		t.MultiQueryMax < 1 || t.MultiQueryMax > 3 || t.BM25K1 <= 0 || t.BM25K1 > 3 || t.BM25B < 0 || t.BM25B > 1 {
		return fmt.Errorf("证据预算需介于子块上限与 16000 token；历史预算为 0–8000，子查询上限为 1–3，BM25 K1 为 (0,3]、B 为 [0,1]。")
	}
	if t.ChunkSize < 200 || t.ChunkSize > 2000 || t.ChunkOverlap < 0 || t.ChunkOverlap >= t.ChunkSize {
		return fmt.Errorf("分块大小需为 200–2000 个字符，重叠大小需小于分块大小。")
	}
	if strings.TrimSpace(t.IndexVersion) == "" || utf8.RuneCountInString(t.IndexVersion) > 64 {
		return fmt.Errorf("索引版本需为 1–64 个字符。")
	}
	if t.VectorTopK < 1 || t.VectorTopK > 100 || t.KeywordTopK < 1 || t.KeywordTopK > 100 ||
		t.TopK < 1 || t.TopK > 20 || t.RerankCandidateTopK < t.TopK || t.RerankCandidateTopK > 100 ||
		t.RerankCandidateTopK > (t.VectorTopK+t.KeywordTopK)*(1+t.MultiQueryMax) {
		return fmt.Errorf("召回 TopK 需为 1–100，最终 TopK 需为 1–20；融合候选数需介于最终 TopK 与多查询召回总数之间，最多 100。")
	}
	if t.DynamicTopKMin < 1 || t.DynamicTopKMax > 20 || t.DynamicTopKMin > t.DynamicTopKMax ||
		(t.DynamicTopKEnabled && t.DynamicTopKMax > t.RerankCandidateTopK) {
		return fmt.Errorf("动态 TopK 需满足 1 ≤ 下限 ≤ 上限 ≤ 20；启用时上限不能超过融合候选数。")
	}
	if t.RRFK < 1 || t.RRFK > 200 || t.RRFVectorWeight < 0 || t.RRFKeywordWeight < 0 ||
		math.Abs(t.RRFVectorWeight+t.RRFKeywordWeight-1) > 0.000001 {
		return fmt.Errorf("RRF K 需为 1–200，两路权重需非负且合计为 1。")
	}
	for _, n := range []float64{t.MinSimilarity, t.RRFVectorWeight, t.RRFKeywordWeight, t.RerankThreshold, t.BM25K1, t.BM25B, t.EvidenceDiversityWeight} {
		if math.IsNaN(n) || math.IsInf(n, 0) {
			return fmt.Errorf("阈值和权重必须为有限数值。")
		}
	}
	if t.EvidenceDiversityWeight < 0 || t.EvidenceDiversityWeight > 1 {
		return fmt.Errorf("证据多样性权重需为 0–1。")
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
	if json.Unmarshal(data, &t) != nil {
		return t, errNotConfigured
	}
	// Older installations can have a candidate pool smaller than the new defaults.
	if strings.TrimSpace(values["rag.dynamicTopKMax"]) == "" {
		t.DynamicTopKMax = min(t.DynamicTopKMax, t.RerankCandidateTopK)
	}
	if strings.TrimSpace(values["rag.dynamicTopKMin"]) == "" {
		t.DynamicTopKMin = min(t.DynamicTopKMin, t.DynamicTopKMax)
	}
	if validateTuning(t) != nil {
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
	return domain.AdminSettings{Tuning: settings.tuning, Enabled: p.Enabled, ChatChannels: channels, TokenEncoding: infrarag.TokenEncoding,
		PrimaryModel: primary.Model, FallbackModel: p.Fallback.Model, EmbeddingModel: p.EmbeddingModel, RerankModel: p.RerankModel,
		PrimaryConfigured: foundPrimary, FallbackConfigured: chatConfigured(p.Fallback),
		EmbeddingConfigured: s.embeddingConfigured(), RerankConfigured: rerankErr == nil}, nil
}

func (s *Service) UpdateTuning(ctx context.Context, tuning domain.Tuning) (domain.AdminSettings, error) {
	// Old clients omit the new fields and retain the fixed mode they requested.
	if tuning.DynamicTopKMin == 0 && tuning.DynamicTopKMax == 0 {
		tuning.DynamicTopKMax = min(defaultTuning().DynamicTopKMax, tuning.RerankCandidateTopK)
		tuning.DynamicTopKMin = min(defaultTuning().DynamicTopKMin, tuning.DynamicTopKMax)
	}
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
		} else if strings.HasSuffix(field, "Enabled") || field == "rerankFallback" {
			valueType = "bool"
		} else if field == "minSimilarity" || field == "rrfVectorWeight" || field == "rrfKeywordWeight" || field == "rerankThreshold" || field == "bm25K1" || field == "bm25B" || field == "evidenceDiversityWeight" {
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
