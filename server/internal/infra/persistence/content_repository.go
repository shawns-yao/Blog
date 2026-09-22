package persistence

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/shawns-yao/shawn-blog/server/internal/app/contentutil"
	"github.com/shawns-yao/shawn-blog/server/internal/domain/content"
	"github.com/shawns-yao/shawn-blog/server/internal/infra/persistence/model"
)

type ContentRepository struct {
	db *gorm.DB
}

func (r *ContentRepository) CreateColumn(ctx context.Context, column *content.MomentColumn) error {
	return NewMomentColumnRepository(r.db).Create(ctx, column)
}

func (r *ContentRepository) GetColumnByID(ctx context.Context, id int64) (*content.MomentColumn, error) {
	var rec model.MomentColumn
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&rec).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, content.ErrColumnNotFound
		}
		return nil, err
	}
	return mapColumnToDomain(rec), nil
}

func (r *ContentRepository) GetColumnByShortURL(ctx context.Context, shortURL string) (*content.MomentColumn, error) {
	var rec model.MomentColumn
	if err := r.db.WithContext(ctx).Where("short_url = ?", shortURL).First(&rec).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, content.ErrColumnNotFound
		}
		return nil, err
	}
	return mapColumnToDomain(rec), nil
}

func (r *ContentRepository) ListColumns(ctx context.Context) ([]*content.MomentColumn, error) {
	var records []model.MomentColumn
	if err := r.db.WithContext(ctx).Order("name ASC").Find(&records).Error; err != nil {
		return nil, err
	}
	result := make([]*content.MomentColumn, len(records))
	for i, rec := range records {
		result[i] = mapColumnToDomain(rec)
	}
	return result, nil
}

func (r *ContentRepository) UpdateColumn(ctx context.Context, column *content.MomentColumn) error {
	return NewMomentColumnRepository(r.db).Update(ctx, column)
}

func (r *ContentRepository) DeleteColumn(ctx context.Context, id int64) error {
	return NewMomentColumnRepository(r.db).Delete(ctx, id)
}

func (r *ContentRepository) CreateTag(ctx context.Context, tag *content.Tag) error {
	rec := model.Tag{Name: tag.Name}
	if err := r.db.WithContext(ctx).Create(&rec).Error; err != nil {
		return err
	}
	tag.ID = rec.ID
	tag.CreatedAt = rec.CreatedAt
	tag.UpdatedAt = rec.UpdatedAt
	return nil
}

func (r *ContentRepository) GetTagByID(ctx context.Context, id int64) (*content.Tag, error) {
	var rec model.Tag
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&rec).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, content.ErrTagNotFound
		}
		return nil, err
	}
	return mapTagToDomain(rec), nil
}

func (r *ContentRepository) GetTagByName(ctx context.Context, name string) (*content.Tag, error) {
	var rec model.Tag
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&rec).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, content.ErrTagNotFound
		}
		return nil, err
	}
	return mapTagToDomain(rec), nil
}

func (r *ContentRepository) ListTags(ctx context.Context) ([]*content.Tag, error) {
	var records []model.Tag
	if err := r.db.WithContext(ctx).Order("name ASC").Find(&records).Error; err != nil {
		return nil, err
	}
	result := make([]*content.Tag, len(records))
	for i, rec := range records {
		result[i] = mapTagToDomain(rec)
	}
	return result, nil
}

func (r *ContentRepository) UpdateTag(ctx context.Context, tag *content.Tag) error {
	result := r.db.WithContext(ctx).
		Model(&model.Tag{}).
		Where("id = ?", tag.ID).
		Update("name", tag.Name)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return content.ErrTagNotFound
	}
	return nil
}

func (r *ContentRepository) DeleteTag(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Tag{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return content.ErrTagNotFound
	}
	return nil
}

