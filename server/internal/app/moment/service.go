package moment

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"time"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/contentutil"
	appEvent "github.com/grtsinry43/grtblog-v2/server/internal/app/event"
	domaincomment "github.com/grtsinry43/grtblog-v2/server/internal/domain/comment"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/content"
)

type Service struct {
	repo        content.Repository
	commentRepo domaincomment.CommentRepository
	events      appEvent.Bus
}

func NewService(repo content.Repository, commentRepo domaincomment.CommentRepository, events appEvent.Bus) *Service {
	if events == nil {
		events = appEvent.NopBus{}
	}
	return &Service{repo: repo, commentRepo: commentRepo, events: events}
}

// CreateMoment 创建手记
func (s *Service) CreateMoment(ctx context.Context, authorID int64, cmd CreateMomentCmd) (*content.Moment, error) {
	shortURL := ""
	if cmd.ShortURL != nil {
		shortURL = strings.TrimSpace(*cmd.ShortURL)
	}
	if shortURL == "" {
		shortURL = contentutil.GenerateShortURLFromTitle(cmd.Title)
	}
	shortURL, err := s.ensureShortURLAvailable(ctx, shortURL)
	if err != nil {
		return nil, err
	}

	if cmd.ColumnID != nil {
		if _, err := s.repo.GetColumnByID(ctx, *cmd.ColumnID); err != nil {
			return nil, err
		}
	}
	if err := s.ensureTagsExist(ctx, cmd.TopicIDs); err != nil {
		return nil, err
	}

	createdAt := time.Now()
	if cmd.CreatedAt != nil {
		createdAt = *cmd.CreatedAt
	}

	toc := contentutil.GenerateTOC(cmd.Content)
	summary := contentutil.BuildSummary(cmd.Summary, cmd.Content)

	moment := &content.Moment{
		Title:            cmd.Title,
		Summary:          summary,
		AISummary:        cmd.AISummary,
		TOC:              toc,
		Content:          cmd.Content,
		ContentHash:      content.MomentContentHash(cmd.Title, summary, cmd.Content),
		AuthorID:         authorID,
		Cover:            cmd.Cover,
		ColumnID:         cmd.ColumnID,
		ShortURL:         shortURL,
		IsPublished:      cmd.IsPublished,
		IsTop:            cmd.IsTop,
		IsHot:            false,
		IsOriginal:       cmd.IsOriginal,
		ExtInfo:          mergeExtInfoKeepingFederation(nil, cmd.ExtInfo),
		ContentUpdatedAt: createdAt,
		CreatedAt:        createdAt,
	}
	if cmd.Views != nil && *cmd.Views > 0 {
		moment.InitialViews = *cmd.Views
	}

	if err := s.repo.CreateMoment(ctx, moment); err != nil {
		return nil, err
	}
	if err := s.applyCommentAreaStatus(ctx, moment.CommentID, cmd.AllowComment); err != nil {
		return nil, err
	}

	if len(cmd.TopicIDs) > 0 {
		if err := s.repo.SyncTopicsToMoment(ctx, moment.ID, cmd.TopicIDs); err != nil {
			return nil, err
		}
	}

	now := time.Now()
	_ = s.events.Publish(ctx, MomentCreated{
		ID:        moment.ID,
		AuthorID:  moment.AuthorID,
		Title:     moment.Title,
		ShortURL:  moment.ShortURL,
		Published: moment.IsPublished,
		At:        now,
	})
	if moment.IsPublished {
		prevExtInfo := append([]byte(nil), moment.ExtInfo...)
		_ = s.events.Publish(ctx, MomentPublished{
			ID:       moment.ID,
			AuthorID: moment.AuthorID,
			Title:    moment.Title,
			ShortURL: moment.ShortURL,
			At:       now,
		})
		publishFederationSignals(ctx, s.events, moment, cmd.Content)
		if !bytes.Equal(prevExtInfo, moment.ExtInfo) {
			if err := s.repo.UpdateMoment(ctx, moment); err != nil {
				return nil, err
			}
		}
	}

	return moment, nil
}

