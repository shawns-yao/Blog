package persistence

import (
	"context"

	domain "github.com/shawns-yao/shawn-blog/server/internal/domain/rag"
)

// All status and chunk counts use one grouped join, including source edits and
// profile changes not yet reconciled by the worker. Private sources have no live chunks.
const ragAdminBase = `WITH live AS (
    SELECT c.moment_id, count(*) AS chunks,
        min(cardinality(c.embedding)) AS min_dimension, max(cardinality(c.embedding)) AS max_dimension
    FROM rag_chunk c ` + ragLiveIndex + ` GROUP BY c.moment_id
), base AS (
    SELECT COALESCE(m.id, s.moment_id) AS moment_id, COALESCE(m.title, '已移除的文档') AS title,
        COALESCE(m.ext_info->>'contentKind', '') AS content_kind,
        COALESCE(m.is_published AND m.deleted_at IS NULL, FALSE) AS published,
        CASE WHEN NOT COALESCE(` + ragEligible + `, FALSE) THEN 'excluded'
             WHEN s.moment_id IS NULL THEN 'unindexed'
             WHEN (s.desired_profile <> '' AND s.desired_profile <> ?) OR
                  s.source_hash IS DISTINCT FROM ` + ragCurrentHash + ` THEN 'outdated'
             ELSE s.status END AS status,
        COALESCE(l.chunks, 0) AS chunks, l.min_dimension, l.max_dimension,
        COALESCE(s.attempts, 0) AS attempts, COALESCE(s.last_error, '') AS last_error,
        s.updated_at, s.indexed_at, s.index_duration_ms
    FROM moment m FULL JOIN rag_index_state s ON s.moment_id = m.id
    LEFT JOIN live l ON l.moment_id = m.id
    WHERE s.moment_id IS NOT NULL OR m.ext_info->>'contentKind' IN ('article', 'note')
)
`

const ragDocumentFilter = ` WHERE (? = '' OR position(lower(?) IN lower(title)) > 0)
    AND (? = '' OR status = ?) AND (? = '' OR content_kind = ?)`

func (r *RAGRepository) Documents(ctx context.Context, profile string, filter domain.DocumentFilter) (domain.DocumentPage, error) {
	result := domain.DocumentPage{Items: []domain.Document{}, Page: filter.Page, PageSize: filter.PageSize}
	params := []any{profile, profile, filter.Search, filter.Search, filter.Status, filter.Status, filter.ContentKind, filter.ContentKind}
	if err := r.db.WithContext(ctx).Raw(ragAdminBase+`SELECT count(*) FROM base`+ragDocumentFilter, params...).Scan(&result.Total).Error; err != nil {
		return result, err
	}
	params = append(params, filter.PageSize, (filter.Page-1)*filter.PageSize)
	err := r.db.WithContext(ctx).Raw(ragAdminBase+`SELECT moment_id, title, content_kind, status, published, chunks,
attempts, last_error, updated_at, indexed_at, index_duration_ms FROM base`+ragDocumentFilter+`
ORDER BY updated_at DESC NULLS LAST, moment_id DESC LIMIT ? OFFSET ?`, params...).Scan(&result.Items).Error
	return result, err
}

func (r *RAGRepository) DocumentChunks(ctx context.Context, profile string, momentID int64, page, pageSize int) ([]domain.Chunk, int64, error) {
	chunks := make([]domain.Chunk, 0)
	var total int64
	if err := r.db.WithContext(ctx).Raw(`SELECT count(*) FROM rag_chunk c `+ragLiveIndex+` AND c.moment_id = ?`, profile, momentID).Scan(&total).Error; err != nil {
		return nil, 0, err
	}
	rows, err := r.db.WithContext(ctx).Raw(`SELECT c.seq, c.content, c.context_header, c.kind, c.start_at AS start, c.end_at AS "end"
FROM rag_chunk c `+ragLiveIndex+` AND c.moment_id = ? ORDER BY c.seq LIMIT ? OFFSET ?`, profile, momentID, pageSize, (page-1)*pageSize).Rows()
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	for rows.Next() {
		var chunk domain.Chunk
		if err := rows.Scan(&chunk.Seq, &chunk.Content, &chunk.ContextHeader, &chunk.Kind, &chunk.Start, &chunk.End); err != nil {
			return nil, 0, err
		}
		chunks = append(chunks, chunk)
	}
	return chunks, total, rows.Err()
}

func (r *RAGRepository) ReindexDocument(ctx context.Context, profile string, momentID int64) (bool, error) {
	result := r.db.WithContext(ctx).Exec(`INSERT INTO rag_index_state (moment_id, source_hash, desired_profile)
SELECT m.id, `+ragCurrentHash+`, ? FROM moment m WHERE m.id = ? AND `+ragEligible+` FOR SHARE OF m
ON CONFLICT (moment_id) DO UPDATE SET source_hash = EXCLUDED.source_hash, desired_profile = EXCLUDED.desired_profile,
revision = rag_index_state.revision + 1, status = 'pending', attempts = 0, next_attempt_at = now(),
lease_token = NULL, lease_until = NULL, last_error = NULL, updated_at = now()`, profile, momentID)
	return result.RowsAffected > 0, result.Error
}

