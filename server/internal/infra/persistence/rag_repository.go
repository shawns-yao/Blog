package persistence

import (
	"context"
	"encoding/json"
	"strings"
	"time"
	"unicode"

	domain "github.com/shawns-yao/shawn-blog/server/internal/domain/rag"
	"gorm.io/gorm"
)

type RAGRepository struct{ db *gorm.DB }

func NewRAGRepository(db *gorm.DB) *RAGRepository { return &RAGRepository{db: db} }

const ragEligible = `m.is_published = TRUE AND m.deleted_at IS NULL
    AND m.ext_info->>'contentKind' IN ('article', 'note')`

const ragCurrentHash = `rag_source_hash(m.title, m.summary, m.content, m.ext_info->>'contentKind', m.short_url)`

func (r *RAGRepository) Reconcile(ctx context.Context, profile string, force bool) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`INSERT INTO rag_index_state (moment_id, source_hash, desired_profile)
SELECT m.id, `+ragCurrentHash+`, ? FROM moment m WHERE `+ragEligible+`
ON CONFLICT (moment_id) DO UPDATE SET
    source_hash = EXCLUDED.source_hash, desired_profile = EXCLUDED.desired_profile,
    revision = rag_index_state.revision + 1, status = 'pending', attempts = 0,
    next_attempt_at = now(), lease_token = NULL, lease_until = NULL, last_error = NULL, updated_at = now()
WHERE rag_index_state.source_hash IS DISTINCT FROM EXCLUDED.source_hash
   OR rag_index_state.desired_profile IS DISTINCT FROM EXCLUDED.desired_profile
   OR rag_index_state.status = 'excluded' OR ?`, profile, force).Error; err != nil {
			return err
		}
		if err := tx.Exec(`UPDATE rag_index_state s SET status = 'excluded', revision = revision + 1,
active_hash = NULL, active_profile = NULL, lease_token = NULL, lease_until = NULL, updated_at = now()
WHERE (status <> 'excluded' OR active_hash IS NOT NULL)
AND NOT EXISTS (SELECT 1 FROM moment m WHERE m.id = s.moment_id AND ` + ragEligible + `)`).Error; err != nil {
			return err
		}
		if err := tx.Exec(`UPDATE rag_index_state SET status = 'failed', last_error = 'lease_expired',
lease_token = NULL, lease_until = NULL, updated_at = now()
WHERE status = 'running' AND lease_until < now() AND attempts >= 5`).Error; err != nil {
			return err
		}
		return tx.Exec(`DELETE FROM rag_chunk c USING rag_index_state s
WHERE c.moment_id = s.moment_id AND s.status = 'excluded'`).Error
	})
}

// Each worker iteration atomically leases one document, not a list followed by
// per-row lookup. Chunk writes and source validation below are bulk operations.
func (r *RAGRepository) Claim(ctx context.Context, profile, leaseToken string) (*domain.Source, error) {
	var rows []domain.Source
	err := r.db.WithContext(ctx).Raw(`WITH target AS (
    SELECT moment_id FROM rag_index_state WHERE desired_profile = ? AND attempts < 5
    AND ((status IN ('pending', 'failed') AND next_attempt_at <= now())
      OR (status = 'running' AND lease_until < now()))
    ORDER BY next_attempt_at, moment_id LIMIT 1 FOR UPDATE SKIP LOCKED
), claimed AS (
    UPDATE rag_index_state s SET status = 'running', attempts = attempts + 1,
        lease_token = ?, lease_until = now() + interval '10 minutes', updated_at = now()
    FROM target WHERE s.moment_id = target.moment_id RETURNING s.*
)
SELECT m.id AS moment_id, m.title, m.summary, m.content, s.source_hash, s.revision, s.attempts, s.lease_token
FROM claimed s JOIN moment m ON m.id = s.moment_id
WHERE `+ragEligible+` AND s.source_hash = `+ragCurrentHash, profile, leaseToken).Scan(&rows).Error
	if err != nil || len(rows) == 0 {
		return nil, err
	}
	return &rows[0], nil
}

func (r *RAGRepository) CurrentSource(ctx context.Context, source domain.Source, profile string) (bool, error) {
	var result struct{ Current bool }
	err := r.db.WithContext(ctx).Raw(`SELECT EXISTS (SELECT 1 FROM moment m
JOIN rag_index_state s ON s.moment_id = m.id
WHERE m.id = ? AND `+ragEligible+` AND `+ragCurrentHash+` = ?
AND s.revision = ? AND s.lease_token = ? AND s.desired_profile = ? AND s.status = 'running') AS current`,
		source.MomentID, source.SourceHash, source.Revision, source.LeaseToken, profile).Scan(&result).Error
	return result.Current, err
}