// UpdateMoment 更新手记
func (s *Service) UpdateMoment(ctx context.Context, cmd UpdateMomentCmd) (*content.Moment, error) {
	existing, err := s.repo.GetMomentByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}
	prevPublished := existing.IsPublished
	prevContentHash := existing.ContentHash

	if cmd.ColumnID != nil {
		if _, err := s.repo.GetColumnByID(ctx, *cmd.ColumnID); err != nil {
			return nil, err
		}
	}
	if err := s.ensureTagsExist(ctx, cmd.TopicIDs); err != nil {
		return nil, err
	}

	toc := contentutil.GenerateTOC(cmd.Content)
	summary := contentutil.BuildSummary(cmd.Summary, cmd.Content)

	existing.Title = cmd.Title
	existing.Summary = summary
	existing.AISummary = cmd.AISummary
	existing.TOC = toc
	existing.Content = cmd.Content
	existing.ContentHash = content.MomentContentHash(cmd.Title, summary, cmd.Content)
	if prevContentHash != existing.ContentHash {
		existing.ContentUpdatedAt = time.Now()
	}
	existing.Cover = cmd.Cover
	existing.ColumnID = cmd.ColumnID
	shortURL := strings.TrimSpace(cmd.ShortURL)
	if shortURL == "" {
		shortURL = existing.ShortURL
	}
	if shortURL != existing.ShortURL {
		shortURL, err = s.ensureShortURLAvailable(ctx, shortURL)
		if err != nil {
			return nil, err
		}
	}
	existing.ShortURL = shortURL
	existing.IsPublished = cmd.IsPublished
	existing.IsTop = cmd.IsTop
	existing.IsOriginal = cmd.IsOriginal
	existing.ExtInfo = mergeExtInfoKeepingFederation(existing.ExtInfo, cmd.ExtInfo)

	if err := s.repo.UpdateMoment(ctx, existing); err != nil {
		return nil, err
	}
	if err := s.applyCommentAreaStatus(ctx, existing.CommentID, cmd.AllowComment); err != nil {
		return nil, err
	}

	if err := s.repo.SyncTopicsToMoment(ctx, existing.ID, cmd.TopicIDs); err != nil {
		return nil, err
	}

	now := time.Now()
	_ = s.events.Publish(ctx, MomentUpdated{
		ID:          existing.ID,
		AuthorID:    existing.AuthorID,
		Title:       existing.Title,
		ShortURL:    existing.ShortURL,
		Published:   existing.IsPublished,
		ContentHash: existing.ContentHash,
		Summary:     existing.Summary,
		TOC:         existing.TOC,
		Content:     existing.Content,
		At:          now,
	})
	if !prevPublished && existing.IsPublished {
		_ = s.events.Publish(ctx, MomentPublished{
			ID:       existing.ID,
			AuthorID: existing.AuthorID,
			Title:    existing.Title,
			ShortURL: existing.ShortURL,
			At:       now,
		})
	}
	if prevPublished && !existing.IsPublished {
		_ = s.events.Publish(ctx, MomentUnpublished{
			ID:       existing.ID,
			AuthorID: existing.AuthorID,
			Title:    existing.Title,
			ShortURL: existing.ShortURL,
			At:       now,
		})
	}
	if existing.IsPublished && (!prevPublished || prevContentHash != existing.ContentHash) {
		prevExtInfo := append([]byte(nil), existing.ExtInfo...)
		publishFederationSignals(ctx, s.events, existing, existing.Content)
		if !bytes.Equal(prevExtInfo, existing.ExtInfo) {
			if err := s.repo.UpdateMoment(ctx, existing); err != nil {
				return nil, err
			}
		}
	}

	return existing, nil
}

// ResetFederationSignals 重置手记 ext_info 中记录的联合条目状态，并按需重新触发分发。
func (s *Service) ResetFederationSignals(ctx context.Context, cmd ResetFederationSignalsCmd) (*content.Moment, bool, error) {
	existing, err := s.repo.GetMomentByID(ctx, cmd.ID)
	if err != nil {
		return nil, false, err
	}

	resetAll := len(cmd.Mentions) == 0 && len(cmd.Citations) == 0
	if updated, changed := resetDeliveredSignals(existing.ExtInfo, cmd.Mentions, cmd.Citations, resetAll); changed {
		existing.ExtInfo = updated
		if err := s.repo.UpdateMoment(ctx, existing); err != nil {
			return nil, false, err
		}
	}

	retriggered := cmd.Retrigger && existing.IsPublished
	if retriggered {
		prevExtInfo := append([]byte(nil), existing.ExtInfo...)
		publishFederationSignals(ctx, s.events, existing, existing.Content)
		if !bytes.Equal(prevExtInfo, existing.ExtInfo) {
			if err := s.repo.UpdateMoment(ctx, existing); err != nil {
				return nil, false, err
			}
		}
	}

	return existing, retriggered, nil
}