func durationMetric(value *int64) (int, int64) {
	if value == nil {
		return 0, 0
	}
	return 1, *value
}

func flagMetric(value bool) int {
	if value {
		return 1
	}
	return 0
}

func (r *RAGRepository) RecordQuery(ctx context.Context, run domain.QueryRun) error {
	embedCalls, embedMs := durationMetric(run.EmbeddingMs)
	retrieveCalls, retrieveMs := durationMetric(run.RetrievalMs)
	rerankCalls, rerankMs := durationMetric(run.RerankMs)
	generateCalls, generateMs := durationMetric(run.GenerationMs)
	return r.db.WithContext(ctx).Exec(`INSERT INTO rag_query_metric
(day, outcome, reason, requests, duration_ms, embedding_calls, embedding_ms, retrieval_calls, retrieval_ms,
 rerank_calls, rerank_ms, generation_calls, generation_ms, primary_failures, fallback_requests, rerank_degraded)
VALUES ((CURRENT_TIMESTAMP AT TIME ZONE 'UTC')::date, ?, ?, 1, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT (day, outcome, reason) DO UPDATE SET
requests = rag_query_metric.requests + 1, duration_ms = rag_query_metric.duration_ms + EXCLUDED.duration_ms,
embedding_calls = rag_query_metric.embedding_calls + EXCLUDED.embedding_calls,
embedding_ms = rag_query_metric.embedding_ms + EXCLUDED.embedding_ms,
retrieval_calls = rag_query_metric.retrieval_calls + EXCLUDED.retrieval_calls,
retrieval_ms = rag_query_metric.retrieval_ms + EXCLUDED.retrieval_ms,
rerank_calls = rag_query_metric.rerank_calls + EXCLUDED.rerank_calls,
rerank_ms = rag_query_metric.rerank_ms + EXCLUDED.rerank_ms,
generation_calls = rag_query_metric.generation_calls + EXCLUDED.generation_calls,
generation_ms = rag_query_metric.generation_ms + EXCLUDED.generation_ms,
primary_failures = rag_query_metric.primary_failures + EXCLUDED.primary_failures,
fallback_requests = rag_query_metric.fallback_requests + EXCLUDED.fallback_requests,
rerank_degraded = rag_query_metric.rerank_degraded + EXCLUDED.rerank_degraded`,
		run.Status, run.Reason, run.DurationMs, embedCalls, embedMs, retrieveCalls, retrieveMs, rerankCalls, rerankMs,
		generateCalls, generateMs, flagMetric(run.PrimaryFailed), flagMetric(run.UsedFallback), flagMetric(run.RerankDegraded)).Error
}

func (r *RAGRepository) QueryMetrics(ctx context.Context, days int) (domain.QueryMetrics, error) {
	result := domain.QueryMetrics{Days: days, Failures: []domain.FailureCount{}}
	const window = ` FROM rag_query_metric WHERE day >= (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')::date - (?::int - 1)`
	err := r.db.WithContext(ctx).Raw(`SELECT COALESCE(sum(requests), 0) AS requests,
COALESCE(sum(requests) FILTER (WHERE outcome = 'answered'), 0) AS answered,
COALESCE(sum(requests) FILTER (WHERE outcome = 'no_evidence'), 0) AS no_evidence,
COALESCE(sum(requests) FILTER (WHERE outcome = 'temporarily_unavailable'), 0) AS unavailable,
COALESCE(sum(primary_failures), 0) AS primary_failures, COALESCE(sum(fallback_requests), 0) AS fallback_requests,
COALESCE(sum(rerank_degraded), 0) AS rerank_degraded,
sum(duration_ms)::double precision / NULLIF(sum(requests), 0) AS avg_duration_ms,
sum(embedding_ms)::double precision / NULLIF(sum(embedding_calls), 0) AS avg_embedding_ms,
sum(retrieval_ms)::double precision / NULLIF(sum(retrieval_calls), 0) AS avg_retrieval_ms,
sum(rerank_ms)::double precision / NULLIF(sum(rerank_calls), 0) AS avg_rerank_ms,
sum(generation_ms)::double precision / NULLIF(sum(generation_calls), 0) AS avg_generation_ms`+window, days).Row().Scan(
		&result.Requests, &result.Answered, &result.NoEvidence, &result.Unavailable,
		&result.PrimaryFailures, &result.FallbackRequests, &result.RerankDegraded,
		&result.AvgDurationMs, &result.AvgEmbeddingMs, &result.AvgRetrievalMs,
		&result.AvgRerankMs, &result.AvgGenerationMs)
	if err != nil {
		return result, err
	}
	err = r.db.WithContext(ctx).Raw(`SELECT reason, sum(requests) AS count`+window+`
AND outcome = 'temporarily_unavailable' GROUP BY reason ORDER BY count DESC, reason`, days).Scan(&result.Failures).Error
	return result, err
}