func (r *ContentRepository) TagIDsExist(ctx context.Context, ids []int64) (bool, error) {
	if len(ids) == 0 {
		return true, nil
	}
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&model.Tag{}).
		Where("id IN ?", ids).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count == int64(len(ids)), nil
}

func (r *ContentRepository) AddTopicsToMoment(ctx context.Context, momentID int64, tagIDs []int64) error {
	if len(tagIDs) == 0 {
		return nil
	}
	records := make([]model.MomentTopic, 0, len(tagIDs))
	for _, tagID := range tagIDs {
		records = append(records, model.MomentTopic{
			MomentID: momentID,
			TagID:    tagID,
		})
	}
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "moment_id"}, {Name: "tag_id"}},
			DoNothing: true,
		}).
		Create(&records).Error
}

func (r *ContentRepository) SyncTopicsToMoment(ctx context.Context, momentID int64, tagIDs []int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("moment_id = ?", momentID).Delete(&model.MomentTopic{}).Error; err != nil {
			return err
		}
		if len(tagIDs) == 0 {
			return nil
		}
		records := make([]model.MomentTopic, 0, len(tagIDs))
		for _, tagID := range tagIDs {
			records = append(records, model.MomentTopic{
				MomentID: momentID,
				TagID:    tagID,
			})
		}
		return tx.Create(&records).Error
	})
}

func (r *ContentRepository) GetTopicsByMomentID(ctx context.Context, momentID int64) ([]*content.Tag, error) {
	var records []model.Tag
	err := r.db.WithContext(ctx).
		Model(&model.Tag{}).
		Joins("JOIN moment_topic ON moment_topic.tag_id = tag.id").
		Where("moment_topic.moment_id = ?", momentID).
		Order("tag.name ASC").
		Find(&records).Error
	if err != nil {
		return nil, err
	}
	result := make([]*content.Tag, len(records))
	for i, rec := range records {
		result[i] = mapTagToDomain(rec)
	}
	return result, nil
}

func (r *ContentRepository) UpdateMomentViews(ctx context.Context, momentID int64) error {
	result := r.db.WithContext(ctx).
		Model(&model.MomentMetrics{}).
		Where("moment_id = ?", momentID).
		UpdateColumn("views", gorm.Expr("views + ?", 1))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected > 0 {
		return nil
	}
	rec := model.MomentMetrics{
		MomentID: momentID,
		Views:    1,
	}
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "moment_id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"views":      gorm.Expr("moment_metrics.views + 1"),
				"updated_at": time.Now(),
			}),
		}).
		Create(&rec).Error
}

func (r *ContentRepository) GetMomentMetrics(ctx context.Context, momentID int64) (*content.MomentMetrics, error) {
	var rec model.MomentMetrics
	result := r.db.WithContext(ctx).Where("moment_id = ?", momentID).Limit(1).Find(&rec)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &content.MomentMetrics{
		MomentID:  rec.MomentID,
		Views:     rec.Views,
		Likes:     rec.Likes,
		Comments:  rec.Comments,
		UpdatedAt: rec.UpdatedAt,
	}, nil
}

func NewContentRepository(db *gorm.DB) *ContentRepository {
	return &ContentRepository{db: db}
}

func createCommentArea(tx *gorm.DB, areaType, displayType, title string, contentID int64) (int64, error) {
	area := model.CommentArea{
		AreaName:  contentutil.BuildCommentAreaName(displayType, title),
		AreaType:  areaType,
		ContentID: &contentID,
		IsClosed:  false,
	}
	if err := tx.Create(&area).Error; err != nil {
		return 0, err
	}
	return area.ID, nil
}

func deleteCommentArea(tx *gorm.DB, areaID int64) error {
	if err := tx.Where("area_id = ?", areaID).Delete(&model.Comment{}).Error; err != nil {
		return err
	}
	return tx.Delete(&model.CommentArea{}, areaID).Error
}