// GetMomentByID 根据 ID 获取手记
func (s *Service) GetMomentByID(ctx context.Context, id int64) (*content.Moment, error) {
	return s.repo.GetMomentByID(ctx, id)
}

// GetMomentByShortURL 根据短链接获取手记
func (s *Service) GetMomentByShortURL(ctx context.Context, shortURL string) (*content.Moment, error) {
	moment, err := s.repo.GetMomentByShortURL(ctx, shortURL)
	if err != nil {
		return nil, err
	}

	return moment, nil
}

// ListMoments 获取手记列表
func (s *Service) ListMoments(ctx context.Context, options content.MomentListOptionsInternal) ([]*content.Moment, int64, error) {
	return s.repo.ListMoments(ctx, options)
}

// ListPublicMoments 获取公开手记列表
func (s *Service) ListPublicMoments(ctx context.Context, options content.MomentListOptions) ([]*content.Moment, int64, error) {
	return s.repo.ListPublicMoments(ctx, options)
}

// BatchSetPublished 批量设置手记发布状态。
func (s *Service) BatchSetPublished(ctx context.Context, cmd BatchSetPublishedCmd) error {
	ids := normalizeIDs(cmd.IDs)
	for _, id := range ids {
		momentItem, err := s.repo.GetMomentByID(ctx, id)
		if err != nil {
			return err
		}
		prevPublished := momentItem.IsPublished
		if prevPublished == cmd.IsPublished {
			continue
		}
		momentItem.IsPublished = cmd.IsPublished
		if err := s.repo.UpdateMoment(ctx, momentItem); err != nil {
			return err
		}

		now := time.Now()
		_ = s.events.Publish(ctx, MomentUpdated{
			ID:          momentItem.ID,
			AuthorID:    momentItem.AuthorID,
			Title:       momentItem.Title,
			ShortURL:    momentItem.ShortURL,
			Published:   momentItem.IsPublished,
			ContentHash: momentItem.ContentHash,
			Summary:     momentItem.Summary,
			TOC:         momentItem.TOC,
			Content:     momentItem.Content,
			At:          now,
		})
		if cmd.IsPublished {
			_ = s.events.Publish(ctx, MomentPublished{
				ID:       momentItem.ID,
				AuthorID: momentItem.AuthorID,
				Title:    momentItem.Title,
				ShortURL: momentItem.ShortURL,
				At:       now,
			})
			// 草稿→发布时检测联合信号
			prevExtInfo := append([]byte(nil), momentItem.ExtInfo...)
			publishFederationSignals(ctx, s.events, momentItem, momentItem.Content)
			if !bytes.Equal(prevExtInfo, momentItem.ExtInfo) {
				if err := s.repo.UpdateMoment(ctx, momentItem); err != nil {
					return err
				}
			}
		} else {
			_ = s.events.Publish(ctx, MomentUnpublished{
				ID:       momentItem.ID,
				AuthorID: momentItem.AuthorID,
				Title:    momentItem.Title,
				ShortURL: momentItem.ShortURL,
				At:       now,
			})
		}
	}
	return nil
}

// BatchSetTop 批量设置手记置顶状态。
func (s *Service) BatchSetTop(ctx context.Context, cmd BatchSetTopCmd) error {
	ids := normalizeIDs(cmd.IDs)
	for _, id := range ids {
		momentItem, err := s.repo.GetMomentByID(ctx, id)
		if err != nil {
			return err
		}
		if momentItem.IsTop == cmd.IsTop {
			continue
		}
		momentItem.IsTop = cmd.IsTop
		if err := s.repo.UpdateMoment(ctx, momentItem); err != nil {
			return err
		}
		_ = s.events.Publish(ctx, MomentUpdated{
			ID:          momentItem.ID,
			AuthorID:    momentItem.AuthorID,
			Title:       momentItem.Title,
			ShortURL:    momentItem.ShortURL,
			Published:   momentItem.IsPublished,
			ContentHash: momentItem.ContentHash,
			Summary:     momentItem.Summary,
			TOC:         momentItem.TOC,
			Content:     momentItem.Content,
			At:          time.Now(),
		})
	}
	return nil
}

