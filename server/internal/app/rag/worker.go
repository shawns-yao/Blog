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
			if err == nil {
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
	workCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
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
		if chunk.Tokens > settings.tuning.ChunkMaxTokens {
			reason = "oversized_atomic_block"
			break
		}
	}
	dimensions := 0
	// One live-source check per HTTP batch stops further dispatch after withdrawal.
	// This is a source lifecycle check, not per-chunk SQL or a list-query N+1.
	for start := 0; reason == "" && start < len(chunks); start += 16 {
		current, err := s.repo.CurrentSource(workCtx, *source, settings.profile)
		if err != nil || !current {
			reason = "source_changed"
			break
		}
		end := min(start+16, len(chunks))
		texts := make([]string, end-start)
		for i := start; i < end; i++ {
			texts[i-start] = chunks[i].ContextHeader + "\n\n" + chunks[i].Content
		}
		vectors, err := settings.embedder.BatchEmbed(workCtx, texts)
		if err != nil {
			reason = "embedding_unavailable"
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