// CreateMoment 创建手记
func (r *ContentRepository) CreateMoment(ctx context.Context, moment *content.Moment) error {
	tocBytes, err := tocToBytes(moment.TOC)
	if err != nil {
		return err
	}

	momentModel := &model.Moment{
		Title:            moment.Title,
		Summary:          moment.Summary,
		AISummary:        moment.AISummary,
		TOC:              tocBytes,
		Content:          moment.Content,
		ContentHash:      moment.ContentHash,
		AuthorID:         moment.AuthorID,
		Cover:            moment.Cover,
		ColumnID:         moment.ColumnID,
		ShortURL:         moment.ShortURL,
		IsPublished:      moment.IsPublished,
		IsTop:            moment.IsTop,
		IsHot:            moment.IsHot,
		IsOriginal:       moment.IsOriginal,
		ExtInfo:          moment.ExtInfo,
		ContentUpdatedAt: moment.ContentUpdatedAt,
		CreatedAt:        moment.CreatedAt,
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(momentModel).Error; err != nil {
			if isMomentShortURLConstraint(err) {
				return content.ErrMomentShortURLExists
			}
			return err
		}

		areaID, err := createCommentArea(tx, contentutil.CommentAreaTypeMoment, "手记", momentModel.Title, momentModel.ID)
		if err != nil {
			return err
		}
		if err := tx.Model(&model.Moment{}).
			Where("id = ?", momentModel.ID).
			Update("comment_id", areaID).Error; err != nil {
			return err
		}
		momentModel.CommentID = &areaID

		metrics := model.MomentMetrics{
			MomentID: momentModel.ID,
			Views:    moment.InitialViews,
			Likes:    0,
			Comments: 0,
		}
		if err := tx.Create(&metrics).Error; err != nil {
			return err
		}

		moment.ID = momentModel.ID
		moment.CommentID = momentModel.CommentID
		moment.UpdatedAt = momentModel.UpdatedAt
		return nil
	})
}

// GetMomentByID 根据ID获取手记
func (r *ContentRepository) GetMomentByID(ctx context.Context, id int64) (*content.Moment, error) {
	var momentModel model.Moment
	result := r.db.WithContext(ctx).Where("id = ?", id).Limit(1).Find(&momentModel)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, content.ErrMomentNotFound
	}

	return r.modelToMoment(&momentModel), nil
}

// GetMomentByShortURL 根据短链接获取手记
func (r *ContentRepository) GetMomentByShortURL(ctx context.Context, shortURL string) (*content.Moment, error) {
	var momentModel model.Moment
	result := r.db.WithContext(ctx).Where("short_url = ?", shortURL).Limit(1).Find(&momentModel)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, content.ErrMomentNotFound
	}

	return r.modelToMoment(&momentModel), nil
}

func (r *ContentRepository) GetMomentByActivityPubObjectID(ctx context.Context, objectID string) (*content.Moment, error) {
	objectID = strings.TrimSpace(objectID)
	if objectID == "" {
		return nil, content.ErrMomentNotFound
	}
	var momentModel model.Moment
	result := r.db.WithContext(ctx).Where("activitypub_object_id = ?", objectID).Limit(1).Find(&momentModel)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, content.ErrMomentNotFound
	}
	return r.modelToMoment(&momentModel), nil
}