// DeleteMoment 删除手记
func (s *Service) DeleteMoment(ctx context.Context, id int64) error {
	momentItem, err := s.repo.GetMomentByID(ctx, id)
	if err != nil {
		return err
	}
	if err := s.repo.DeleteMoment(ctx, id); err != nil {
		return err
	}
	_ = s.events.Publish(ctx, MomentDeleted{
		ID:       momentItem.ID,
		AuthorID: momentItem.AuthorID,
		Title:    momentItem.Title,
		ShortURL: momentItem.ShortURL,
		At:       time.Now(),
	})
	return nil
}

// BatchDelete 批量删除手记。
func (s *Service) BatchDelete(ctx context.Context, cmd BatchDeleteCmd) error {
	ids := normalizeIDs(cmd.IDs)
	for _, id := range ids {
		if err := s.DeleteMoment(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

// GetMomentWithTopics 获取手记及其话题。
func (s *Service) GetMomentWithTopics(ctx context.Context, id int64) (*content.Moment, []*content.Tag, error) {
	momentItem, err := s.repo.GetMomentByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	tags, err := s.repo.GetTopicsByMomentID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	return momentItem, tags, nil
}

// GetMomentTopics 获取手记话题。
func (s *Service) GetMomentTopics(ctx context.Context, momentID int64) ([]*content.Tag, error) {
	return s.repo.GetTopicsByMomentID(ctx, momentID)
}

// GetMomentMetrics 获取手记指标
func (s *Service) GetMomentMetrics(ctx context.Context, momentID int64) (*content.MomentMetrics, error) {
	return s.repo.GetMomentMetrics(ctx, momentID)
}

// UpdateHotMoments 根据指标更新热门手记状态
func (s *Service) UpdateHotMoments(ctx context.Context, viewThreshold, likeThreshold, commentThreshold int64) error {
	marked, err := s.repo.SyncHotMoments(ctx, viewThreshold, likeThreshold, commentThreshold)
	if err != nil {
		return err
	}

	now := time.Now()
	for _, item := range marked {
		if !item.IsPublished || strings.TrimSpace(item.ShortURL) == "" {
			continue
		}
		_ = s.events.Publish(ctx, MomentHotMarked{
			ID:       item.ID,
			Title:    item.Title,
			ShortURL: item.ShortURL,
			At:       now,
		})
	}
	return nil
}

func (s *Service) ensureShortURLAvailable(ctx context.Context, shortURL string) (string, error) {
	shortURL = strings.TrimSpace(shortURL)
	if shortURL == "" {
		for i := 0; i < 5; i++ {
			candidate := contentutil.GenerateRandomShortURL()
			_, err := s.repo.GetMomentByShortURL(ctx, candidate)
			if err != nil {
				if errors.Is(err, content.ErrMomentNotFound) {
					return candidate, nil
				}
				return "", err
			}
		}
		return "", content.ErrMomentShortURLExists
	}

	existing, err := s.repo.GetMomentByShortURL(ctx, shortURL)
	if err != nil && !errors.Is(err, content.ErrMomentNotFound) {
		return "", err
	}
	if err == nil && existing != nil {
		return "", content.ErrMomentShortURLExists
	}
	return shortURL, nil
}

func (s *Service) ensureTagsExist(ctx context.Context, tagIDs []int64) error {
	if len(tagIDs) == 0 {
		return nil
	}
	unique := make(map[int64]struct{}, len(tagIDs))
	for _, id := range tagIDs {
		if id <= 0 {
			return content.ErrTagNotFound
		}
		unique[id] = struct{}{}
	}
	ids := make([]int64, 0, len(unique))
	for id := range unique {
		ids = append(ids, id)
	}
	ok, err := s.repo.TagIDsExist(ctx, ids)
	if err != nil {
		return err
	}
	if !ok {
		return content.ErrTagNotFound
	}
	return nil
}

func (s *Service) applyCommentAreaStatus(ctx context.Context, areaID *int64, allowComment *bool) error {
	if s.commentRepo == nil || areaID == nil || *areaID <= 0 || allowComment == nil {
		return nil
	}
	return s.commentRepo.SetAreaClosed(ctx, *areaID, !*allowComment)
}

func normalizeIDs(ids []int64) []int64 {
	if len(ids) == 0 {
		return nil
	}
	seen := make(map[int64]struct{}, len(ids))
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}
