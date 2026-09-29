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
	queries := plan.Queries
	if len(queries) == 0 {
		queries = []string{plan.Query}
	}
	vectors, embedErr := settings.embedder.BatchEmbed(embedCtx, queries)
	run.EmbeddingMs = elapsedMs(started)
	cancel()
	if embedErr == nil {
		if len(vectors) != len(queries) {
			return nil, false, &queryFailure{"embedding_dimension_changed", "嵌入模型维度与索引不一致，需要重建索引。"}
		}
		for _, vector := range vectors {
			if stats.EmbeddingDimension != len(vector) {
				return nil, false, &queryFailure{"embedding_dimension_changed", "嵌入模型维度与索引不一致，需要重建索引。"}
			}
		}
	} else {
		trace.EmbeddingDegraded = true
		vectors = nil
	}
	started = time.Now()
	vector, keyword, err := s.repo.Retrieve(ctx, settings.profile, queries, contentKind, vectors, settings.tuning)
	elapsed := time.Since(started).Milliseconds()
	if run.RetrievalMs != nil {
		elapsed += *run.RetrievalMs
	}
	run.RetrievalMs = &elapsed
	if err != nil {
		return nil, false, &queryFailure{"retrieval_unavailable", "检索服务暂时不可用，请稍后重试。"}
	}
	for _, list := range vector {
		trace.VectorCandidates += len(list)
	}
	for _, list := range keyword {
		trace.KeywordCandidates += len(list)
	}
	captureEvaluationStage(ctx, "vector", queries, vector)
	captureEvaluationStage(ctx, "keyword", queries, keyword)
	if embedErr != nil && trace.KeywordCandidates == 0 {
		return nil, false, &queryFailure{"embedding_unavailable", "嵌入服务暂时不可用，请稍后重试。"}
	}
	tuning := settings.tuning
	if embedErr != nil {
		tuning.RRFVectorWeight, tuning.RRFKeywordWeight = 0, 1
	}
	candidates := infrarag.FuseMany(vector, keyword, tuning)
	trace.FusedCandidates = len(candidates)
	captureEvaluationStage(ctx, "fused", nil, [][]domain.Evidence{candidates})
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
	captureEvaluationStage(ctx, "reranked", nil, [][]domain.Evidence{candidates})
	limit := adaptiveEvidenceLimit(tuning, plan, candidates, tuning.RerankEnabled && !run.RerankDegraded)
	trace.DynamicTopK = tuning.DynamicTopKEnabled
	trace.TopKMinimum, trace.TopKMaximum, trace.TopKTarget, trace.TopKReason = limit.Minimum, limit.Maximum, limit.Target, limit.Reason
	if plan.Strategy == "MULTI_HOP" && tuning.MultiQueryEnabled && len(queries) > 1 {
		candidates, trace.SubqueryAnchors = prioritizeSubqueryEvidence(candidates, vector, keyword, tuning)
		captureEvaluationStage(ctx, "coverage", queries, [][]domain.Evidence{candidates})
	}
	contextStarted := time.Now()
	evidence, stoppedBy, contextErr := s.buildContext(ctx, settings, plan, candidates, limit, trace.SubqueryAnchors)
	trace.TopKStoppedBy = stoppedBy
	trace.ContextMs = time.Since(contextStarted).Milliseconds()
	if contextErr != nil {
		return nil, false, &queryFailure{"source_changed", "来源内容正在更新，请稍后重试。"}
	}
	trace.ContextTokens = evidenceTokens(evidence)
	trace.EvidenceCount = len(evidence)
	captureEvaluationStage(ctx, "context", nil, [][]domain.Evidence{evidence})
	if len(evidence) > 0 {
		valid, err := s.repo.Validate(ctx, settings.profile, evidence)
		if err != nil || !valid {
			return nil, false, &queryFailure{"source_changed", "来源内容正在更新，请稍后重试。"}
		}
	}
	return evidence, false, nil
}

// Global reranking can favor the first step and bury another step's evidence.
// Reserve the best already-admitted RRF hit for each derived query, then fill
// from the original reranking. This uses no extra model call or new candidates.
func prioritizeSubqueryEvidence(candidates []domain.Evidence, vector, keyword [][]domain.Evidence, tuning domain.Tuning) ([]domain.Evidence, int) {
	admitted := make(map[int64]domain.Evidence, len(candidates))
	for _, item := range candidates {
		admitted[item.ID] = item
	}
	result := make([]domain.Evidence, 0, len(candidates))
	seen := map[int64]bool{}
	for query := 1; query < max(len(vector), len(keyword)); query++ {
		var vectors, keywords []domain.Evidence
		if query < len(vector) {
			vectors = vector[query]
		}
		if query < len(keyword) {
			keywords = keyword[query]
		}
		for _, hit := range infrarag.Fuse(vectors, keywords, tuning) {
			if item, ok := admitted[hit.ID]; ok {
				if !seen[item.ID] {
					result = append(result, item)
					seen[item.ID] = true
				}
				break
			}
		}
	}
	anchors := len(result)
	for _, item := range candidates {
		if !seen[item.ID] {
			result = append(result, item)
		}
	}
	return result, anchors
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