// UpdateMoment 更新手记
func (r *ContentRepository) UpdateMoment(ctx context.Context, moment *content.Moment) error {
	tocBytes, err := tocToBytes(moment.TOC)
	if err != nil {
		return err
	}

	now := time.Now()
	updates := map[string]any{
		"title":                         moment.Title,
		"summary":                       moment.Summary,
		"ai_summary":                    moment.AISummary,
		"toc":                           tocBytes,
		"content":                       moment.Content,
		"content_hash":                  moment.ContentHash,
		"column_id":                     moment.ColumnID,
		"cover":                         moment.Cover,
		"short_url":                     moment.ShortURL,
		"activitypub_object_id":         moment.ActivityPubObjectID,
		"activitypub_last_published_at": moment.ActivityPubLastPublishedAt,
		"is_published":                  moment.IsPublished,
		"is_top":                        moment.IsTop,
		"is_hot":                        moment.IsHot,
		"is_original":                   moment.IsOriginal,
		"ext_info":                      moment.ExtInfo,
		"content_updated_at":            moment.ContentUpdatedAt,
		"updated_at":                    now,
	}
	if err := r.db.WithContext(ctx).
		Model(&model.Moment{}).
		Where("id = ?", moment.ID).
		Updates(updates).Error; err != nil {
		if isMomentShortURLConstraint(err) {
			return content.ErrMomentShortURLExists
		}
		return err
	}
	if moment.CommentID != nil {
		_ = r.db.WithContext(ctx).
			Model(&model.CommentArea{}).
			Where("id = ?", *moment.CommentID).
			Update("area_name", contentutil.BuildCommentAreaName("手记", moment.Title)).Error
	}

	moment.UpdatedAt = now
	return nil
}

// DeleteMoment 删除手记（软删除）
func (r *ContentRepository) DeleteMoment(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var rec model.Moment
		if err := tx.Select("id", "comment_id").Where("id = ?", id).First(&rec).Error; err != nil {
			return err
		}
		if rec.CommentID != nil {
			if err := deleteCommentArea(tx, *rec.CommentID); err != nil {
				return err
			}
		}
		if err := tx.Where("id = ?", id).Delete(&model.Moment{}).Error; err != nil {
			return err
		}
		return tx.Where("moment_id = ?", id).Delete(&model.MomentMetrics{}).Error
	})
}

