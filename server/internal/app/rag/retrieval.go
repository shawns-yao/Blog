package rag

import (
	"context"
	"fmt"
	"strings"
	"time"

	domain "github.com/shawns-yao/shawn-blog/server/internal/domain/rag"
	infraai "github.com/shawns-yao/shawn-blog/server/internal/infra/ai"
	infrarag "github.com/shawns-yao/shawn-blog/server/internal/infra/rag"
)

type queryFailure struct{ reason, message string }

func (s *Service) retrieveEvidence(ctx context.Context, settings settings, plan queryPlan, contentKind string, run *domain.QueryRun, trace *domain.QueryTrace) ([]domain.Evidence, bool, *queryFailure) {
	stats, err := s.repo.Stats(ctx, settings.profile)
	if err != nil {
		return nil, false, &queryFailure{"index_unavailable", "索引服务暂时不可用，请稍后重试。"}
	}
	if stats.Chunks == 0 {
		return nil, false, &queryFailure{"index_not_ready", "公开内容索引尚未就绪，请稍后重试。"}
	}
	if plan.Intent == domain.IntentDocumentSearch {
		started := time.Now()
		matches, err := s.repo.DiscoverDocuments(ctx, settings.profile, discoveryTitlePattern(plan.Query), contentKind, settings.tuning.TopK)
		run.RetrievalMs = elapsedMs(started)
		if err != nil {
			return nil, false, &queryFailure{"retrieval_unavailable", "文档检索暂时不可用，请稍后重试。"}
		}
		if len(matches) > 0 {
			valid, err := s.repo.Validate(ctx, settings.profile, matches)
			if err != nil || !valid {
				return nil, false, &queryFailure{"source_changed", "来源内容正在更新，请稍后重试。"}
			}
			trace.KeywordCandidates, trace.EvidenceCount = len(matches), len(matches)
			return matches, true, nil
		}
	}
	embedCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	started := time.Now()
	vectors, embedErr := settings.embedder.BatchEmbed(embedCtx, []string{plan.Query})
	run.EmbeddingMs = elapsedMs(started)
	cancel()
	queryVector := []float64{}
	if embedErr == nil {
		if len(vectors) != 1 || stats.EmbeddingDimension != len(vectors[0]) {
			return nil, false, &queryFailure{"embedding_dimension_changed", "嵌入模型维度与索引不一致，需要重建索引。"}
		}
		queryVector = vectors[0]
	} else {
		trace.EmbeddingDegraded = true
	}
	started = time.Now()
	vector, keyword, err := s.repo.Retrieve(ctx, settings.profile, plan.Query, contentKind, queryVector, settings.tuning)
	elapsed := time.Since(started).Milliseconds()
	if run.RetrievalMs != nil {
		elapsed += *run.RetrievalMs
	}
	run.RetrievalMs = &elapsed
	if err != nil {
		return nil, false, &queryFailure{"retrieval_unavailable", "检索服务暂时不可用，请稍后重试。"}
	}
	trace.VectorCandidates, trace.KeywordCandidates = len(vector), len(keyword)
	if embedErr != nil && len(keyword) == 0 {
		return nil, false, &queryFailure{"embedding_unavailable", "嵌入服务暂时不可用，请稍后重试。"}
	}
	tuning := settings.tuning
	if embedErr != nil {
		tuning.RRFVectorWeight, tuning.RRFKeywordWeight = 0, 1
	}
	candidates := infrarag.Fuse(vector, keyword, tuning)
	trace.FusedCandidates = len(candidates)
	if len(candidates) > 0 {
		valid, err := s.repo.Validate(ctx, settings.profile, candidates)
		if err != nil || !valid {
			return nil, false, &queryFailure{"source_changed", "来源内容正在更新，请稍后重试。"}
		}
	}
	if tuning.RerankEnabled && len(candidates) > 0 {
		ranked, rerankErr := []domain.Evidence(nil), infraai.ErrRerankUnavailable
		if settings.reranker != nil {
			started = time.Now()
			ranked, rerankErr = settings.reranker.Rerank(ctx, plan.Query, candidates, tuning.RerankThreshold)
			run.RerankMs = elapsedMs(started)
		}
		if rerankErr == nil {
			candidates = ranked
		} else if tuning.RerankFallback {
			run.RerankDegraded = true
		} else {
			return nil, false, &queryFailure{"rerank_unavailable", "重排序服务暂时不可用，请稍后重试。"}
		}
	}
	trace.RerankedCandidates = len(candidates)
	evidence := infrarag.SelectEvidence(candidates, tuning.TopK)
	trace.EvidenceCount = len(evidence)
	if len(evidence) > 0 {
		valid, err := s.repo.Validate(ctx, settings.profile, evidence)
		if err != nil || !valid {
			return nil, false, &queryFailure{"source_changed", "来源内容正在更新，请稍后重试。"}
		}
	}
	return evidence, false, nil
}

func catalogAnswer(evidence []domain.Evidence, profile string) (domain.Answer, error) {
	var lines []string
	answer := domain.Answer{Status: "answered", Mode: "grounded", IndexVersion: profile, Citations: []domain.Citation{}}
	for i, item := range evidence {
		lines = append(lines, fmt.Sprintf("%s [%d]", item.Title, i+1))
		citation, err := sourceCitation(i+1, item)
		if err != nil {
			return domain.Answer{}, err
		}
		answer.Citations = append(answer.Citations, citation)
	}
	answer.Answer = "找到这些相关文档：\n" + strings.Join(lines, "\n")
	return answer, nil
}