func (r *RAGRepository) Complete(ctx context.Context, source domain.Source, profile string, chunks []domain.Chunk, durationMs int64) error {
	payload, err := json.Marshal(chunks)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Lock source before index state, matching source-write/trigger lock order.
		var live []struct{ SourceHash string }
		if err := tx.Raw(`SELECT `+ragCurrentHash+` AS source_hash FROM moment m
WHERE m.id = ? AND `+ragEligible+` FOR SHARE OF m`, source.MomentID).Scan(&live).Error; err != nil {
			return err
		}
		if len(live) == 0 || live[0].SourceHash != source.SourceHash {
			return domain.ErrStaleSource
		}
		var states []struct{ Revision int64 }
		if err := tx.Raw(`SELECT revision FROM rag_index_state WHERE moment_id = ?
AND revision = ? AND lease_token = ? AND desired_profile = ? AND status = 'running' FOR UPDATE`,
			source.MomentID, source.Revision, source.LeaseToken, profile).Scan(&states).Error; err != nil {
			return err
		}
		if len(states) == 0 {
			return domain.ErrStaleSource
		}
		if err := tx.Exec(`DELETE FROM rag_chunk WHERE moment_id = ?`, source.MomentID).Error; err != nil {
			return err
		}
		if err := tx.Exec(`INSERT INTO rag_chunk
(moment_id, source_hash, profile, seq, content, context_header, kind, start_at, end_at, embedding)
SELECT ?, ?, ?, x.seq, x.content, x."contextHeader", x.kind, x.start, x."end", x.vector
FROM jsonb_to_recordset(?::jsonb) AS x(seq INTEGER, content TEXT, "contextHeader" TEXT,
kind TEXT, start INTEGER, "end" INTEGER, vector DOUBLE PRECISION[])`,
			source.MomentID, source.SourceHash, profile, string(payload)).Error; err != nil {
			return err
		}
		return tx.Exec(`UPDATE rag_index_state SET active_hash = source_hash, active_profile = ?, status = 'ready',
lease_token = NULL, lease_until = NULL, last_error = NULL, indexed_at = now(), index_duration_ms = ?, updated_at = now() WHERE moment_id = ?`,
			profile, durationMs, source.MomentID).Error
	})
}

func (r *RAGRepository) Fail(ctx context.Context, source domain.Source, reason string) error {
	return r.db.WithContext(ctx).Exec(`UPDATE rag_index_state SET status = 'failed', last_error = ?,
lease_token = NULL, lease_until = NULL,
next_attempt_at = now() + LEAST(300, power(2, attempts) * 5) * interval '1 second', updated_at = now()
WHERE moment_id = ? AND revision = ? AND lease_token = ?`,
		reason, source.MomentID, source.Revision, source.LeaseToken).Error
}

const ragLiveIndex = `JOIN rag_index_state s ON s.moment_id = c.moment_id
JOIN moment m ON m.id = c.moment_id
WHERE ` + ragEligible + ` AND c.profile = ? AND s.active_profile = c.profile
AND c.source_hash = s.active_hash AND c.source_hash = ` + ragCurrentHash

func (r *RAGRepository) Stats(ctx context.Context, profile string) (domain.IndexStats, error) {
	var stats domain.IndexStats
	err := r.db.WithContext(ctx).Raw(ragAdminBase+`
SELECT count(*) FILTER (WHERE status = 'unindexed') AS unindexed,
count(*) FILTER (WHERE status = 'outdated') AS outdated,
count(*) FILTER (WHERE status = 'pending') AS pending,
count(*) FILTER (WHERE status = 'running') AS running,
count(*) FILTER (WHERE status = 'ready' AND chunks > 0) AS ready,
count(*) FILTER (WHERE status = 'failed') AS failed,
count(*) FILTER (WHERE status = 'excluded') AS excluded,
COALESCE(sum(chunks), 0) AS chunks,
COALESCE(CASE WHEN min(min_dimension) = max(max_dimension) THEN min(min_dimension) ELSE 0 END, 0)
AS embedding_dimension FROM base`, profile, profile).Scan(&stats).Error
	return stats, err
}