// SyncHotMoments 根据指标同步热门手记状态
func (r *ContentRepository) SyncHotMoments(ctx context.Context, vT, lT, cT int64) ([]content.HotMomentMarked, error) {
	result := make([]content.HotMomentMarked, 0)
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		hotMetrics := tx.Model(&model.MomentMetrics{}).
			Select("moment_id").
			Where("views >= ? OR likes >= ? OR comments >= ?", vT, lT, cT)

		var promoteRows []struct {
			ID          int64  `gorm:"column:id"`
			Title       string `gorm:"column:title"`
			ShortURL    string `gorm:"column:short_url"`
			IsPublished bool   `gorm:"column:is_published"`
		}
		if err := tx.Model(&model.Moment{}).
			Select("id", "title", "short_url", "is_published").
			Where("id IN (?)", hotMetrics).
			Where("is_hot = ?", false).
			Find(&promoteRows).Error; err != nil {
			return err
		}

		if len(promoteRows) > 0 {
			ids := make([]int64, 0, len(promoteRows))
			for _, row := range promoteRows {
				ids = append(ids, row.ID)
				result = append(result, content.HotMomentMarked{
					ID:          row.ID,
					Title:       row.Title,
					ShortURL:    row.ShortURL,
					IsPublished: row.IsPublished,
				})
			}
			if err := tx.Model(&model.Moment{}).
				Where("id IN ?", ids).
				Updates(map[string]any{"is_hot": true, "updated_at": time.Now()}).Error; err != nil {
				return err
			}
		}

		// 将不再满足阈值且当前是热门的手记取消热门状态
		if err := tx.Model(&model.Moment{}).
			Where("id NOT IN (?)", hotMetrics).
			Where("is_hot = ?", true).
			Updates(map[string]any{"is_hot": false, "updated_at": time.Now()}).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ListMoments 获取手记列表（内部使用，包含未发布）
func (r *ContentRepository) ListMoments(ctx context.Context, options content.MomentListOptionsInternal) ([]*content.Moment, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.Moment{})
	query = filterContentKind(query, options.ContentKind)

	if options.ColumnID != nil {
		query = query.Where("column_id = ?", *options.ColumnID)
	}
	if options.AuthorID != nil {
		query = query.Where("author_id = ?", *options.AuthorID)
	}
	if options.Published != nil {
		query = query.Where("is_published = ?", *options.Published)
	}
	if options.TopicID != nil {
		subQuery := r.db.WithContext(ctx).
			Model(&model.MomentTopic{}).
			Select("moment_id").
			Where("tag_id = ?", *options.TopicID)
		query = query.Where("id IN (?)", subQuery)
	}
	if options.Search != nil && *options.Search != "" {
		search := "%" + *options.Search + "%"
		query = query.Where("title ILIKE ? OR summary ILIKE ? OR content ILIKE ?", search, search, search)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (options.Page - 1) * options.PageSize
	var momentModels []*model.Moment
	if err := query.Order("created_at DESC").
		Offset(offset).
		Limit(options.PageSize).
		Find(&momentModels).Error; err != nil {
		return nil, 0, err
	}

	moments := make([]*content.Moment, len(momentModels))
	for i, mm := range momentModels {
		moments[i] = r.modelToMoment(mm)
	}

	return moments, total, nil
}

// ListPublicMoments 获取公开手记列表
func (r *ContentRepository) ListPublicMoments(ctx context.Context, options content.MomentListOptions) ([]*content.Moment, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.Moment{}).Where("is_published = ?", true)
	query = filterContentKind(query, options.ContentKind)

	if options.ColumnID != nil {
		if options.IncludeChildren {
			children := r.db.WithContext(ctx).Model(&model.MomentColumn{}).Select("id").Where("parent_id = ?", *options.ColumnID)
			query = query.Where("(column_id = ? OR column_id IN (?))", *options.ColumnID, children)
		} else {
			query = query.Where("column_id = ?", *options.ColumnID)
		}
	}
	if options.AuthorID != nil {
		query = query.Where("author_id = ?", *options.AuthorID)
	}
	if options.TopicID != nil {
		subQuery := r.db.WithContext(ctx).
			Model(&model.MomentTopic{}).
			Select("moment_id").
			Where("tag_id = ?", *options.TopicID)
		query = query.Where("id IN (?)", subQuery)
	}
	if options.Search != nil && *options.Search != "" {
		search := "%" + *options.Search + "%"
		query = query.Where("title ILIKE ? OR summary ILIKE ?", search, search)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (options.Page - 1) * options.PageSize
	var momentModels []*model.Moment
	order := "is_top DESC, created_at DESC"
	if options.NewestFirst {
		order = "created_at DESC, id DESC"
	}
	if err := query.Order(order).
		Offset(offset).
		Limit(options.PageSize).
		Find(&momentModels).Error; err != nil {
		return nil, 0, err
	}

	moments := make([]*content.Moment, len(momentModels))
	for i, mm := range momentModels {
		moments[i] = r.modelToMoment(mm)
	}

	return moments, total, nil
}

func filterContentKind(query *gorm.DB, kind string) *gorm.DB {
	switch kind {
	case content.KindNote, content.KindArticle:
		return query.Where("ext_info ->> 'contentKind' = ?", kind)
	case content.KindUnclassified:
		return query.Where("COALESCE(ext_info ->> 'contentKind', '') NOT IN ?", []string{content.KindNote, content.KindArticle})
	default:
		return query
	}
}

func (r *ContentRepository) ListPublishedMomentsByCreatedAtRange(ctx context.Context, start time.Time, end time.Time, limit int) ([]*content.Moment, error) {
	if limit <= 0 {
		limit = 2
	}
	if start.After(end) {
		start, end = end, start
	}

	var momentModels []*model.Moment
	if err := r.db.WithContext(ctx).
		Model(&model.Moment{}).
		Where("is_published = ?", true).
		Where("created_at >= ? AND created_at <= ?", start, end).
		Order("is_top DESC, created_at DESC").
		Limit(limit).
		Find(&momentModels).Error; err != nil {
		return nil, err
	}

	moments := make([]*content.Moment, len(momentModels))
	for i, mm := range momentModels {
		moments[i] = r.modelToMoment(mm)
	}
	return moments, nil
}

// ListPublicMomentsForFederation 获取用于联合时间线的公开手记列表。
func (r *ContentRepository) ListPublicMomentsForFederation(ctx context.Context, since *time.Time, until *time.Time, page int, pageSize int) ([]*content.Moment, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.Moment{}).Where("is_published = ?", true)

	if since != nil {
		query = query.Where("created_at >= ?", *since)
	}
	if until != nil {
		query = query.Where("created_at <= ?", *until)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var momentModels []*model.Moment
	if err := query.
		Select("id, short_url, title, summary, cover, author_id, created_at, updated_at").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&momentModels).Error; err != nil {
		return nil, 0, err
	}

	moments := make([]*content.Moment, len(momentModels))
	for i, mm := range momentModels {
		moments[i] = r.modelToMoment(mm)
	}

	return moments, total, nil
}

