package rag

import (
	"context"
	"time"

	domain "github.com/shawns-yao/shawn-blog/server/internal/domain/rag"
	infraai "github.com/shawns-yao/shawn-blog/server/internal/infra/ai"
	infrarag "github.com/shawns-yao/shawn-blog/server/internal/infra/rag"
)

type queryFailure struct{ reason, message string }

func (s *Service) retrieveEvidence(ctx context.Context, settings settings, plan queryPlan, contentKind string, run *domain.QueryRun, trace *domain.QueryTrace) ([]domain.Evidence, *queryFailure) {
	stats, err := s.repo.Stats(ctx, settings.profile)
	if err != nil {
		return nil, &queryFailure{"index_unavailable", "索引服务暂时不可用，请稍后重试。"}
	}
	if stats.Chunks == 0 {
		return nil, &queryFailure{"index_not_ready", "公开内容索引尚未就绪，请稍后重试。"}
	}
	tuning, policy := createRetrievalPolicy(settings.tuning, plan)
	settings.tuning = tuning
	trace.RetrievalPolicy = &policy
	started := time.Now()
	queries := plan.Queries
	if len(queries) == 0 {
		queries = []string{plan.Query}
	}
	var vectors [][]float64
	var embedErr error
	if tuning.VectorTopK > 0 {
		var provider infraai.ProviderTrace
		vectors, provider, embedErr = settings.embedder.BatchEmbedWithTrace(ctx, infraai.EmbeddingQueries(queries, s.providers.EmbeddingInstruction))
		trace.EmbeddingProvider, trace.EmbeddingAttempts, trace.EmbeddingFailures = provider.Provider, provider.Attempts, provider.Failures
		trace.EmbeddingFallbackUsed = provider.Fallback
		run.EmbeddingMs = elapsedMs(started)
	}
	if embedErr == nil && tuning.VectorTopK > 0 {
		if len(vectors) != len(queries) {
			return nil, &queryFailure{"embedding_dimension_changed", "嵌入模型维度与索引不一致，需要重建索引。"}
		}
		for _, vector := range vectors {
			if stats.EmbeddingDimension != len(vector) {
				return nil, &queryFailure{"embedding_dimension_changed", "嵌入模型维度与索引不一致，需要重建索引。"}
			}
		}
	} else if embedErr != nil {
		trace.EmbeddingDegraded = true
		vectors = nil
	}
	started = time.Now()
	vector, keyword, err := s.repo.Retrieve(ctx, settings.profile, queries, contentKind, vectors, tuning)
	elapsed := time.Since(started).Milliseconds()
	if run.RetrievalMs != nil {
		elapsed += *run.RetrievalMs
	}
	run.RetrievalMs = &elapsed
	if err != nil {
		return nil, &queryFailure{"retrieval_unavailable", "检索服务暂时不可用，请稍后重试。"}
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
		return nil, &queryFailure{"embedding_unavailable", "嵌入服务暂时不可用，请稍后重试。"}
	}
	if embedErr != nil {
		tuning.RRFVectorWeight, tuning.RRFKeywordWeight = 0, 1
	}
	candidates := infrarag.FuseMany(vector, keyword, tuning)
	trace.FusedCandidates = len(candidates)
	captureEvaluationStage(ctx, "fused", nil, [][]domain.Evidence{candidates})
	if len(candidates) > 0 {
		valid, err := s.repo.Validate(ctx, settings.profile, candidates)
		if err != nil || !valid {
			return nil, &queryFailure{"source_changed", "来源内容正在更新，请稍后重试。"}
		}
	}
	if tuning.RerankEnabled && len(candidates) > 0 {
		ranked, rerankErr := []domain.Evidence(nil), infraai.ErrRerankUnavailable
		if settings.reranker != nil {
			started = time.Now()
			var provider infraai.ProviderTrace
			ranked, provider, rerankErr = settings.reranker.Rerank(ctx, plan.Query, candidates, tuning.RerankThreshold)
			trace.RerankProvider, trace.RerankAttempts, trace.RerankFailures = provider.Provider, provider.Attempts, provider.Failures
			trace.RerankFallbackUsed = provider.Fallback
			run.RerankMs = elapsedMs(started)
		}
		if rerankErr == nil {
			candidates = ranked
		} else if tuning.RerankFallback {
			run.RerankDegraded = true
		} else {
			return nil, &queryFailure{"rerank_unavailable", "重排序服务暂时不可用，请稍后重试。"}
		}
	}
	trace.RerankedCandidates = len(candidates)
	captureEvaluationStage(ctx, "reranked", nil, [][]domain.Evidence{candidates})
	limit := adaptiveEvidenceLimit(tuning, plan, candidates, tuning.RerankEnabled && !run.RerankDegraded)
	trace.DynamicTopK = tuning.DynamicTopKEnabled
	trace.TopKMinimum, trace.TopKMaximum, trace.TopKTarget, trace.TopKReason = limit.Minimum, limit.Maximum, limit.Target, limit.Reason
	if (plan.Strategy == "MULTI_HOP" || plan.Strategy == "COMPARE") && tuning.MultiQueryEnabled && len(queries) > 1 {
		candidates, trace.SubqueryAnchors = prioritizeSubqueryEvidence(candidates, vector, keyword, tuning)
		captureEvaluationStage(ctx, "coverage", queries, [][]domain.Evidence{candidates})
	}
	if tuning.EvidenceSelectionEnabled {
		candidates = diversifyEvidence(candidates, trace.SubqueryAnchors, tuning.EvidenceDiversityWeight)
		trace.EvidenceSelection = "rerank_leader+subquery_anchors+lexical_diversity"
		captureEvaluationStage(ctx, "selection", nil, [][]domain.Evidence{candidates})
	} else {
		trace.EvidenceSelection = "rank_order"
	}
	contextStarted := time.Now()
	evidence, stoppedBy, contextErr := s.buildContext(ctx, settings, plan, candidates, limit, trace.SubqueryAnchors)
	trace.TopKStoppedBy = stoppedBy
	trace.ContextMs = time.Since(contextStarted).Milliseconds()
	if contextErr != nil {
		return nil, &queryFailure{"source_changed", "来源内容正在更新，请稍后重试。"}
	}
	trace.ContextTokens = evidenceTokens(evidence)
	trace.EvidenceCount = len(evidence)
	captureEvaluationStage(ctx, "context", nil, [][]domain.Evidence{evidence})
	if len(evidence) > 0 {
		valid, err := s.repo.Validate(ctx, settings.profile, evidence)
		if err != nil || !valid {
			return nil, &queryFailure{"source_changed", "来源内容正在更新，请稍后重试。"}
		}
	}
	return evidence, nil
}

// Global reranking can favor the first step and bury another step's evidence.
// Keep the reranked leader, then reserve an already-admitted RRF hit for each
// derived query. Coverage must not displace the best answer to the full question.
// The protected prefix uses no extra model call or new candidates.
func prioritizeSubqueryEvidence(candidates []domain.Evidence, vector, keyword [][]domain.Evidence, tuning domain.Tuning) ([]domain.Evidence, int) {
	admitted := make(map[int64]domain.Evidence, len(candidates))
	for _, item := range candidates {
		admitted[item.ID] = item
	}
	result := make([]domain.Evidence, 0, len(candidates))
	seen := map[int64]bool{}
	if len(candidates) > 0 {
		result = append(result, candidates[0])
		seen[candidates[0].ID] = true
	}
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
