package rag

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	domain "github.com/shawns-yao/shawn-blog/server/internal/domain/rag"
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
	return s.repo.DocumentChunks(ctx, settings.profile, momentID, page, pageSize)
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
