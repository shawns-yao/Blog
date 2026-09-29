package rag

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	domain "github.com/shawns-yao/shawn-blog/server/internal/domain/rag"
	infrarag "github.com/shawns-yao/shawn-blog/server/internal/infra/rag"
)

func (s *Service) Run(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	var reconciledAt time.Time
	lastProfile := ""
	for {
		settings, err := s.loadIndexSettings(ctx)
		if err == nil {
			if settings.profile != lastProfile || time.Since(reconciledAt) >= time.Minute {
				reconcileCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
				err = s.repo.Reconcile(reconcileCtx, settings.profile, false)
				cancel()
				if err == nil {
					lastProfile, reconciledAt = settings.profile, time.Now()
				}
			}
			if err == nil && settings.embedder.Ready() {
				s.indexNext(ctx, settings)
			}
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (s *Service) indexNext(ctx context.Context, settings settings) {
	// Leave one minute before the repository's ten-minute lease expires.
	workCtx, cancel := context.WithTimeout(ctx, 9*time.Minute)
	defer cancel()
	source, err := s.repo.Claim(workCtx, settings.profile, uuid.NewString())
	if err != nil {
		log.Print("[rag] index queue unavailable")
		return
	}
	if source == nil {
		return
	}
	started := time.Now()
	chunks := infrarag.SplitMarkdownWithTuning(source.Title, source.Content, settings.tuning)
	reason := ""
	if len(chunks) == 0 {
		reason = "empty_content"
	}
	for _, chunk := range chunks {
		limit := infrarag.ChunkTokenLimit(chunk.Kind, settings.tuning.ChunkMaxTokens)
		if chunk.Tokens > limit {
			log.Printf("[rag] oversized chunk moment_id=%d kind=%s tokens=%d limit=%d start=%d end=%d",
				source.MomentID, chunk.Kind, chunk.Tokens, limit, chunk.Start, chunk.End)
			reason = "oversized_atomic_block"
			break
		}
	}
	dimensions := 0
	// One live-source check per HTTP attempt stops further dispatch after withdrawal, including failover.
	// This is a source lifecycle check, not per-chunk SQL or a list-query N+1.
	for start := 0; reason == "" && start < len(chunks); start += 16 {
		end := min(start+16, len(chunks))
		texts := make([]string, end-start)
		for i := start; i < end; i++ {
			texts[i-start] = chunks[i].ContextHeader + "\n\n" + chunks[i].Content
		}
		vectors, provider, err := settings.embedder.BatchEmbedChecked(workCtx, texts, func(attempt context.Context) error {
			current, err := s.repo.CurrentSource(attempt, *source, settings.profile)
			if err != nil || !current {
				return domain.ErrStaleSource
			}
			return nil
		})
		if err != nil {
			for _, failure := range provider.Failures {
				log.Printf("[rag] embedding failed moment_id=%d provider=%s reason=%s batch=%d",
					source.MomentID, failure.Provider, failure.Reason, len(texts))
			}
			reason = "embedding_unavailable"
			if errors.Is(err, domain.ErrStaleSource) {
				reason = "source_changed"
			}
			break
		}
		for i, vector := range vectors {
			if dimensions == 0 {
				dimensions = len(vector)
			}
			if len(vector) != dimensions {
				reason = "embedding_dimension_changed"
				break
			}
			chunks[start+i].Vector = vector
		}
	}
	if reason == "" {
		current, err := s.loadIndexSettings(workCtx)
		if err != nil || current.profile != settings.profile {
			reason = "configuration_changed"
		} else if err := s.repo.Complete(workCtx, *source, settings.profile, chunks, time.Since(started).Milliseconds()); err != nil {
			if errors.Is(err, domain.ErrStaleSource) {
				reason = "source_changed"
			} else {
				reason = "index_write_failed"
			}
		}
	}
	if reason != "" {
		updateCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.repo.Fail(updateCtx, *source, reason); err != nil {
			log.Printf("[rag] index status update failed moment_id=%d", source.MomentID)
		}
		log.Printf("[rag] index failed moment_id=%d revision=%d reason=%s", source.MomentID, source.Revision, reason)
	}
}