func (r *RAGRepository) Retrieve(ctx context.Context, profile, question, contentKind string, vector []float64, tuning domain.Tuning) ([]domain.Evidence, []domain.Evidence, error) {
	queryVector, err := json.Marshal(vector)
	if err != nil {
		return nil, nil, err
	}
	terms, err := json.Marshal(lexicalTerms(question))
	if err != nil {
		return nil, nil, err
	}
	type row struct {
		domain.Evidence
		Channel string
	}
	var rows []row
	err = r.db.WithContext(ctx).Raw(`WITH q AS (
    SELECT ARRAY(SELECT value::double precision FROM jsonb_array_elements_text(?::jsonb)) AS embedding,
        ?::jsonb AS terms
), base AS MATERIALIZED (
    SELECT c.id, c.moment_id, m.title, m.short_url, c.content, c.context_header, c.kind,
        m.ext_info->>'contentKind' AS content_kind, c.start_at AS start, c.end_at AS "end",
        c.source_hash, c.profile AS index_version, m.created_at, m.updated_at, c.embedding
    FROM rag_chunk c `+ragLiveIndex+` AND (? = '' OR m.ext_info->>'contentKind' = ?)
), vector_scores AS (
    SELECT b.*, (SELECT sum(v.a * v.b) FROM unnest(b.embedding, q.embedding) AS v(a, b)) AS score
    FROM base b, q WHERE cardinality(b.embedding) = cardinality(q.embedding)
), vector_hits AS (
    SELECT *, 'vector'::text AS channel FROM vector_scores WHERE score >= ?
    ORDER BY score DESC, id LIMIT ?
), keyword_scores AS (
    SELECT b.*, (SELECT sum(CASE WHEN position(term IN lower(b.title || ' ' || b.context_header)) > 0
        THEN 3.0 ELSE 1.0 END)
        FROM jsonb_array_elements_text(q.terms) AS t(term)
        WHERE position(term IN lower(b.title || ' ' || b.context_header || ' ' || b.content)) > 0) AS score
    FROM base b, q
), keyword_hits AS (
    SELECT *, 'keyword'::text AS channel FROM keyword_scores WHERE score > 0
    ORDER BY score DESC, id LIMIT ?
)
SELECT id, moment_id, title, short_url, content, context_header, kind, content_kind,
start, "end", source_hash, index_version, created_at, updated_at, score, channel FROM vector_hits
UNION ALL
SELECT id, moment_id, title, short_url, content, context_header, kind, content_kind,
start, "end", source_hash, index_version, created_at, updated_at, score, channel FROM keyword_hits
ORDER BY channel, score DESC, id`,
		string(queryVector), string(terms), profile, contentKind, contentKind, tuning.MinSimilarity, tuning.VectorTopK, tuning.KeywordTopK).Scan(&rows).Error
	vectorResults, keywordResults := make([]domain.Evidence, 0), make([]domain.Evidence, 0)
	for _, row := range rows {
		if row.Channel == "vector" {
			vectorResults = append(vectorResults, row.Evidence)
		} else {
			keywordResults = append(keywordResults, row.Evidence)
		}
	}
	return vectorResults, keywordResults, err
}

func (r *RAGRepository) Validate(ctx context.Context, profile string, evidence []domain.Evidence) (bool, error) {
	if len(evidence) == 0 {
		return false, nil
	}
	type ref struct {
		ID        int64     `json:"id"`
		Hash      string    `json:"hash"`
		CreatedAt time.Time `json:"createdAt"`
		UpdatedAt time.Time `json:"updatedAt"`
	}
	refs := make([]ref, len(evidence))
	for i, item := range evidence {
		refs[i] = ref{item.ID, item.SourceHash, item.CreatedAt, item.UpdatedAt}
	}
	payload, err := json.Marshal(refs)
	if err != nil {
		return false, err
	}
	var result struct{ Count int64 }
	err = r.db.WithContext(ctx).Raw(`SELECT count(*) FROM rag_chunk c `+ragLiveIndex+`
AND EXISTS (SELECT 1 FROM jsonb_to_recordset(?::jsonb) AS ref(id BIGINT, hash TEXT,
    "createdAt" TIMESTAMPTZ, "updatedAt" TIMESTAMPTZ)
    WHERE ref.id = c.id AND ref.hash = c.source_hash
    AND ref."createdAt" = m.created_at AND ref."updatedAt" = m.updated_at)`, profile, string(payload)).Scan(&result).Error
	return result.Count == int64(len(evidence)), err
}

// Keep exact names and code terms; Chinese bigrams provide a substring channel
// without introducing a segmentation dependency. RRF combines only rank order.
func lexicalTerms(question string) []string {
	terms := make([]string, 0)
	seen := make(map[string]bool)
	add := func(term string) {
		if term != "" && !seen[term] && len(terms) < 48 {
			seen[term] = true
			terms = append(terms, term)
		}
	}
	var word []rune
	flush := func() { add(string(word)); word = nil }
	var previous rune
	for _, ch := range strings.ToLower(question) {
		if unicode.Is(unicode.Han, ch) {
			flush()
			if previous != 0 {
				add(string([]rune{previous, ch}))
			}
			previous = ch
		} else {
			previous = 0
			if unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' || ch == '#' || ch == '+' {
				word = append(word, ch)
			} else {
				flush()
			}
		}
	}
	flush()
	return terms
}
