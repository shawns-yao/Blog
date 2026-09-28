-- +goose Up
-- Index metadata is rebuildable derived data. Daily metrics are operational
-- aggregates, independent of core content: no questions, sessions or secrets.
ALTER TABLE rag_index_state ADD COLUMN indexed_at TIMESTAMPTZ;
ALTER TABLE rag_index_state ADD COLUMN index_duration_ms BIGINT CHECK (index_duration_ms >= 0);

CREATE TABLE rag_query_metric (
    day DATE NOT NULL,
    outcome TEXT NOT NULL CHECK (outcome IN ('answered', 'no_evidence', 'temporarily_unavailable')),
    reason TEXT NOT NULL,
    requests BIGINT NOT NULL DEFAULT 0,
    duration_ms BIGINT NOT NULL DEFAULT 0,
    embedding_calls BIGINT NOT NULL DEFAULT 0,
    embedding_ms BIGINT NOT NULL DEFAULT 0,
    retrieval_calls BIGINT NOT NULL DEFAULT 0,
    retrieval_ms BIGINT NOT NULL DEFAULT 0,
    rerank_calls BIGINT NOT NULL DEFAULT 0,
    rerank_ms BIGINT NOT NULL DEFAULT 0,
    generation_calls BIGINT NOT NULL DEFAULT 0,
    generation_ms BIGINT NOT NULL DEFAULT 0,
    primary_failures BIGINT NOT NULL DEFAULT 0,
    fallback_requests BIGINT NOT NULL DEFAULT 0,
    rerank_degraded BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (day, outcome, reason)
);

INSERT INTO sys_config (config_key, value, is_sensitive, group_path, label, description, value_type, sort, meta)
VALUES
    ('rag.vectorTopK', '20', FALSE, 'rag', '向量召回 TopK', '每次向量检索最多返回的候选分块数', 'number', 10, '{}'::jsonb),
    ('rag.keywordTopK', '20', FALSE, 'rag', '关键词召回 TopK', '每次关键词检索最多返回的候选分块数', 'number', 20, '{}'::jsonb),
    ('rag.topK', '6', FALSE, 'rag', '最终 TopK', '经过重排序与去重后最多送入问答模型的分块数', 'number', 30, '{}'::jsonb),
    ('rag.rrfK', '60', FALSE, 'rag', 'RRF 平滑常数', '加权倒数排名融合的 K，采用 WeKnora 默认值', 'number', 90, '{}'::jsonb),
    ('rag.rrfVectorWeight', '0.7', FALSE, 'rag', 'RRF 向量权重', '与关键词权重之和必须为 1', 'string', 100, '{}'::jsonb),
    ('rag.rrfKeywordWeight', '0.3', FALSE, 'rag', 'RRF 关键词权重', '与向量权重之和必须为 1', 'string', 110, '{}'::jsonb),
    ('rag.rerankEnabled', 'true', FALSE, 'rag', '启用重排序', '使用服务端环境变量中配置的重排序模型', 'bool', 120, '{}'::jsonb),
    ('rag.rerankCandidateTopK', '40', FALSE, 'rag', '融合候选 TopK', '送入重排序的 RRF 候选上限，最多 100', 'number', 130, '{}'::jsonb),
    ('rag.rerankThreshold', '0.2', FALSE, 'rag', '重排序分数下限', '按重排序模型实际分数范围调整', 'string', 140, '{}'::jsonb),
    ('rag.rerankFallback', 'true', FALSE, 'rag', '重排序失败回退', '模型失败时使用 RRF 排名并记录降级次数', 'bool', 150, '{}'::jsonb)
ON CONFLICT (config_key) DO NOTHING;

-- +goose Down
-- Removes only RAG operational aggregates, derived metadata and tuning keys.
DROP TABLE rag_query_metric;
ALTER TABLE rag_index_state DROP COLUMN index_duration_ms;
ALTER TABLE rag_index_state DROP COLUMN indexed_at;
DELETE FROM sys_config WHERE config_key IN (
    'rag.vectorTopK', 'rag.keywordTopK', 'rag.topK', 'rag.rrfK', 'rag.rrfVectorWeight',
    'rag.rrfKeywordWeight', 'rag.rerankEnabled', 'rag.rerankCandidateTopK', 'rag.rerankThreshold', 'rag.rerankFallback'
);
