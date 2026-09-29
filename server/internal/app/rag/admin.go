package rag

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	domain "github.com/shawns-yao/shawn-blog/server/internal/domain/rag"
	infrarag "github.com/shawns-yao/shawn-blog/server/internal/infra/rag"
)

func (s *Service) Documents(ctx context.Context, filter domain.DocumentFilter) (domain.DocumentPage, error) {
	if filter.Page < 1 || filter.Page > 100000 || filter.PageSize < 1 || filter.PageSize > 100 ||
		utf8.RuneCountInString(filter.Search) > 100 ||
		(filter.ContentKind != "" && filter.ContentKind != "article" && filter.ContentKind != "note") {
		return domain.DocumentPage{}, fmt.Errorf("文档筛选参数无效。")
	}
	validStatus := filter.Status == ""
	for _, status := range []string{"unindexed", "outdated", "pending", "running", "ready", "failed", "excluded"} {
		validStatus = validStatus || filter.Status == status
	}
	if !validStatus {
		return domain.DocumentPage{}, fmt.Errorf("文档索引状态无效。")
	}
	settings, err := s.loadTuning(ctx)
	if err != nil {
		return domain.DocumentPage{}, err
	}
	filter.Search = strings.TrimSpace(filter.Search)
	return s.repo.Documents(ctx, settings.profile, filter)
}

func (s *Service) DocumentChunks(ctx context.Context, momentID int64, page, pageSize int) ([]domain.Chunk, int64, error) {
	if momentID <= 0 || page < 1 || page > 100000 || pageSize < 1 || pageSize > 20 {
		return nil, 0, fmt.Errorf("分块分页参数无效。")
	}
	settings, err := s.loadTuning(ctx)
	if err != nil {
		return nil, 0, err
	}
	chunks, total, err := s.repo.DocumentChunks(ctx, settings.profile, momentID, page, pageSize)
	if err != nil || len(chunks) == 0 {
		return chunks, total, err
	}
	sources, err := s.repo.ContextSources(ctx, settings.profile, []int64{momentID})
	if err != nil || len(sources) != 1 {
		return nil, 0, domain.ErrStaleSource
	}
	hierarchy := infrarag.SplitMarkdownWithTuning(sources[0].Title, sources[0].Content, settings.tuning)
	for i, chunk := range chunks {
		if chunk.Seq >= len(hierarchy) || hierarchy[chunk.Seq].Content != chunk.Content ||
			hierarchy[chunk.Seq].Start != chunk.Start || hierarchy[chunk.Seq].End != chunk.End {
			return nil, 0, domain.ErrStaleSource
		}
		chunks[i] = hierarchy[chunk.Seq]
	}
	infrarag.BindChunkIDs(chunks, momentID, settings.profile)
	return chunks, total, nil
}

func (s *Service) ReindexDocument(ctx context.Context, momentID int64) error {
	if momentID <= 0 {
		return fmt.Errorf("文档标识无效。")
	}
	settings, err := s.loadTuning(ctx)
	if err != nil {
		return err
	}
	queued, err := s.repo.ReindexDocument(ctx, settings.profile, momentID)
	if err != nil {
		return err
	}
	if !queued {
		return fmt.Errorf("仅可重建已发布的文章与手记。")
	}
	return nil
}

func (s *Service) QueryMetrics(ctx context.Context) (domain.QueryMetrics, error) {
	return s.repo.QueryMetrics(ctx, 7)
}