// modelToMoment 将数据库模型转换为领域对象
func (r *ContentRepository) modelToMoment(mm *model.Moment) *content.Moment {
	toc, err := bytesToToc(mm.TOC)
	if err != nil {
		toc = []content.TOCNode{}
	}

	return &content.Moment{
		ID:                         mm.ID,
		Title:                      mm.Title,
		Summary:                    mm.Summary,
		AISummary:                  mm.AISummary,
		Content:                    mm.Content,
		ContentHash:                mm.ContentHash,
		AuthorID:                   mm.AuthorID,
		TOC:                        toc,
		Cover:                      mm.Cover,
		ColumnID:                   mm.ColumnID,
		CommentID:                  mm.CommentID,
		ShortURL:                   mm.ShortURL,
		ActivityPubObjectID:        mm.ActivityPubObjectID,
		ActivityPubLastPublishedAt: mm.ActivityPubLastPublishedAt,
		IsPublished:                mm.IsPublished,
		IsTop:                      mm.IsTop,
		IsHot:                      mm.IsHot,
		IsOriginal:                 mm.IsOriginal,
		ExtInfo:                    mm.ExtInfo,
		ContentUpdatedAt:           mm.ContentUpdatedAt,
		CreatedAt:                  mm.CreatedAt,
		UpdatedAt:                  mm.UpdatedAt,
		DeletedAt:                  timeToTimePtr(mm.DeletedAt.Time),
	}
}

// 辅助函数
func timeToTimePtr(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

// JSONB 转换辅助函数
func tocToBytes(toc []content.TOCNode) ([]byte, error) {
	if toc == nil {
		return []byte("[]"), nil
	}
	return json.Marshal(toc)
}

func bytesToToc(data []byte) ([]content.TOCNode, error) {
	var toc []content.TOCNode
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return []content.TOCNode{}, nil
	}
	if err := json.Unmarshal(trimmed, &toc); err == nil {
		return toc, nil
	} else if len(trimmed) > 0 && trimmed[0] == '{' {
		return []content.TOCNode{}, nil
	} else {
		return nil, err
	}
}

func mapTagToDomain(rec model.Tag) *content.Tag {
	return &content.Tag{
		ID:        rec.ID,
		Name:      rec.Name,
		CreatedAt: rec.CreatedAt,
		UpdatedAt: rec.UpdatedAt,
		DeletedAt: deletedAtToPtr(rec.DeletedAt),
	}
}

func deletedAtToPtr(deleted gorm.DeletedAt) *time.Time {
	if !deleted.Valid {
		return nil
	}
	return &deleted.Time
}

func optionalString(val *string) string {
	if val == nil {
		return ""
	}
	return *val
}

func stringToPtr(val string) *string {
	if val == "" {
		return nil
	}
	return &val
}

func isMomentShortURLConstraint(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "uq_moment_short_url")
}
